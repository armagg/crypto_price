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
		Timeout: 5 * time.Second,
	}
	url := "https://api.binance.com/api/v3/ticker/price"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Binance API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve data from Binance. Status code: %v", resp.StatusCode)
	}

	var tickers []BinanceTicker
	err = json.NewDecoder(resp.Body).Decode(&tickers)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response from Binance API: %w", err)
	}

	prices := make(map[string]float64)
	for _, ticker := range tickers {
		price, err := strconv.ParseFloat(ticker.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price for %s: %w", ticker.Symbol, err)
		}
		prices[ticker.Symbol] = price
	}

	return prices, nil
}
