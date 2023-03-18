package exchanges 

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
    baseURL = "https://api.kucoin.com/api/v1/market/orderbook/level1"
)

func GetAllPrices(cryptoList []string) {
    // list of cryptocurrencies to retrieve the price for

    // set up a wait group for parallel execution
    var wg sync.WaitGroup
    wg.Add(len(cryptoList))

    // set up a map to store the prices of all cryptocurrencies
    cryptoPrices := make(map[string]float64)

    // channel to handle errors
    errChan := make(chan error)

    // set up a HTTP client with timeout options
    client := &http.Client{
        Timeout: 5 * time.Second,
    }

    // loop through each cryptocurrency in the list and send concurrent requests
    for _, crypto := range cryptoList {
        go func(crypto string) {
            // add to the wait group
            defer wg.Done()

            // send a GET request to KuCoin API for the order book level 1
            resp, err := client.Get(fmt.Sprintf("%s?symbol=%s-USDT", baseURL, crypto))
            if err != nil {
                errChan <- err
                return
            }

            // read the response body
            body, err := ioutil.ReadAll(resp.Body)
            resp.Body.Close()
            if err != nil {
                errChan <- err
                return
            }
            // check the status code of the response
            if resp.StatusCode != 200 {
                errChan <- fmt.Errorf("Failed to retrieve price for %s. Status code: %v", crypto, resp.StatusCode)
                return
            }

            // parse the response body for price
            var price struct {
                Data struct {
                    Price string `json:"price"`
                } `json:"data"`
            }

            err = UnmarshalJSON(body, &price)
            if err != nil {
                errChan <- fmt.Errorf("Failed to unmarshal response for %s. Error: %v", crypto, err)
                return
            }

            p, err := strconv.ParseFloat(price.Data.Price, 64)
			if err != nil {
				errChan <- err
				return
			}
			cryptoPrices[crypto] = p
        }(crypto)
    }

    go func() {
        for {
            err := <-errChan
            if err != nil {
                fmt.Println(err)
            }
        }
    }()

    // wait for all requests to complete
    wg.Wait()

    // print the prices
    for crypto, price := range cryptoPrices {
        fmt.Printf("%s: %.2f\n", crypto, price)
    }
}

// UnmarshalJSON is a custom unmarshaler to handle JSON response with unknown fields
func UnmarshalJSON(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}