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

func GetPricesKucoin(cryptoList []string) (map[string]float64, error){
	var wg sync.WaitGroup
	cryptoPrices := make(map[string]float64)
	var cryptoPricesMutex sync.Mutex
	errChan := make(chan error, len(cryptoList)) // Buffered channel to prevent deadlock

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
				errChan <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errChan <- fmt.Errorf("failed to retrieve price for %s. Status code: %v", crypto, resp.StatusCode)
				return
			}

			var priceResp PriceResponse
			if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
				errChan <- err
				return
			}
			price, err := strconv.ParseFloat(priceResp.Data.Price, 64)
			if err != nil {
				errChan <- err
				return
			}

			cryptoPricesMutex.Lock()
			cryptoPrices[fmt.Sprintf("%sUSDT", crypto)] = price
			cryptoPricesMutex.Unlock()
		}(crypto)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		fmt.Println(err) // Handle errors more robustly depending on requirements
	}

	for crypto, price := range cryptoPrices {
		if 0.01 < price && price < 1{
			fmt.Printf("%s: %.4f\n", crypto, price)
		} else if price <= 0.01 {
			fmt.Printf("%s: %.9f\n", crypto, price)
		} else {
			fmt.Printf("%s: %.2f\n", crypto, price)
		}
	}
    return cryptoPrices, nil
}
