package exchanges

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type PriceResponse struct {
	Data struct {
		Price string `json:"price"`
	} `json:"data"`
}

const (
	baseURL = "https://api.kucoin.com/api/v1/market/orderbook/level1"
)

func GetPricesKucoin(cryptoList []string) (map[string]float64, error) {
	var wg sync.WaitGroup
	cryptoPrices := make(map[string]float64)
	var cryptoPricesMutex sync.Mutex
	var errors []error
	var errorsMutex sync.Mutex

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, crypto := range cryptoList {
		wg.Add(1)
		go func(crypto string) {
			defer wg.Done()

			url := fmt.Sprintf("%s?symbol=%s-USDT", baseURL, crypto)
			resp, err := client.Get(url)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to fetch %s: %w", crypto, err))
				errorsMutex.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to retrieve price for %s: HTTP %d", crypto, resp.StatusCode))
				errorsMutex.Unlock()
				return
			}

			var priceResp PriceResponse
			if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to decode response for %s: %w", crypto, err))
				errorsMutex.Unlock()
				return
			}

			price, err := strconv.ParseFloat(priceResp.Data.Price, 64)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("failed to parse price for %s: %w", crypto, err))
				errorsMutex.Unlock()
				return
			}

			cryptoPricesMutex.Lock()
			cryptoPrices[fmt.Sprintf("%sUSDT", crypto)] = price
			cryptoPricesMutex.Unlock()
		}(crypto)
	}

	wg.Wait()

	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("KuCoin API error: %v\n", err)
		}

		// Return a compound error with all issues
		return cryptoPrices, fmt.Errorf("encountered %d errors while fetching KuCoin prices: %v", len(errors), errors[0])
	}

	return cryptoPrices, nil
}
