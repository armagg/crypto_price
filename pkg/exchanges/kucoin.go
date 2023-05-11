package exchanges 

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var price struct {
    Data struct {
        Price string `json:"price"`
    } `json:"data"`
}

const (
    baseURL = "https://api.kucoin.com/api/v1/market/orderbook/level1"
)

func GetPricesKucoin(cryptoList []string) {
    
    var wg sync.WaitGroup
    wg.Add(len(cryptoList))

    // set up a map to store the prices of all cryptocurrencies
    cryptoPrices := make(map[string]float64)
    cryptoPricesMutex := &sync.Mutex{}
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

            

            err = UnmarshalJSON(body, &price)
            if err != nil {
                errChan <- fmt.Errorf("Failed to unmarshal response for %s. Error: %v", crypto, err)
                return
            }

            p, err := strconv.ParseFloat(price.Data.Price, 128)
			if err != nil {
				errChan <- err
				return
			}
            writeToDict(crypto, p, cryptoPricesMutex, cryptoPrices)
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

    wg.Wait()

    // print the prices
    for crypto, price := range cryptoPrices {
        if price < 1{
        fmt.Printf("%s: %.9f\n", crypto, price)
        } else{
            fmt.Printf("%s: %.2f\n", crypto, price)
        }   
    }
}


func writeToDict(key string, val float64, mutex *sync.Mutex, dict map[string]float64) {
    mutex.Lock()
    defer mutex.Unlock()
    dict[key] = val
}