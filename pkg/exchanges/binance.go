package exchanges

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type BinanceTicker struct {
	Symbol string  `json:"symbol"`
	Price  string  `json:"price"` // Adjusted to string to match the JSON response and parse it later
}

func GetAllBinancePrices() (map[string]float64, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := "https://api.binance.com/api/v3/ticker/price"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Binance API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var tickers []BinanceTicker
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from Binance API: %w", err)
	}

	if len(tickers) == 0 {
		return nil, fmt.Errorf("binance API returned empty ticker list")
	}

	prices := make(map[string]float64)
	for _, ticker := range tickers {
		if ticker.Symbol == "" {
			continue
		}

		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price for symbol %s (value: %s): %w", ticker.Symbol, ticker.Price, err)
		}

		if price <= 0 {
			return nil, fmt.Errorf("invalid price for symbol %s: %f", ticker.Symbol, price)
		}

		prices[ticker.Symbol] = price
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no valid prices found in Binance API response")
	}

	return prices, nil
}

func GetNLastCandlesOfBinance(symbol string, limit int) ([]map[string]interface{}, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	if limit <= 0 || limit > 1000 {
		return nil, fmt.Errorf("limit must be between 1 and 1000, got %d", limit)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=1m&limit=%d", symbol, limit)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Binance API for symbol %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned HTTP %d for symbol %s: %s", resp.StatusCode, symbol, resp.Status)
	}

	var candles []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&candles); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response for symbol %s: %w", symbol, err)
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("binance API returned no candle data for symbol %s", symbol)
	}

	// Validate that we have the expected number of candles (or at least some data)
	if len(candles) != limit {
		// This might not be an error if Binance doesn't have enough historical data
		// Just log a warning but don't fail
		fmt.Printf("Warning: Expected %d candles for %s, got %d\n", limit, symbol, len(candles))
	}

	return candles, nil
}

