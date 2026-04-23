package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/HasData/hasdata-cli/internal/spec"
)

//go:embed templates/*.tmpl
var templates embed.FS

// complexArraySlugs are property names whose "type: array" schema has no items
// metadata but is actually an array of complex objects, so we expose them as
// JSON-only flags.
var complexArraySlugs = map[string]bool{
	"jsScenario": true,
}

const defaultEndpoint = "https://api.hasdata.com"

type flags struct {
	endpoint  string
	outDir    string
	specCache string
	strict    bool
	timeout   time.Duration
}

func main() {
	var f flags
	flag.StringVar(&f.endpoint, "endpoint", envOr("HASDATA_ENDPOINT", defaultEndpoint), "API endpoint to introspect")
	flag.StringVar(&f.outDir, "out", "cmd", "output directory for generated command files")
	flag.StringVar(&f.specCache, "hash-file", "internal/gen/spec-hash.txt", "path to file containing normalized-spec hash")
	flag.BoolVar(&f.strict, "strict", false, "fail if any property has an unsupported type (default: degrade + warn)")
	flag.DurationVar(&f.timeout, "timeout", 60*time.Second, "HTTP timeout for spec fetches")
	flag.Parse()

	if err := run(f); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func run(f flags) error {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout*4)
	defer cancel()

	httpClient := &http.Client{Timeout: f.timeout}

	catalog, err := fetchCatalog(ctx, httpClient, f.endpoint)
	if err != nil {
		return fmt.Errorf("fetch catalog: %w", err)
	}

	// Sort categories by title, APIs by slug for deterministic output.
	sort.SliceStable(catalog, func(i, j int) bool { return catalog[i].Title < catalog[j].Title })
	for i := range catalog {
		sort.SliceStable(catalog[i].APIs, func(a, b int) bool { return catalog[i].APIs[a].Slug < catalog[i].APIs[b].Slug })
	}

	details := make([]*spec.Detail, 0)
	for _, cat := range catalog {
		for _, api := range cat.APIs {
			d, err := fetchDetail(ctx, httpClient, f.endpoint, api.Slug)
			if err != nil {
				return fmt.Errorf("fetch %s: %w", api.Slug, err)
			}
			// Overlay category from catalog if the detail doesn't carry it.
			if d.Category.Slug == "" {
				d.Category = spec.Category{ID: cat.ID, Title: cat.Title, Slug: cat.Slug}
			}
			if d.Mode == "" && api.Mode != "" {
				d.Mode = api.Mode
			}
			details = append(details, d)
		}
	}

	// Deterministic hash of normalized specs (no timestamps, sorted keys).
	hash, err := normalizedHash(details)
	if err != nil {
		return err
	}
	prev := readPrevHash(f.specCache)
	if prev == hash && !f.strict {
		fmt.Printf("gen: spec hash unchanged (%s) — no files written\n", hash[:12])
		// Still idempotent: ensure hash file exists.
		_ = os.WriteFile(f.specCache, []byte(hash+"\n"), 0o644)
		return nil
	}

	// Load templates.
	cmdTmpl, err := template.ParseFS(templates, "templates/command.go.tmpl")
	if err != nil {
		return err
	}
	helpersTmpl, err := template.ParseFS(templates, "templates/helpers.go.tmpl")
	if err != nil {
		return err
	}

	// Remove all existing generated files so deletions propagate.
	if err := removeGenerated(f.outDir); err != nil {
		return err
	}

	// Write helpers.go.
	var hbuf bytes.Buffer
	if err := helpersTmpl.Execute(&hbuf, nil); err != nil {
		return err
	}
	if err := writeGoFile(filepath.Join(f.outDir, "gen_helpers.go"), hbuf.Bytes()); err != nil {
		return err
	}

	// Report collector.
	reportLines := []string{"# Generation Report", "", fmt.Sprintf("Generated at: %s", time.Now().UTC().Format(time.RFC3339)), ""}
	addedSlugs := []string{}
	degraded := []string{}
	deprecated := []string{}

	// Write one file per API.
	for _, d := range details {
		data, notes, err := buildTemplateData(d, f.strict)
		if err != nil {
			return fmt.Errorf("%s: %w", d.Slug, err)
		}
		addedSlugs = append(addedSlugs, d.Slug)
		if d.Mode == "deprecated" {
			deprecated = append(deprecated, d.Slug)
		}
		for _, n := range notes {
			degraded = append(degraded, fmt.Sprintf("- `%s.%s`: %s", d.Slug, n.field, n.reason))
		}
		var buf bytes.Buffer
		if err := cmdTmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("render %s: %w", d.Slug, err)
		}
		path := filepath.Join(f.outDir, "gen_"+strings.ReplaceAll(d.Slug, "-", "_")+".go")
		if err := writeGoFile(path, buf.Bytes()); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	reportLines = append(reportLines, fmt.Sprintf("## APIs generated (%d)", len(addedSlugs)))
	for _, s := range addedSlugs {
		reportLines = append(reportLines, "- "+s)
	}
	if len(deprecated) > 0 {
		reportLines = append(reportLines, "", "## Deprecated APIs (hidden in help)")
		for _, s := range deprecated {
			reportLines = append(reportLines, "- "+s)
		}
	}
	if len(degraded) > 0 {
		reportLines = append(reportLines, "", "## Degraded flags (unsupported schema → JSON fallback)")
		reportLines = append(reportLines, degraded...)
	}
	if err := os.WriteFile("GENERATION_REPORT.md", []byte(strings.Join(reportLines, "\n")+"\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(f.specCache, []byte(hash+"\n"), 0o644); err != nil {
		return err
	}
	fmt.Printf("gen: wrote %d commands (hash %s)\n", len(addedSlugs), hash[:12])
	return nil
}

// ------- fetch --------

func fetchCatalog(ctx context.Context, client *http.Client, endpoint string) ([]spec.CatalogEntry, error) {
	b, err := httpGet(ctx, client, endpoint+"/apis")
	if err != nil {
		return nil, err
	}
	var out []spec.CatalogEntry
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}
	return out, nil
}

func fetchDetail(ctx context.Context, client *http.Client, endpoint, slug string) (*spec.Detail, error) {
	b, err := httpGet(ctx, client, endpoint+"/apis/"+slug)
	if err != nil {
		return nil, err
	}
	var out spec.Detail
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", slug, err)
	}
	return &out, nil
}

func httpGet(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "hasdata-cli-codegen")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s: HTTP %d: %s", url, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return io.ReadAll(resp.Body)
}

// ------- hashing --------

func normalizedHash(details []*spec.Detail) (string, error) {
	// Build canonical JSON stripping created_at/updated_at (which we don't parse,
	// but still: marshal in a deterministic way).
	type norm struct {
		Slug   string               `json:"slug"`
		Title  string               `json:"title"`
		Cat    string               `json:"category"`
		Mode   string               `json:"mode"`
		Method string               `json:"method"`
		URL    string               `json:"url"`
		Cost   int                  `json:"cost"`
		Req    []string             `json:"required"`
		Props  map[string]spec.Property `json:"properties"`
	}
	list := make([]norm, 0, len(details))
	for _, d := range details {
		r := append([]string(nil), d.Schema.Required...)
		sort.Strings(r)
		list = append(list, norm{
			Slug: d.Slug, Title: d.Title, Cat: d.Category.Slug, Mode: d.Mode,
			Method: d.Schema.Method, URL: d.Schema.URL, Cost: d.Schema.Cost,
			Req: r, Props: d.Schema.Properties,
		})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Slug < list[j].Slug })
	b, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func readPrevHash(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ------- template data --------

type propData struct {
	Name          string // original JSON name
	Flag          string // CLI flag name (kebab-case)
	Var           string // Go variable name
	Title         string
	Help          string
	Required      bool
	Kind          string // string|int|float|bool|stringSlice|kvSlice|json
	DefaultLit    string // Go literal for default value
	EnumValuesStr string // `"a", "b", "c"`
	EnumValuesInt string // `1, 2, 3`
	BoolPairNeg   bool   // emit paired --no-<flag>
}

type cmdData struct {
	Use            string
	Ident          string
	CategoryID     string
	CategoryTitle  string
	Method         string
	URL            string
	Short          string
	Long           string
	Hidden         bool
	HasDeprecation bool
	DeprecationMsg string
	Props          []propData
}

type degradeNote struct {
	field  string
	reason string
}

func buildTemplateData(d *spec.Detail, strict bool) (cmdData, []degradeNote, error) {
	data := cmdData{
		Use:           d.Slug,
		Ident:         slugToIdent(d.Slug),
		CategoryID:    d.Category.Slug,
		CategoryTitle: d.Category.Title,
		Method:        strings.ToUpper(d.Schema.Method),
		URL:           d.Schema.URL,
		Short:         shortTitle(d),
		Long:          buildLong(d),
		Hidden:        d.Mode == "deprecated",
		HasDeprecation: d.Mode == "deprecated",
		DeprecationMsg: d.ModeMessage,
	}
	if data.Method == "" {
		data.Method = "GET"
	}

	requiredSet := map[string]bool{}
	for _, r := range d.Schema.Required {
		requiredSet[r] = true
	}

	// Sort property names for determinism.
	names := make([]string, 0, len(d.Schema.Properties))
	for n := range d.Schema.Properties {
		names = append(names, n)
	}
	sort.Strings(names)

	var notes []degradeNote

	for _, name := range names {
		p := d.Schema.Properties[name]
		pd, note, err := buildProp(name, p, requiredSet[name], strict)
		if err != nil {
			return data, nil, err
		}
		if note != nil {
			notes = append(notes, *note)
		}
		data.Props = append(data.Props, pd)
	}
	return data, notes, nil
}

func buildProp(name string, p spec.Property, required, strict bool) (propData, *degradeNote, error) {
	// Normalize the serialization key: trailing "[]" on the schema key is a
	// PHP/Laravel array marker, not part of the actual field name. Strip it
	// — the stringSlice template always re-appends "[]" at request time so
	// `lr` and `homeTypes[]` both end up as `name[]=value` on the wire.
	serName := strings.TrimSuffix(name, "[]")
	pd := propData{
		Name:     serName,
		Flag:     flagName(name),
		Var:      "p_" + strings.ReplaceAll(varName(name), "-", "_"),
		Title:    p.Title,
		Required: required,
	}
	pd.Help = buildHelp(p, required)

	switch p.Type {
	case "string":
		pd.Kind = "string"
		pd.DefaultLit = strconv.Quote(jsonStringDefault(p.Default, ""))
		if strs := enumStrings(p.Enum); len(strs) > 0 {
			pd.EnumValuesStr = quoteJoin(strs)
		}
	case "integer":
		pd.Kind = "int"
		def := 0
		if len(p.Default) > 0 {
			_ = json.Unmarshal(p.Default, &def)
		}
		pd.DefaultLit = strconv.Itoa(def)
		if ints := enumInts(p.Enum); len(ints) > 0 {
			pd.EnumValuesInt = intsJoin(ints)
		}
	case "number":
		pd.Kind = "float"
		def := 0.0
		if len(p.Default) > 0 {
			_ = json.Unmarshal(p.Default, &def)
		}
		pd.DefaultLit = strconv.FormatFloat(def, 'f', -1, 64)
	case "boolean":
		pd.Kind = "bool"
		def := false
		if len(p.Default) > 0 {
			_ = json.Unmarshal(p.Default, &def)
		}
		pd.DefaultLit = strconv.FormatBool(def)
		if def {
			// User is more likely to reach for --no-<flag>; emit a paired negated flag.
			pd.BoolPairNeg = true
		}
	case "array":
		switch {
		case complexArraySlugs[name]:
			// Curated complex-item arrays (e.g. jsScenario) → JSON-only.
			pd.Kind = "json"
			pd.Help = buildJSONHelp(p, required)
		case p.Items != nil && (p.Items.Type == "string" || p.Items.Type == "integer" || p.Items.Type == "number"):
			pd.Kind = "stringSlice"
			if strs := enumStrings(p.Items.Enum); len(strs) > 0 {
				pd.EnumValuesStr = quoteJoin(strs)
			}
		default:
			// items missing or non-scalar: default to string slice. If the API
			// turns out to want complex objects here, add the slug to
			// complexArraySlugs.
			pd.Kind = "stringSlice"
		}
	case "object":
		if len(p.AdditionalProperties) > 0 {
			// `additionalProperties` object (e.g. headers, extractRules) → kv + json.
			pd.Kind = "kvSlice"
		} else {
			pd.Kind = "json"
			pd.Help = buildJSONHelp(p, required)
		}
	default:
		msg := fmt.Sprintf("unsupported schema type %q — falling back to --%s-json", p.Type, pd.Flag)
		if strict {
			return pd, nil, errors.New(msg)
		}
		pd.Kind = "json"
		normalizeJSONFlag(&pd)
		return pd, &degradeNote{field: name, reason: msg}, nil
	}
	normalizeJSONFlag(&pd)
	return pd, nil, nil
}

// normalizeJSONFlag appends a "-json" suffix to pure-JSON flags, so every flag
// that accepts raw JSON / @file / stdin has a consistent shape regardless of
// whether it shares a name with a kv-form pair.
func normalizeJSONFlag(pd *propData) {
	if pd.Kind == "json" && !strings.HasSuffix(pd.Flag, "-json") {
		pd.Flag = pd.Flag + "-json"
	}
}

// ------- helpers --------

func enumStrings(raws []json.RawMessage) []string {
	out := make([]string, 0, len(raws))
	for _, r := range raws {
		var s string
		if err := json.Unmarshal(r, &s); err == nil {
			out = append(out, s)
		}
	}
	return out
}

func enumInts(raws []json.RawMessage) []int {
	out := make([]int, 0, len(raws))
	for _, r := range raws {
		var n int
		if err := json.Unmarshal(r, &n); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func quoteJoin(items []string) string {
	var b strings.Builder
	for i, s := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Quote(s))
	}
	return b.String()
}

func intsJoin(items []int) string {
	var b strings.Builder
	for i, n := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(n))
	}
	return b.String()
}

func jsonStringDefault(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return fallback
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return fallback
}

func shortTitle(d *spec.Detail) string {
	t := d.Title
	if t == "" {
		t = d.Schema.Title
	}
	if d.Mode == "beta" {
		t = "[beta] " + t
	}
	if d.Mode == "deprecated" {
		t = "[deprecated] " + t
	}
	t += fmt.Sprintf("  (%d credits/call)", d.Schema.Cost)
	return t
}

func buildLong(d *spec.Detail) string {
	var b strings.Builder
	desc := d.Description
	if desc == "" {
		desc = d.Schema.Description
	}
	b.WriteString(desc)
	if d.Mode == "deprecated" {
		b.WriteString("\n\nThis API is deprecated")
		if d.ModeMessage != "" {
			b.WriteString(" — ")
			b.WriteString(d.ModeMessage)
		}
		b.WriteString(".")
	} else if d.Mode == "beta" {
		b.WriteString("\n\nThis API is in beta.")
	}
	b.WriteString(fmt.Sprintf("\n\nEndpoint: %s %s", strings.ToUpper(d.Schema.Method), d.Schema.URL))
	b.WriteString(fmt.Sprintf("\nCost: %d credits per call.", d.Schema.Cost))
	return b.String()
}

func buildHelp(p spec.Property, required bool) string {
	parts := []string{}
	if p.Title != "" {
		parts = append(parts, p.Title+":")
	}
	if p.Description != "" {
		parts = append(parts, p.Description)
	}
	help := strings.Join(parts, " ")
	if len(p.Enum) > 0 {
		strs := enumStrings(p.Enum)
		if len(strs) > 0 {
			help += " [allowed: " + strings.Join(strs, "|") + "]"
		} else {
			ints := enumInts(p.Enum)
			if len(ints) > 0 {
				help += " [allowed: " + intsJoin(ints) + "]"
			}
		}
	}
	if required {
		help += " (required)"
	}
	return help
}

// buildJSONHelp returns the usage string for a pure-JSON flag, making the
// accepted input shapes explicit in --help.
func buildJSONHelp(p spec.Property, required bool) string {
	h := buildHelp(p, required)
	return h + " — accepts raw JSON, @file, or - (stdin)"
}

func removeGenerated(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.MkdirAll(outDir, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "gen_") && strings.HasSuffix(n, ".go") {
			if err := os.Remove(filepath.Join(outDir, n)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeGoFile(path string, data []byte) error {
	formatted, err := format.Source(data)
	if err != nil {
		// Write the unformatted source so the user can debug.
		_ = os.WriteFile(path, data, 0o644)
		return fmt.Errorf("gofmt %s: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0o644)
}
