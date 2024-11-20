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
	"strconv"
	"time"
    "strings"
    "regexp"
)

type PriceInfo struct {
	Price     float64
	Timestamp time.Time
}


type PriceResponse struct {
    Symbol  string  `json:"symbol"`
    Source  string  `json:"source"`
    Price   float64 `json:"price"`
    Elapsed float64 `json:"elapsed"`
    Note    string  `json:"note,omitempty"`
}


func HandlePriceRequest(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	source := r.URL.Query().Get("source")
	quote := r.URL.Query().Get("quote")
	source_usdt := r.URL.Query().Get("source_usdt")

    w.Header().Set("Content-Type", "application/json")

	if base == "" {
		http.Error(w, "Please specify 'base'  to get price", http.StatusBadRequest)
		return
	}
	if quote == "" {
		quote = "USDT"
	}
	if source == "" {
		source = "binance"
	}
	if source_usdt == "" {
		source_usdt = "nobitex"
	}
    if !isValidSymbol(base) || !isValidSource(source) || !isValidQuote(quote) {
        http.Error(w, "Invalid input parameters.", http.StatusBadRequest)
        return
    }

	ctx := r.Context()
    priceInfo := PriceInfo{}
    var err error
	symbol := base + "USDT"


	if quote == "USDT" {
		log.Printf("Fetching price for %s from %s", symbol, source)
		priceInfo, err = getPriceFromRedis(ctx, symbol, source)
		if err != nil {
			log.Printf("Error retrieving price for %s from %s: %v", symbol, source, err)
			http.Error(w, fmt.Sprintf("Error retrieving price: %v", err), http.StatusInternalServerError)
			return

		}
		log.Printf("Fetching price for %s from %s", symbol, source)
	} else if quote == "IRR" || quote == "IRT" {
        usdtPrice, err := GetUsdtIrrFromRedis(ctx, source_usdt, false)
        if err!= nil {
            log.Printf("Error retrieving price for %s from %s: %v", symbol, source_usdt, err)
            http.Error(w, fmt.Sprintf("Error retrieving price: %v", err), http.StatusInternalServerError)
            return
        }

        basePrice, err := getPriceFromRedis(ctx, symbol, source)
		if err != nil {
			log.Printf("Error retrieving price for %s from %s: %v", symbol, source, err)
			http.Error(w, fmt.Sprintf("Error retrieving price: %v", err), http.StatusInternalServerError)
			return

		}

		priceInfo = PriceInfo{
			Price:  basePrice.Price * usdtPrice,
			Timestamp: basePrice.Timestamp,
		}

	} else {
        http.Error(w,  "Invalid input parameters.", http.StatusBadRequest)
        return
    }
    elapsed := time.Since(priceInfo.Timestamp).Seconds()

		response := map[string]interface{}{
			"symbol":  symbol,
			"source":  source,
			"price":   priceInfo.Price,
			"elapsed": elapsed,
		}

		if elapsed > 20 { // Assuming 20 seconds as the threshold for short-term freshness
			response["note"] = "Price may be outdated."
		}

		json.NewEncoder(w).Encode(response)
}

func getPriceFromRedis(ctx context.Context, symbol, source string) (PriceInfo, error) {
	var priceInfo PriceInfo

	// Reuse the Redis client (assumed to be initialized at the package level)
	rdb := db.CreateRedisClient()
    defer rdb.Close()

	// Short-term keys
	shortTermKey := fmt.Sprintf("%s:%s:short", source, symbol)
	shortTermTimeKey := fmt.Sprintf("%s:%s:short:time", source, symbol)

	// Attempt to get the price from the short-term key
	price, err := rdb.Get(ctx, shortTermKey).Result()
	if err == redis.Nil {
		// If the short-term price is not available, fall back to the long-term price
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

func GetUsdtIrrFromRedis(ctx context.Context, source_usdt string, adjust_other_exchanges bool) (float64, error) {
    rdb := db.CreateRedisClient()
    defer rdb.Close()

    usdtIrrKey := fmt.Sprintf("usdtirr:%s", source_usdt)

    value , err := rdb.Get(ctx, usdtIrrKey).Result()
     
    if err != nil {
        if err == redis.Nil {
            return -1 , fmt.Errorf("price not available for usdtirr from %s", source_usdt)
        }
        return -1 , fmt.Errorf("error retrieving usdtirr price from Redis %v", err)
    }
    var priceStruct models.MarketSourceResult
    err = json.Unmarshal([]byte(value), &priceStruct)
    if err != nil {
        return -1 , fmt.Errorf("price not available for usdtirr from %s", source_usdt)
    }

    return priceStruct.WeightedMean, nil
}


func isValidSource(source string) bool {
    validSources := map[string]bool{
        "binance": true,
        "kucoin":  true,
        // Add other valid sources
    }
    return validSources[source]
}

func isValidQuote(quote string) bool {
    validQuotes := map[string]bool{
        "usdt": true,
        "irr":  true,
    }
    return validQuotes[strings.ToLower(quote)]
}

func isValidSymbol(symbol string) bool {
    validSymbol := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
    return validSymbol.MatchString(symbol)
}
