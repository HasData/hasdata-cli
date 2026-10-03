## Changes

 GENERATION_REPORT.md                       | 31 +++++++----
 cmd/gen_airbnb_listing.go                  | 10 ++--
 cmd/gen_airbnb_property.go                 |  4 +-
 cmd/gen_amazon_product.go                  |  8 +--
 cmd/gen_amazon_reviews.go                  |  6 +--
 cmd/gen_amazon_search.go                   |  6 +--
 cmd/gen_amazon_seller.go                   |  6 +--
 cmd/gen_amazon_seller_products.go          |  6 +--
 cmd/gen_booking_place.go                   |  4 +-
 cmd/gen_booking_search.go                  |  4 +-
 cmd/gen_glassdoor_job.go                   |  4 +-
 cmd/gen_glassdoor_listing.go               | 56 ++++++++++++++++++--
 cmd/gen_google_ai_mode.go                  |  4 +-
 cmd/gen_google_events.go                   | 55 +++----------------
 cmd/gen_google_flights.go                  |  4 +-
 cmd/gen_google_hotels.go                   |  4 +-
 cmd/gen_google_images.go                   |  4 +-
 cmd/gen_google_immersive_product.go        |  4 +-
 cmd/gen_google_maps.go                     |  4 +-
 cmd/gen_google_maps_contributor_reviews.go |  4 +-
 cmd/gen_google_maps_photos.go              |  4 +-
 cmd/gen_google_maps_place.go               |  9 +++-
 cmd/gen_google_maps_posts.go               |  5 --
 cmd/gen_google_maps_reviews.go             |  4 +-
 cmd/gen_google_short_videos.go             |  9 +++-
 cmd/gen_google_trends.go                   |  4 +-
 cmd/gen_indeed_job.go                      |  4 +-
 cmd/gen_indeed_listing.go                  |  4 +-
 cmd/gen_instagram_posts.go                 |  4 +-
 cmd/gen_instagram_profile.go               |  4 +-
 cmd/gen_redfin_listing.go                  |  6 +--
 cmd/gen_redfin_property.go                 |  6 +--
 cmd/gen_shopify_collections.go             |  6 +--
 cmd/gen_shopify_products.go                |  6 +--
 cmd/gen_tiktok_comments.go                 |  2 +-
 cmd/gen_yellowpages_place.go               |  4 +-
 cmd/gen_yellowpages_search.go              |  4 +-
 cmd/gen_zillow_listing.go                  | 23 ++------
 cmd/gen_zillow_property.go                 |  6 +--
 internal/gen/spec-hash.txt                 |  2 +-
 pr-body.md                                 | 84 ------------------------------
 41 files changed, 176 insertions(+), 252 deletions(-)

## Generation Report
# Generation Report

Generated at: 2026-10-03T07:13:50Z

## APIs generated (64)
- chat-gpt-chat
- airbnb-listing
- airbnb-property
- bing-serp
- booking-place
- booking-search
- yellowpages-place
- yellowpages-search
- yelp-place
- yelp-reviews
- yelp-search
- duckduckgo
- amazon-product
- amazon-reviews
- amazon-search
- amazon-seller
- amazon-seller-products
- shopify-collections
- shopify-products
- walmart-product
- walmart-reviews
- walmart-search
- google-images
- google-maps
- google-maps-contributor-reviews
- google-maps-photos
- google-maps-place
- google-maps-posts
- google-maps-reviews
- google-ai-mode
- google-events
- google-immersive-product
- google-news
- google-serp
- google-serp-light
- google-shopping
- google-short-videos
- google-scholar
- google-scholar-cite
- google-flights
- google-flights-deals
- google-hotels
- google-trends
- glassdoor-job
- glassdoor-listing
- indeed-job
- indeed-listing
- redfin-listing
- redfin-property
- zillow-listing
- zillow-property
- facebook-profile
- instagram-comments
- instagram-posts
- instagram-profile
- tiktok-comments
- tiktok-posts
- tiktok-profile
- tiktok-search
- web-scraping
- youtube-channel-api
- youtube-search-api
- youtube-transcript-api
- youtube-video-api

## Deprecated APIs (hidden in help)
- amazon-reviews
