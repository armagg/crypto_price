package controller

import (
	"context"
	"crypto_price/pkg/db"
	"crypto_price/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
    DEFUALT_SOURCE = "binance"
    DEFUALT_QUOTE   = "usdt"
    DEFUALT_USDT_SOURCE = "nobitex"
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
		quote = DEFUALT_QUOTE
	} else if !isValidQuote(quote) {
		return "", "", "", "", fmt.Errorf("invalid 'quote' parameter")
	}

	if source == "" {
		source = DEFUALT_SOURCE
	} else if !isValidSource(source) {
		return "", "", "", "", fmt.Errorf("invalid 'source' parameter")
	}

	if sourceUsdt == "" {
		sourceUsdt = DEFUALT_USDT_SOURCE
	}

	return base, source, quote, sourceUsdt, nil
}

// fetchPrice retrieves the price information based on the provided parameters.
func fetchPrice(ctx context.Context, base, source, quote, sourceUsdt string) (PriceInfo, error) {
	var priceInfo PriceInfo
	symbol := base + "USDT"

	switch strings.ToUpper(quote) {
	case "USDT":
		log.Printf("Fetching price for %s from %s", symbol, source)
		price, err := getPriceFromRedis(ctx, symbol, source)
		if err != nil {
			return PriceInfo{}, fmt.Errorf("error retrieving price for %s from %s: %v", symbol, source, err)
		}
		priceInfo = price

	case "IRR", "IRT":
		usdtPrice, err := getUsdtIrrFromRedis(ctx, sourceUsdt)
		if err != nil {
			return PriceInfo{}, fmt.Errorf("error retrieving USDT price from %s: %v", sourceUsdt, err)
		}

		if base == "USDT" || base == "USDC" {
			priceInfo = PriceInfo{
				Price:     usdtPrice,
				Timestamp: time.Now(),
			}
		} else {
			basePrice, err := getPriceFromRedis(ctx, symbol, source)
			if err != nil {
				return PriceInfo{}, fmt.Errorf("error retrieving price for %s from %s: %v", symbol, source, err)
			}
			priceInfo = PriceInfo{
				Price:     basePrice.Price * usdtPrice,
				Timestamp: basePrice.Timestamp,
			}
		}

	default:
		return PriceInfo{}, fmt.Errorf("invalid 'quote' parameter")
	}

	return priceInfo, nil
}

// buildPriceResponse constructs the response to be sent back to the client.
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

	if elapsed > 20 { // Assuming 20 seconds as the threshold for freshness
		response.Note = "Price may be outdated."
	}

	return response
}

// getPriceFromRedis retrieves the price of a symbol from Redis.
func getPriceFromRedis(ctx context.Context, symbol, source string) (PriceInfo, error) {
	var priceInfo PriceInfo

	rdb := db.CreateRedisClient()
    defer rdb.Close()

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
	rdb := db.CreateRedisClient()
    defer rdb.Close()
	usdtIrrKey := fmt.Sprintf("usdtirr:%s", sourceUsdt)

	value, err := rdb.Get(ctx, usdtIrrKey).Result()
	if err != nil {
		if err == redis.Nil {
			return -1, fmt.Errorf("price not available for USDTIRR from %s", sourceUsdt)
		}
		return -1, fmt.Errorf("error retrieving USDTIRR price from Redis: %v", err)
	}

	var priceStruct models.MarketSourceResult
	if err := json.Unmarshal([]byte(value), &priceStruct); err != nil {
		return -1, fmt.Errorf("error parsing USDTIRR price from %s: %v", sourceUsdt, err)
	}

	return priceStruct.WeightedMean, nil
}

func isValidSource(source string) bool {
	validSources := map[string]bool{
		"binance": true,
		"kucoin":  true,
	}
	return validSources[strings.ToLower(source)]
}


// isValidQuote checks if the provided quote is valid.
func isValidQuote(quote string) bool {
	validQuotes := map[string]bool{
		"usdt": true,
		"irr":  true,
		"irt":  true,
	}
	return validQuotes[strings.ToLower(quote)]
}

// isValidSymbol checks if the provided symbol is valid.
func isValidSymbol(symbol string) bool {
	validSymbol := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	return validSymbol.MatchString(symbol)
}

