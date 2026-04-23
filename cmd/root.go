package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hasdata-com/hasdata-cli/internal/client"
	"github.com/hasdata-com/hasdata-cli/internal/config"
	"github.com/spf13/cobra"
)

type userError struct{ err error }

func (e *userError) Error() string { return e.err.Error() }
func (e *userError) Unwrap() error { return e.err }

func userErr(msg string, args ...any) error {
	return &userError{err: fmt.Errorf(msg, args...)}
}

type networkError struct{ err error }

func (e *networkError) Error() string { return e.err.Error() }
func (e *networkError) Unwrap() error { return e.err }

type apiError struct {
	Status int
	Body   []byte
}

func (e *apiError) Error() string {
	return fmt.Sprintf("api returned HTTP %d: %s", e.Status, strings.TrimSpace(string(e.Body)))
}

// ExitCodeFor returns the process exit code for a given error.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var ue *userError
	if errors.As(err, &ue) {
		return 1
	}
	var ne *networkError
	if errors.As(err, &ne) {
		return 2
	}
	var ae *apiError
	if errors.As(err, &ae) {
		if ae.Status >= 500 {
			return 4
		}
		return 3
	}
	return 1
}

type globalOpts struct {
	APIKey     string
	ConfigPath string
	Endpoint   string
	Output     string
	Pretty     bool
	Raw        bool
	Verbose    bool
	Timeout    time.Duration
	Retries    int
}

var (
	opts       globalOpts
	loadedCfg  *config.Config
	cliClient  *client.Client
	clientOnce sync.Once
	clientErr  error

	versionStr = "dev"
	commitStr  = "none"
	dateStr    = "unknown"
)

func SetVersionInfo(v, c, d string) {
	versionStr, commitStr, dateStr = v, c, d
	rootCmd.Version = v
}

var rootCmd = &cobra.Command{
	Use:   "hasdata",
	Short: "hasdata — command-line interface for api.hasdata.com",
	Long: `hasdata is the official CLI for api.hasdata.com.

Every API exposed by https://api.hasdata.com/apis is available as a subcommand,
with flags auto-generated from the API schema. Run 'hasdata <api-slug> --help'
for per-API usage.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(c *cobra.Command, _ []string) error {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			return userErr("%v", err)
		}
		loadedCfg = cfg
		maybeNotifyUpdate(c.ErrOrStderr())
		return nil
	},
}

func Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
}

func RootCmd() *cobra.Command { return rootCmd }

func init() {
	f := rootCmd.PersistentFlags()
	f.StringVar(&opts.APIKey, "api-key", "", "API key (overrides HASDATA_API_KEY and config file)")
	f.StringVar(&opts.ConfigPath, "config", "", "path to config file (default ~/.hasdata/config.yaml)")
	f.StringVar(&opts.Endpoint, "endpoint", "", "override API endpoint (default "+config.DefaultEndpoint+")")
	f.StringVarP(&opts.Output, "output", "o", "", "write response body to file instead of stdout")
	f.BoolVar(&opts.Pretty, "pretty", false, "pretty-print JSON responses")
	f.BoolVar(&opts.Raw, "raw", false, "write raw response bytes without formatting")
	f.BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output (prints request URL and rate-limit headers to stderr)")
	f.DurationVar(&opts.Timeout, "timeout", 120*time.Second, "request timeout")
	f.IntVar(&opts.Retries, "retries", 2, "max retries on 429/5xx responses")
}

// Client returns the lazily-constructed HTTP client used by generated commands.
func Client() (*client.Client, error) {
	clientOnce.Do(func() {
		apiKey := config.ResolveAPIKey(opts.APIKey, loadedCfg)
		endpoint := config.ResolveEndpoint(opts.Endpoint, loadedCfg)
		cliClient = client.New(client.Options{
			APIKey:   apiKey,
			Endpoint: endpoint,
			Timeout:  opts.Timeout,
			Retries:  opts.Retries,
			Verbose:  opts.Verbose,
			Stderr:   os.Stderr,
		})
	})
	return cliClient, clientErr
}

// groupedCommands tracks categories in insertion order for help layout.
var (
	groupOrder []string
	groupSeen  = map[string]bool{}
)

// RegisterAPICommand is invoked by generated files to attach a command to the
// root, grouped by its API category.
func RegisterAPICommand(categoryID, categoryTitle string, cmd *cobra.Command) {
	if !groupSeen[categoryID] {
		rootCmd.AddGroup(&cobra.Group{ID: categoryID, Title: categoryTitle + ":"})
		groupSeen[categoryID] = true
		groupOrder = append(groupOrder, categoryID)
	}
	cmd.GroupID = categoryID
	rootCmd.AddCommand(cmd)
}

// Output formatting helpers used by generated commands.

func WriteResponse(cmd *cobra.Command, resp *client.Response) error {
	if resp.Status >= 400 {
		return &apiError{Status: resp.Status, Body: resp.Body}
	}
	data := resp.Body
	ct := strings.ToLower(resp.ContentType)
	isJSON := strings.Contains(ct, "json")
	if isJSON && !opts.Raw {
		if opts.Pretty || isTerminal(os.Stdout) {
			pretty, err := prettifyJSON(data)
			if err == nil {
				data = pretty
			}
		}
	}
	if opts.Output != "" {
		return os.WriteFile(opts.Output, data, 0o644)
	}
	_, err := cmd.OutOrStdout().Write(data)
	if err != nil {
		return err
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		_, _ = cmd.OutOrStdout().Write([]byte("\n"))
	}
	return nil
}

// Sort category IDs deterministically in help output.
func init() {
	rootCmd.SetHelpCommandGroupID("")
	cobra.EnableCommandSorting = true
	_ = sort.Strings
	_ = io.Discard
}
