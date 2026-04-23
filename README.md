# hasdata-cli

Command-line interface for [hasdata.com](https://hasdata.com).

## Install

**macOS / Linux** (curl | sh, verifies checksum):
```sh
curl -sSL https://raw.githubusercontent.com/hasdata-com/hasdata-cli/main/install.sh | sh
```

**Homebrew** (macOS / Linux):
```sh
brew install hasdata-com/tap/hasdata
```

**Scoop** (Windows):
```powershell
scoop bucket add hasdata https://github.com/hasdata-com/scoop-bucket
scoop install hasdata
```

**winget** (Windows):
```powershell
winget install hasdata-com.hasdata
```

**Manual**: download the archive matching your OS/arch from the [Releases](https://github.com/hasdata-com/hasdata-cli/releases) page, extract, place `hasdata` in your `PATH`.

## Configure

```sh
export HASDATA_API_KEY=your_key_here
# or
hasdata configure
```

Precedence: `--api-key` flag > `HASDATA_API_KEY` env var > `~/.hasdata/config.yaml`.

## Use

```sh
hasdata --help                    # all APIs grouped by category
hasdata google-serp --help        # flags for a specific API
hasdata google-serp --q "coffee" --gl us --num 20
hasdata web-scraping --url https://example.com --no-block-ads --extract-rules-json @rules.json
hasdata amazon-product --asin B08N5WRWNW --pretty
```

Output goes to stdout; use `--output file.json` to write to a file, `--pretty` to indent JSON, `--raw` to skip formatting entirely. `--verbose` prints the request URL and rate-limit headers to stderr.

### Flag value shapes

- **Scalars** (`--q text`, `--num 50`, `--block-ads=false`) — standard.
- **Enum flags** validate against allowed values; shell completion offers the list. See `hasdata <api> --help` for each flag's allowed values.
- **Boolean flags with default `true`** have a paired `--no-<flag>` form (e.g. `--no-block-ads`).
- **List flags** accept repeated values or a comma-separated form: `--lr lang_en --lr lang_fr` or `--lr lang_en,lang_fr`.
- **Any flag ending in `-json`** accepts raw JSON, a `@file` path, or `-` for stdin — e.g. `--ai-extract-rules-json @rules.json`, `--js-scenario-json '[{"wait":2000}]'`, `echo '{...}' | ... --extract-rules-json -`.
- **`additionalProperties` objects** (e.g. `headers`, `extractRules`) are exposed as **two flags**: `--headers k=v` (repeatable, splits on the first `=`) and `--headers-json '{...}'` as an escape hatch. When both are given, the JSON is the base and kv items override per key.

### Exit codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | user / CLI-input error |
| 2 | network error |
| 3 | API returned 4xx |
| 4 | API returned 5xx |

## Update

```sh
hasdata update              # upgrade to latest release
hasdata update --check      # report whether an update is available
```

A once-per-24h background check will print a one-line notice to stderr when a newer version is available. Disable it by writing `check_updates: false` into `~/.hasdata/config.yaml`.

## How it stays in sync

The CLI is regenerated from `/apis` and `/apis/<slug>` at build time (`internal/gen/main.go`). A scheduled GitHub Action (and a `repository_dispatch` trigger from the API side) re-runs the generator daily and opens a PR when the spec hash changes. New APIs reach users through a new release + `hasdata update`.

Contributors — if you're changing the CLI manually:
```sh
go generate ./...     # regenerate from live API specs
go build ./...
go test ./...
```

## License

MIT — see `LICENSE`.
