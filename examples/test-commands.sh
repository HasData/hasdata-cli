#!/usr/bin/env bash
# hasdata-cli — manual smoke tests covering every flag-shape the generator can
# produce. Run commands individually (copy/paste) or step through the whole file:
#
#   export HASDATA_API_KEY=...
#   bash -x examples/test-commands.sh
#
# Each block exercises a specific mechanic — see the comment above it.

set -eu

: "${HASDATA_API_KEY:?set HASDATA_API_KEY before running}"

# -----------------------------------------------------------------------------
# Web Scraping
# -----------------------------------------------------------------------------

# 1. Trivial: URL only. Tests required-flag default, GET→POST body assembly.
./hasdata web-scraping --url "https://example.com" --pretty

# 2. Scalars + enums + paired-bool negated forms (--no-block-ads, --no-screenshot).
./hasdata web-scraping \
  --url "https://quotes.toscrape.com" \
  --proxy-country UK \
  --proxy-type residential \
  --no-block-ads \
  --no-screenshot \
  --wait 1500 \
  --pretty

# 3. kvSlice: repeatable --headers key=value. Values with '=' are preserved
#    because parseKVs uses SplitN(s, "=", 2).
./hasdata web-scraping \
  --url "https://httpbin.org/headers" \
  --headers "User-Agent=hasdata-cli-test" \
  --headers "X-Debug=a=b=c" \
  --headers "Accept-Language=en-US,en;q=0.9" \
  --no-screenshot --no-block-resources \
  --pretty

# 4. mergeKVOverride: JSON base + kv overrides per key.
./hasdata web-scraping \
  --url "https://httpbin.org/headers" \
  --headers-json '{"User-Agent":"will-be-overridden","X-Common":"shared"}' \
  --headers "User-Agent=final-winner" \
  --no-screenshot \
  --pretty

# 5. extractRules via kv (lightweight form — no JSON).
./hasdata web-scraping \
  --url "https://quotes.toscrape.com" \
  --extract-rules "quote=.quote .text" \
  --extract-rules "author=.quote .author" \
  --no-screenshot \
  --pretty

# 6. jsonOrFile: @file.
cat > /tmp/rules.json <<'EOF'
{ "title": "h1", "links": "a @href", "first_paragraph": "p" }
EOF
./hasdata web-scraping \
  --url "https://example.com" \
  --extract-rules-json @/tmp/rules.json \
  --no-screenshot \
  --pretty

# 7. jsonOrFile: stdin.
echo '{"headline":"h1"}' | ./hasdata web-scraping \
  --url "https://example.com" \
  --extract-rules-json - \
  --no-screenshot --no-block-resources \
  --pretty

# 8. Complex array (jsScenario) via JSON-only flag. Every flag ending in -json
#    accepts raw JSON, @file, or - (stdin).
./hasdata web-scraping \
  --url "https://quotes.toscrape.com/js/" \
  --js-scenario-json '[{"wait":2000},{"scrollY":500},{"wait":1000}]' \
  --wait-for ".quote" \
  --pretty

# 9. AI extract rules (complex nested object, JSON-only).
./hasdata web-scraping \
  --url "https://news.ycombinator.com" \
  --ai-extract-rules-json '{"top_story":{"type":"string","description":"headline of the top story"}}' \
  --no-screenshot \
  --pretty

# 10. stringSlice multi + enum + --output file.
./hasdata web-scraping \
  --url "https://example.com" \
  --output-format html --output-format markdown \
  --exclude-tags script --exclude-tags style \
  --include-only-tags "main,article" \
  --no-screenshot \
  -o /tmp/result.json
ls -la /tmp/result.json

# -----------------------------------------------------------------------------
# Zillow Listing — bracketed params ([max], [min], []), floats, enum, arrays
# -----------------------------------------------------------------------------

# Each allowed-values list is visible in `./hasdata zillow-listing --help`.

# 11. For-sale: price range + beds/baths + home types. Zillow's enums are
#    lowercase camelCase — the generator picked them up from items.enum in the
#    schema, so `--home-types invalid` fails client-side with the allowed list.
./hasdata zillow-listing \
  --keyword "Austin, TX" \
  --type forSale \
  --price-min 400000 --price-max 900000 \
  --beds-min 3 --beds-max 5 \
  --baths-min 2 \
  --home-types house --home-types townhome \
  --sort priceLowToHigh \
  --pretty

# 12. For-rent: pets/parking + must-have-garage (paired bool default=false),
#    hide 55+ communities.
./hasdata zillow-listing \
  --keyword "Seattle, WA" \
  --type forRent \
  --price-max 4000 \
  --pets allowsSmallDogs --pets allowsCats \
  --parking-spots-min 1 \
  --must-have-garage \
  --hide55plus-communities \
  --pretty

# 13. Sold: square-footage + year-built + lot size + sort.
./hasdata zillow-listing \
  --keyword "Miami, FL" \
  --type sold \
  --square-feet-min 1500 --square-feet-max 4000 \
  --year-built-min 2000 --year-built-max 2020 \
  --lot-size-min 5000 \
  --days-on-zillow 12m \
  --sort newest \
  --pretty

# 14. Heavy: many array filters, keywords refinement, single-story only, tours.
./hasdata zillow-listing \
  --keyword "Denver, CO" \
  --type forSale \
  --keywords "open floor plan" \
  --home-types house \
  --other-amenities pool --other-amenities ac \
  --views mountain --views city \
  --basement finished \
  --property-status comingSoon \
  --listing-publish-options agentListed --listing-publish-options ownerPosted \
  --tours open --tours 3d \
  --single-story-only \
  --sort bedrooms \
  --pretty

# -----------------------------------------------------------------------------
# Extras — debugging / error-path / output
# -----------------------------------------------------------------------------

# 15. --verbose: see request URL + rate-limit headers on stderr.
./hasdata web-scraping --url "https://example.com" --no-screenshot --verbose --pretty

# 16. Enum validation rejects bad value (fails BEFORE hitting the API).
./hasdata web-scraping --url "https://example.com" --proxy-country XX || \
  echo "(expected user error — enum rejected)"

# 17. Paired-bool conflict (fails BEFORE hitting the API).
./hasdata web-scraping --url "https://example.com" --block-ads --no-block-ads || \
  echo "(expected user error — cannot set both)"

# 18. Required flag missing (fails BEFORE hitting the API).
./hasdata web-scraping || echo "(expected user error — --url required)"
