package controller


import (
	"time"
	"fmt"
	"strconv"
	"net/http"
	"encoding/json"
	"crypto_price/pkg/db"
	"github.com/go-redis/redis/v8"
	"context"
)


type PriceInfo struct {
    Price     float64
    Timestamp time.Time
}

func HandlePriceRequest(w http.ResponseWriter, r *http.Request) {
    symbol := r.URL.Query().Get("symbol")
    source := r.URL.Query().Get("source")

    if symbol == "" || source == "" {
        http.Error(w, "Please specify both 'symbol' and 'source' query parameters.", http.StatusBadRequest)
        return
    }

    priceInfo, err := getPriceFromRedis(symbol, source)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error retrieving price: %v", err), http.StatusInternalServerError)
        return
    }

    elapsed := time.Since(priceInfo.Timestamp).Seconds()

    response := map[string]interface{}{
        "symbol": symbol,
        "source": source,
        "price":  priceInfo.Price,
        "elapsed": elapsed,
    }

    if elapsed > 20 { // Assuming 20 seconds as the threshold for short-term freshness
        response["note"] = "Price may be outdated."
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func getPriceFromRedis(symbol, source string) (PriceInfo, error) {
    var priceInfo PriceInfo
	rdb := db.CreatRedisClient()
	ctx := context.Background()
    // Attempt to get the price from the short-term key
    shortTermKey := fmt.Sprintf("%s:%s:short", source, symbol)
    price, err := rdb.Get(ctx, shortTermKey).Result()
    if err == redis.Nil {
        // If the short-term price is not available, fall back to the long-term price
        longTermKey := fmt.Sprintf("%s:%s:long", source, symbol)
        longTermTimeKey := fmt.Sprintf("%s:%s:long:time", source, symbol)

        price, err = rdb.Get(ctx, longTermKey).Result()
        if err != nil {
            return priceInfo, fmt.Errorf("price not available for %s from %s", symbol, source)
        }

        // Get the timestamp from the long-term time key
        timestamp, err := rdb.Get(ctx, longTermTimeKey).Int64()
        if err != nil {
            return priceInfo, fmt.Errorf("timestamp not available for %s from %s", symbol, source)
        }

        priceInfo.Timestamp = time.Unix(timestamp, 0)
    } else if err != nil {
        return priceInfo, fmt.Errorf("error retrieving price from Redis: %v", err)
    } else {
        // If the short-term price is available, use the current time as the timestamp
        priceInfo.Timestamp = time.Now()
    }

    priceInfo.Price, err = strconv.ParseFloat(price, 64)
    if err != nil {
        return priceInfo, fmt.Errorf("error parsing price for %s from %s: %v", symbol, source, err)
    }

    return priceInfo, nil
}