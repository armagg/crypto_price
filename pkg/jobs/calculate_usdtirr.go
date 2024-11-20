package jobs

import (
	"context"
	"log"
	"sort"
	"time"
	"github.com/go-redis/redis/v8"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"crypto_price/pkg/config"
	"crypto_price/pkg/db"
)

type Transaction struct {
	Price  float64 `bson:"price"`
	Amount float64 `bson:"amount"`
}

type MarketSourceResult struct {
	MarketName   string
	Source       string
	Median       float64
	WeightedMean float64
	StdDev       float64
	SumAmounts   float64
}

const (
	TIME_INTERVAL_TO_LOAD         = 30
	THRESHOLD_FOR_BIG_TRANSACTION = 50
	USDTIRR_TTL                   = 300
)

func calculateUsdtIrrPriceJob() ([]MarketSourceResult, error) {
	var results []MarketSourceResult
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	data := config.GetConfigs()
	client, err := db.CreatMongoClient(ctx, data["MONGO_HOST"])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer client.Disconnect(ctx)

	collection := client.Database(data["TRADE_DATABASE"]).Collection(data["LAST_TRADE_COLLECTION"])

	marketSources := []struct {
		MarketName string
		Source     string
	}{
		{"USDTIRT", "wallex"},
		{"USDTIRT", "nobitex"},
		{"USDTIRT", "bitpin"},
		{"USDTIRR", "ramzinex"},
	}

	for _, ms := range marketSources {
		timeToLoadLastTrades := time.Now()
		if ms.Source == "wallex" || ms.Source == "ramzinex" {
			timeToLoadLastTrades = timeToLoadLastTrades.Add(-TIME_INTERVAL_TO_LOAD * time.Minute).Add(-210 * time.Minute)
		} else {
			timeToLoadLastTrades = timeToLoadLastTrades.Add(-TIME_INTERVAL_TO_LOAD * time.Minute)
		}

		query := bson.M{
			"market_name": ms.MarketName,
			"source":      ms.Source,
			"time":        bson.M{"$gte": timeToLoadLastTrades.Format("2006-01-02T15:04:05")},
			"amount":      bson.M{"$gt": THRESHOLD_FOR_BIG_TRANSACTION},
		}

		cursor, err := collection.Find(context.TODO(), query, options.Find().SetLimit(100))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		var transactions []Transaction

		if err = cursor.All(context.TODO(), &transactions); err != nil {
			log.Println(err)
			return nil, err
		}

		median, weightedMean, stdDev, sumAmounts := calculateStatistics(transactions)

		results = append(results, MarketSourceResult{
			MarketName:   ms.MarketName,
			Source:       ms.Source,
			Median:       median,
			WeightedMean: weightedMean,
			StdDev:       stdDev,
			SumAmounts:   sumAmounts,
		})
	}
	return results, nil
}
func calculateStatistics(transactions []Transaction) (median float64, weightedMean float64, stdDev float64, sumAmounts float64) {
	n := len(transactions)
	if n == 0 {
		return 0, 0, 0, 0
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Price < transactions[j].Price
	})

	// Median
	if n%2 == 0 {
		median = (transactions[n/2-1].Price + transactions[n/2].Price) / 2
	} else {
		median = transactions[n/2].Price
	}

	// Weighted Mean
	var sumWeightedPrices float64 = 0.0
	sumAmounts = 0.0
	for _, t := range transactions {
		sumWeightedPrices += t.Price * t.Amount
		sumAmounts += t.Amount
	}
	weightedMean = sumWeightedPrices / sumAmounts

	// Weighted Standard Deviation
	var sumWeightedSquaredDiffs float64 = 0.0
	for _, t := range transactions {
		diff := t.Price - weightedMean
		sumWeightedSquaredDiffs += t.Amount * diff * diff
	}
	variance := sumWeightedSquaredDiffs / sumAmounts
	stdDev = math.Sqrt(variance)

	return median, weightedMean, stdDev, sumAmounts
}
func StoreUsdtIrrPricesInRedis(rdb *redis.Client, results []MarketSourceResult) error {
    ctx := context.Background()
    for _, result := range results {
        key := fmt.Sprintf("usdtirr:%s", result.Source)
        value, err := json.Marshal(result)
        if err != nil {
            return err
        }
        if err := rdb.Set(ctx, key, value, time.Second * 300).Err(); err != nil {
            return err
        }
    }
    return nil
}
