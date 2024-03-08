package jobs

import (
	"fmt"
	"log"
	"time"
	// "crypto_price/pkg/db"
	// "crypto_price/pkg/exchanges"
)

func GetData(){
	tickerCalculateUsdtirr := time.NewTicker(2 * time.Minute)
	tickerKucoin := time.NewTicker(15 * time.Second)
	tickerBinance := time.NewTicker(15 * time.Second)

	rdb := db.CreatRedisClient()
	go func(){
		for ; true; <-tickerCalculateUsdtirr.C {
			log.Println("Starting to calculate usdtirr")
		    results, err := calculateUsdtIrrPriceJob()
			if err != nil{
				log.Println("Error calculating usdtirr price:", err)
			}
			fmt.Println(results)
		}

	}()
	
	go func() {
		for ; true; <-tickerKucoin.C {
			symbols, err := db.GetKucoinSymbolsFromDB()
			if err != nil {
				log.Println("Error fetching symbols from DB:", err)
				continue
			}
			prices, err := exchanges.GetPricesKucoin(symbols)
			if err != nil {
				log.Println("Error fetching Kucoin prices:", err)
			} else {
				if err := db.StorePricesInRedis(rdb, prices, "Kucoin"); err != nil {
					log.Println("Error storing prices in Redis:", err)
				}

			}
		}
	}()

	go func() {
		for ; true; <-tickerBinance.C {
			prices, err := exchanges.GetAllBinancePrices()
			if err != nil {
				log.Println("Error fetching Binance prices:", err)
			} else {
				if err := db.StorePricesInRedis(rdb, prices, "Binance"); err != nil {
					log.Println("Error storing prices in Redis:", err)
			}
			}			
		}
	}()

}
