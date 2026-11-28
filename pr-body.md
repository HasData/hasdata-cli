## Changes

 GENERATION_REPORT.md                |  97 +++++++++++++++++--------------
 cmd/gen_airbnb_listing.go           |   2 +-
 cmd/gen_airbnb_property.go          |   2 +-
 cmd/gen_amazon_product.go           |   2 +-
 cmd/gen_amazon_search.go            |   2 +-
 cmd/gen_amazon_seller.go            |   2 +-
 cmd/gen_amazon_seller_products.go   |   2 +-
 cmd/gen_booking_place.go            |   2 +-
 cmd/gen_booking_search.go           |   2 +-
 cmd/gen_chat_gpt_chat.go            |   2 +-
 cmd/gen_glassdoor_job.go            |   2 +-
 cmd/gen_glassdoor_listing.go        |   2 +-
 cmd/gen_google_ai_mode.go           |   2 +-
 cmd/gen_google_events.go            |   2 +-
 cmd/gen_google_finance.go           |   2 +-
 cmd/gen_google_flights.go           |   2 +-
 cmd/gen_google_flights_deals.go     |   2 +-
 cmd/gen_google_gemini.go            |   2 +-
 cmd/gen_google_hotels.go            |   2 +-
 cmd/gen_google_images.go            |   2 +-
 cmd/gen_google_immersive_product.go |   2 +-
 cmd/gen_google_news.go              |   2 +-
 cmd/gen_google_serp.go              |   2 +-
 cmd/gen_google_serp_light.go        |   2 +-
 cmd/gen_google_shopping.go          |   2 +-
 cmd/gen_google_short_videos.go      |   2 +-
 cmd/gen_google_trends.go            |   2 +-
 cmd/gen_indeed_job.go               |   2 +-
 cmd/gen_indeed_listing.go           |   2 +-
 cmd/gen_redfin_listing.go           |   2 +-
 cmd/gen_redfin_property.go          |   2 +-
 cmd/gen_shopify_collections.go      |   2 +-
 cmd/gen_shopify_products.go         |   2 +-
 cmd/gen_tiktok_shop_category_api.go |   2 +-
 cmd/gen_tiktok_shop_product_api.go  |   2 +-
 cmd/gen_tiktok_shop_search_api.go   |   2 +-
 cmd/gen_tiktok_shop_store_api.go    |   2 +-
 cmd/gen_tripadvisor_place.go        |   2 +-
 cmd/gen_tripadvisor_reviews.go      |   2 +-
 cmd/gen_tripadvisor_search.go       |   2 +-
 cmd/gen_trustpilot_company_api.go   |   2 +-
 cmd/gen_trustpilot_reviews_api.go   |   2 +-
 cmd/gen_trustpilot_search_api.go    |   2 +-
 cmd/gen_walmart_product.go          |   2 +-
 cmd/gen_walmart_reviews.go          |   2 +-
 cmd/gen_walmart_search.go           |   2 +-
 cmd/gen_web_scraping.go             |   2 +-
 cmd/gen_yellowpages_place.go        |   2 +-
 cmd/gen_yellowpages_search.go       |   2 +-
 cmd/gen_yelp_place.go               |   2 +-
 cmd/gen_yelp_reviews.go             |   2 +-
 cmd/gen_yelp_search.go              |   2 +-
 cmd/gen_zillow_listing.go           |   2 +-
 cmd/gen_zillow_property.go          |   2 +-
 internal/gen/spec-hash.txt          |   2 +-
 pr-body.md                          | 111 ------------------------------------
 56 files changed, 107 insertions(+), 209 deletions(-)

## Generation Report
# Generation Report

Generated at: 2026-11-28T07:03:35Z

## APIs generated (87)
- airbnb-listing
- airbnb-property
- amazon-product
- amazon-search
- amazon-seller
- amazon-seller-products
- web-scraping
- bing-serp
- booking-place
- booking-search
- chat-gpt-chat
- duckduckgo
- facebook-profile
- facebook-reel-api
- facebook-reels-api
- google-gemini
- glassdoor-job
- glassdoor-listing
- google-maps
- google-maps-contributor-reviews
- google-maps-photos
- google-maps-place
- google-maps-posts
- google-maps-reviews
- google-scholar
- google-scholar-cite
- google-ai-mode
- google-events
- google-finance
- google-images
- google-news
- google-serp
- google-serp-light
- google-short-videos
- google-trends
- google-immersive-product
- google-shopping
- google-flights
- google-flights-deals
- google-hotels
- indeed-job
- indeed-listing
- instagram-comments
- instagram-location-api
- instagram-popular-api
- instagram-posts
- instagram-profile
- instagram-reels
- instagram-trending-api
- redfin-listing
- redfin-property
- shopify-collections
- shopify-products
- skyscanner-flights-api
- skyscanner-hotel-api
- skyscanner-hotels-api
- tiktok-comments
- tiktok-discover-api
- tiktok-posts
- tiktok-profile
- tiktok-search
- tiktok-transcript-api
- tiktok-shop-category-api
- tiktok-shop-product-api
- tiktok-shop-search-api
- tiktok-shop-store-api
- tripadvisor-place
- tripadvisor-reviews
- tripadvisor-search
- trustpilot-company-api
- trustpilot-reviews-api
- trustpilot-search-api
- walmart-product
- walmart-reviews
- walmart-search
- yellowpages-place
- yellowpages-search
- yelp-place
- yelp-reviews
- yelp-search
- youtube-channel-api
- youtube-comments-api
- youtube-search-api
- youtube-transcript-api
- youtube-video-api
- zillow-listing
- zillow-property
