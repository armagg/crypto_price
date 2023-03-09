package binance

import (
    "encoding/json"
    "net/http"
)

type BinanceTicker struct {
    Symbol string  `json:"symbol"`
    Price  float64 `json:"price"`
}

func GetAllBinancePrices() ([]BinanceTicker, error) {
    url := "https://api.binance.com/api/v3/ticker/price"
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var tickers []BinanceTicker
    err = json.NewDecoder(resp.Body).Decode(&tickers)
    if err != nil {
        return nil, err
    }

    return tickers, nil
}
