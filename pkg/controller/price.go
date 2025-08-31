package controller

import (
	"context"
	"crypto_price/pkg/db"
	"crypto_price/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
    DEFAULT_SOURCE = "binance"
    DEFAULT_QUOTE   = "usdt"
    DEFAULT_USDT_SOURCE = "nobitex"

    REDIS_LONG_TERM_EXPIRATION  = 10 * time.Minute
    REDIS_SHORT_TERM_EXPIRATION = 20 * time.Second

    PRICE_FRESHNESS_THRESHOLD = 20 * time.Second
)

type PriceInfo struct {
	Price     float64
	Timestamp time.Time
}

type PriceResponse struct {
	Symbol     string  `json:"symbol"`
	Source     string  `json:"source"`
	Price      float64 `json:"price"`
	Elapsed    float64 `json:"elapsed"`
	SourceUsdt string  `json:"source_usdt"`
	Quote      string  `json:"quote"`
	Note       string  `json:"note,omitempty"`
}

// HandlePriceRequest handles the incoming price request and returns the price information.
func HandlePriceRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse and validate query parameters
	base, source, quote, sourceUsdt, err := parseAndValidateParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the price information
	priceInfo, err := fetchPrice(r.Context(), base, source, quote, sourceUsdt)
	if err != nil {
		log.Printf("Error fetching price: %v", err)
		http.Error(w, fmt.Sprintf("Error retrieving price: %v", err), http.StatusInternalServerError)
		return
	}

	// Build the response
	response := buildPriceResponse(base, source, quote, sourceUsdt, priceInfo)

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// parseAndValidateParams parses and validates the query parameters.
func parseAndValidateParams(r *http.Request) (base, source, quote, sourceUsdt string, err error) {
	query := r.URL.Query()
	base = query.Get("base")
	source = query.Get("source")
	quote = query.Get("quote")
	sourceUsdt = query.Get("source_usdt")

	if base == "" {
		return "", "", "", "", fmt.Errorf("please specify 'base' to get price")
	}
	if !isValidSymbol(base) {
		return "", "", "", "", fmt.Errorf("invalid 'base' parameter")
	}

	if quote == "" {
		quote = DEFAULT_QUOTE
	} else if !isValidQuote(quote) {
		return "", "", "", "", fmt.Errorf("invalid 'quote' parameter")
	}

	if source == "" {
		source = DEFAULT_SOURCE
	} else if !isValidSource(source) {
		return "", "", "", "", fmt.Errorf("invalid 'source' parameter")
	}

	if sourceUsdt == "" {
		sourceUsdt = DEFAULT_USDT_SOURCE
	}

	return base, source, quote, sourceUsdt, nil
}

// fetchPrice retrieves the price information based on the provided parameters.
func fetchPrice(ctx context.Context, base, source, quote, sourceUsdt string) (PriceInfo, error) {
	// Validate inputs
	if base == "" {
		return PriceInfo{}, fmt.Errorf("base currency cannot be empty")
	}
	if source == "" {
		return PriceInfo{}, fmt.Errorf("source cannot be empty")
	}

	var priceInfo PriceInfo
	symbol := base + "USDT"

	switch strings.ToUpper(quote) {
	case "USDT":
		log.Printf("Fetching price for %s from %s", symbol, source)
		price, err := getPriceFromRedis(ctx, symbol, source)
		if err != nil {
			return PriceInfo{}, fmt.Errorf("failed to retrieve USDT price for %s from %s: %w", symbol, source, err)
		}
		priceInfo = price

	case "IRR", "IRT":
		usdtPrice, err := getUsdtIrrFromRedis(ctx, sourceUsdt)
		if err != nil {
			return PriceInfo{}, fmt.Errorf("failed to retrieve USDT/%s conversion rate from %s: %w", strings.ToUpper(quote), sourceUsdt, err)
		}

		if usdtPrice <= 0 {
			return PriceInfo{}, fmt.Errorf("invalid USDT/%s conversion rate: %f", strings.ToUpper(quote), usdtPrice)
		}

		if base == "USDT" || base == "USDC" {
			priceInfo = PriceInfo{
				Price:     usdtPrice,
				Timestamp: time.Now(),
			}
		} else {
			basePrice, err := getPriceFromRedis(ctx, symbol, source)
			if err != nil {
				return PriceInfo{}, fmt.Errorf("failed to retrieve base price for %s from %s: %w", symbol, source, err)
			}

			if basePrice.Price <= 0 {
				return PriceInfo{}, fmt.Errorf("invalid base price for %s: %f", symbol, basePrice.Price)
			}

			priceInfo = PriceInfo{
				Price:     basePrice.Price * usdtPrice,
				Timestamp: basePrice.Timestamp,
			}
		}

	default:
		return PriceInfo{}, fmt.Errorf("unsupported quote currency: %s (supported: USDT, IRR, IRT)", quote)
	}

	if priceInfo.Price <= 0 {
		return PriceInfo{}, fmt.Errorf("calculated price is invalid: %f", priceInfo.Price)
	}

	return priceInfo, nil
}

func buildPriceResponse(base, source, quote, sourceUsdt string, priceInfo PriceInfo) PriceResponse {
	elapsed := time.Since(priceInfo.Timestamp).Seconds()
	symbol := base + "USDT"

	response := PriceResponse{
		Symbol:     symbol,
		Source:     source,
		Price:      priceInfo.Price,
		Elapsed:    elapsed,
		SourceUsdt: sourceUsdt,
		Quote:      quote,
	}

	if elapsed > PRICE_FRESHNESS_THRESHOLD.Seconds() { // Price freshness threshold
		response.Note = "Price may be outdated."
	}

	return response
}

// getPriceFromRedis retrieves the price of a symbol from Redis.
func getPriceFromRedis(ctx context.Context, symbol, source string) (PriceInfo, error) {
	var priceInfo PriceInfo

	rdb, err := db.GetRedisClient()
	if err != nil {
		return priceInfo, fmt.Errorf("failed to get Redis client: %w", err)
	}

	// Short-term keys
	shortTermKey := fmt.Sprintf("%s:%s:short", source, symbol)
	shortTermTimeKey := fmt.Sprintf("%s:%s:short:time", source, symbol)

	// Attempt to get the price from the short-term key
	price, err := rdb.Get(ctx, shortTermKey).Result()
	if err == redis.Nil {
		// Fall back to the long-term price
		longTermKey := fmt.Sprintf("%s:%s:long", source, symbol)
		longTermTimeKey := fmt.Sprintf("%s:%s:long:time", source, symbol)

		price, err = rdb.Get(ctx, longTermKey).Result()
		if err != nil {
			if err == redis.Nil {
				return priceInfo, fmt.Errorf("price not available for %s from %s", symbol, source)
			}
			return priceInfo, fmt.Errorf("error retrieving long-term price from Redis: %v", err)
		}

		// Get the timestamp from the long-term time key
		timestamp, err := rdb.Get(ctx, longTermTimeKey).Int64()
		if err != nil {
			return priceInfo, fmt.Errorf("timestamp not available for %s from %s", symbol, source)
		}
		priceInfo.Timestamp = time.Unix(timestamp, 0)
	} else if err != nil {
		// Error retrieving short-term price
		return priceInfo, fmt.Errorf("error retrieving short-term price from Redis: %v", err)
	} else {
		// Successfully retrieved short-term price, get the timestamp
		timestamp, err := rdb.Get(ctx, shortTermTimeKey).Int64()
		if err != nil {
			// If timestamp is not available, use the current time
			priceInfo.Timestamp = time.Now()
		} else {
			priceInfo.Timestamp = time.Unix(timestamp, 0)
		}
	}

	// Parse the price
	priceInfo.Price, err = strconv.ParseFloat(price, 64)
	if err != nil {
		return priceInfo, fmt.Errorf("error parsing price for %s from %s: %v", symbol, source, err)
	}

	return priceInfo, nil
}

// getUsdtIrrFromRedis retrieves the USDT to IRR conversion rate from Redis.
func getUsdtIrrFromRedis(ctx context.Context, sourceUsdt string) (float64, error) {
	if sourceUsdt == "" {
		return -1, fmt.Errorf("sourceUsdt cannot be empty")
	}

	rdb, err := db.GetRedisClient()
	if err != nil {
		return -1, fmt.Errorf("failed to get Redis client: %w", err)
	}

	usdtIrrKey := fmt.Sprintf("usdtirr:%s", sourceUsdt)

	value, err := rdb.Get(ctx, usdtIrrKey).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, fmt.Errorf("USDT/%s conversion rate not available in cache (key: %s)", strings.ToUpper(sourceUsdt), usdtIrrKey)
		}
		return -1, fmt.Errorf("failed to retrieve USDT/%s rate from Redis: %w", strings.ToUpper(sourceUsdt), err)
	}

	var priceStruct models.MarketSourceResult
	if err := json.Unmarshal([]byte(value), &priceStruct); err != nil {
		return -1, fmt.Errorf("failed to parse USDT/%s rate data from Redis: %w", strings.ToUpper(sourceUsdt), err)
	}

	if priceStruct.WeightedMean <= 0 {
		return -1, fmt.Errorf("invalid USDT/%s conversion rate: %f (must be positive)", strings.ToUpper(sourceUsdt), priceStruct.WeightedMean)
	}

	return priceStruct.WeightedMean, nil
}

func isValidSource(source string) bool {
	if source == "" {
		return false
	}
	validSources := map[string]bool{
		"binance": true,
		"kucoin":  true,
	}
	return validSources[strings.ToLower(source)]
}

// isValidQuote checks if the provided quote is valid.
func isValidQuote(quote string) bool {
	if quote == "" {
		return false
	}
	validQuotes := map[string]bool{
		"usdt": true,
		"irr":  true,
		"irt":  true,
	}
	return validQuotes[strings.ToLower(quote)]
}

// isValidSymbol checks if the provided symbol is valid.
func isValidSymbol(symbol string) bool {
	if symbol == "" || len(symbol) > 10 {
		return false
	}
	validSymbol := regexp.MustCompile(`^[A-Z0-9_]{1,10}$`)
	return validSymbol.MatchString(strings.ToUpper(symbol))
}

