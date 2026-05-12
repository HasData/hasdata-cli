# hasdata-cli

**The official command-line interface for [HasData](https://hasdata.com) — web scraping, SERP, and real-estate/e-commerce data APIs, wired for shell scripts, LLM agents, and RAG pipelines.**

[![CI](https://github.com/HasData/hasdata-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/HasData/hasdata-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/HasData/hasdata-cli?sort=semver)](https://github.com/HasData/hasdata-cli/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

One static binary. Every API at [`hasdata.com`](https://hasdata.com/apis) exposed as a subcommand. No SDK install, no dependencies, no glue code — `curl | sh`, export a key, pipe JSON to `jq` or straight into your LLM prompt.

```sh
curl -sSL https://raw.githubusercontent.com/HasData/hasdata-cli/main/install.sh | sh
export HASDATA_API_KEY=hd_xxx
hasdata google-serp --q "best espresso machine 2026"
```

---

## Why a CLI?

- **Agents & tool use** — drop `hasdata <api>` into LangChain, LlamaIndex, CrewAI, or your own agent loop as a shell tool. Stable JSON in, stable JSON out.
- **RAG ingestion** — stream fresh Google, Amazon, Zillow, and arbitrary web data into your vector store from a cron job or a `Makefile`, no backend required.
- **Prompt-time grounding** — `hasdata google-serp ... | jq .organic_results` ➜ into a system prompt to cut hallucinations on current events, product pricing, real-estate comps, reviews.
- **Dataset building** — parallel GNU-`xargs` invocations produce JSONL for LLM fine-tuning or evals.
- **Humans too** — one-off lookups from your terminal, full `--help` for every flag, tab-completion for every enum.

## Install

| Platform | Command |
|---|---|
| macOS / Linux | `curl -sSL https://raw.githubusercontent.com/HasData/hasdata-cli/main/install.sh \| sh` |
| Windows manual | download the `.zip` from [Releases](https://github.com/HasData/hasdata-cli/releases), extract, put `hasdata.exe` on `%PATH%` |
| From source   | `go install github.com/HasData/hasdata-cli@latest` |

The `install.sh` script detects your OS/arch, downloads the matching asset, and verifies its SHA-256 against the published `checksums.txt` before installing.

## Configure

```sh
export HASDATA_API_KEY=your_key        # preferred for CI / containers / agents
# or
hasdata configure                      # writes ~/.hasdata/config.yaml (0600)
```

Precedence: `--api-key` flag > `HASDATA_API_KEY` env > `~/.hasdata/config.yaml`. Get a key from the [HasData dashboard](https://hasdata.com).

## First calls

```sh
# Google SERP — structured organic / ads / knowledge graph / PAA
hasdata google-serp --q "langchain vs llamaindex" --gl us --pretty

# Render + scrape any URL (JS, proxies, markdown output, AI extraction)
hasdata web-scraping \
  --url "https://news.ycombinator.com" \
  --output-format markdown \
  --ai-extract-rules-json '{"top_story":{"type":"string","description":"headline of the top story"}}' \
  --pretty

# Amazon product lookup for price monitoring / comparison
hasdata amazon-product --asin B08N5WRWNW --pretty

# Zillow listings with complex filters
hasdata zillow-listing \
  --keyword "Austin, TX" --type forSale \
  --price-min 400000 --price-max 900000 \
  --beds-min 3 --home-types house --home-types townhome \
  --sort priceLowToHigh --pretty
```

Every command supports `--help`, `--pretty`, `--raw`, `--output file`, `--verbose`, `--timeout`, `--retries`, shell completion.

## Using it with LLMs

### Agent tool-call (Python + OpenAI-style tools)

```python
import subprocess, json

def hasdata(cmd: list[str]) -> dict:
    """Shell-tool wrapper around the hasdata CLI. Usable as an LLM tool."""
    out = subprocess.check_output(["hasdata", *cmd, "--raw"], text=True)
    return json.loads(out)

tool_spec = {
    "name": "web_search",
    "description": "Run a Google SERP query and return structured results.",
    "parameters": {
        "type": "object",
        "properties": {
            "query":   {"type": "string"},
            "country": {"type": "string", "default": "us"},
            "n":       {"type": "integer", "default": 10},
        },
        "required": ["query"],
    },
}

def web_search(query: str, country: str = "us", n: int = 10) -> dict:
    return hasdata(["google-serp", "--q", query, "--gl", country, "--num", str(n)])
```

Feed `tool_spec` to Claude / GPT / Gemini tool calling — zero Python dependencies on the HasData side.

### RAG ingestion (bash loop)

```sh
for q in "$@"; do
  hasdata google-serp --q "$q" --num 50 --raw \
    | jq -c '.organic_results[] | {url:.link, title, snippet}' \
    >> serp-corpus.jsonl
done
```

Point your embedder at `serp-corpus.jsonl`.

### Prompt-time grounding (no vector store)

```sh
CONTEXT=$(hasdata google-serp --q "latest gpu benchmarks" --num 5 --raw \
  | jq -r '.organic_results[] | "- \(.title): \(.snippet)"')
llm "Answer using this context only:\n$CONTEXT\n\nQuestion: what's the fastest consumer GPU right now?"
```

## Available APIs

`hasdata --help` lists all of them with per-call pricing. Grouped overview:

| Category | Commands |
|---|---|
| **Google SERP** | `google-serp` · `google-serp-light` · `google-ai-mode` · `google-news` · `google-shopping` · `google-immersive-product` · `google-events` · `google-short-videos` |
| **Google Maps** | `google-maps` · `google-maps-place` · `google-maps-reviews` · `google-maps-contributor-reviews` · `google-maps-photos` · `google-maps-posts` |
| **Google Other** | `google-images` · `google-trends` |
| **Search Engines** | `bing-serp` |
| **Web** | `web-scraping` (headless, AI extraction, markdown output, screenshots) |
| **E-commerce** | `amazon-product` · `amazon-search` · `amazon-seller` · `amazon-seller-products` · `shopify-products` · `shopify-collections` |
| **Real Estate** | `zillow-listing` · `zillow-property` · `redfin-listing` · `redfin-property` |
| **Travel** | `booking-search` · `booking-place` · `airbnb-listing` · `airbnb-property` · `google-flights` |
| **Business / Local** | `yelp-search` · `yelp-place` · `yellowpages-search` · `yellowpages-place` |
| **Jobs** | `indeed-listing` · `indeed-job` · `glassdoor-listing` · `glassdoor-job` |
| **Video** | `youtube-search-api` · `youtube-video-api` · `youtube-channel-api` · `youtube-transcript-api` |
| **Social** | `instagram-profile` |

## Flag patterns

- **Scalars / enums** — `--q text`, `--num 50`, `--block-ads=false`. Enum flags validate client-side and offer tab-completion.
- **Booleans defaulting to `true`** — paired negated form: `--no-block-ads`, `--no-screenshot`.
- **Lists** — repeat (`--lr lang_en --lr lang_fr`) or comma-join (`--lr lang_en,lang_fr`). Serialized as `key[]=value` for GET endpoints.
- **Anything ending in `-json`** — accepts raw JSON, `@path/to/file.json`, or `-` for stdin. Works for `--ai-extract-rules-json`, `--js-scenario-json`, `--extract-rules-json`, `--headers-json`, etc.
- **Key-value objects** — e.g. `--headers User-Agent=foo` (repeatable, splits on first `=`, values with `=` preserved). Combine with `--headers-json` for a JSON base; kv items override per key.

## Output & scripting

- JSON responses pretty-print when stdout is a TTY; raw when piped (great for `jq`). Force with `--pretty` / `--raw`.
- `--output file` writes raw response bytes (works for `screenshot` / image endpoints too).
- `--verbose` prints the outgoing URL and `X-RateLimit-*` headers on stderr.
- Exit codes: `0` success · `1` user error · `2` network · `3` API 4xx · `4` API 5xx. Script-safe.

## Shell completion

```sh
# zsh
hasdata completion zsh > "${fpath[1]}/_hasdata"
# bash
hasdata completion bash > /usr/local/etc/bash_completion.d/hasdata
# fish
hasdata completion fish > ~/.config/fish/completions/hasdata.fish
```

Enum values auto-complete (`hasdata google-serp --gl <TAB>` → `us`, `gb`, `ca`, …).

## Update

```sh
hasdata update           # upgrade to latest release
hasdata update --check   # report available version without installing
```

A once-per-24h check prints a one-line notice to stderr when a newer version is out. Disable with `check_updates: false` in `~/.hasdata/config.yaml`.

## How the CLI stays current

Every command here is generated from the live schema at `https://api.hasdata.com/apis`. A scheduled GitHub Action re-runs the generator, and a hash of the normalized spec short-circuits diffs when nothing changed. When HasData ships a new API, a PR lands here within 24 hours, then a release goes out — and `hasdata update` brings it to your machine.

Contributing locally:

```sh
go generate ./...     # regenerate cmd/gen_*.go from api.hasdata.com
go build ./...
go test ./...
```

## Resources

- **HasData docs** — <https://docs.hasdata.com>
- **API catalog** — <https://hasdata.com/apis>
- **Releases** — <https://github.com/HasData/hasdata-cli/releases>
- **Issues & feature requests** — <https://github.com/HasData/hasdata-cli/issues>

## License

[MIT](LICENSE) — use it commercially, embed it in your agent, ship it inside a container. Just don't hold us liable.
