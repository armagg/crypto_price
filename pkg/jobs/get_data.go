package jobs

import (
	"log"
	"time"
	"crypto_price/pkg/db"
	"crypto_price/pkg/exchanges"
)

func GetData(){
	tickerCalculateUsdtirr := time.NewTicker(2 * time.Minute)
	tickerKucoin := time.NewTicker(15 * time.Second)
	tickerBinance := time.NewTicker(15 * time.Second)

	go func(){
        for range tickerCalculateUsdtirr.C {
            rdb := db.CreateRedisClient()
            defer rdb.Close()        
            log.Println("Starting to calculate usdtirr")
            results, err := calculateUsdtIrrPriceJob()
            if err != nil {
                log.Println("Error calculating USDTIRR price:", err)
            } else {
                log.Println(results)
                // Store results in Redis
                if err := StoreUsdtIrrPricesInRedis(rdb, results); err != nil {
                    log.Println("Error storing USDTIRR prices in Redis:", err)
                }
            }
        }
    }()
	
	go func() {
        for range tickerKucoin.C {
            rdb := db.CreateRedisClient()
            defer rdb.Close()        
            symbols, err := db.GetKucoinSymbolsFromDB()
            if err != nil {
                log.Println("Error fetching symbols from DB:", err)
                continue
            }
            prices, err := exchanges.GetPricesKucoin(symbols)
            if err != nil {
                log.Println("Error fetching Kucoin prices:", err)
            } else {
                if err := db.StorePricesInRedis(rdb, prices, "kucoin"); err != nil {
                    log.Println("Error storing prices in Redis:", err)
                }
            }
        }
    }()

    go func() {
        for range tickerBinance.C {
            rdb := db.CreateRedisClient()
            defer rdb.Close()        
            prices, err := exchanges.GetAllBinancePrices()
            if err != nil {
                log.Println("Error fetching Binance prices:", err)
            } else {
                if err := db.StorePricesInRedis(rdb, prices, "binance"); err != nil {
                    log.Println("Error storing prices in Redis:", err)
                }
            }
        }
    }()
}
