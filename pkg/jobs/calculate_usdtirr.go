package jobs

import (
	"context"
	"crypto_price/pkg/config"
	"crypto_price/pkg/db"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	cfg := config.GetConfigs()
	client, err := db.GetMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB client: %w", err)
	}

	_ = ctx // ctx is used implicitly in MongoDB operations

	collection := client.Database(cfg.TradeDatabase).Collection(cfg.LastTradeCollection)

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

		cursor, err := collection.Find(ctx, query, options.Find().SetLimit(100))
		if err != nil {
			log.Printf("Failed to query %s:%s transactions: %v", ms.MarketName, ms.Source, err)
			return nil, fmt.Errorf("failed to query transactions for %s:%s: %w", ms.MarketName, ms.Source, err)
		}
		defer cursor.Close(ctx)

		var transactions []Transaction
		if err = cursor.All(ctx, &transactions); err != nil {
			log.Printf("Failed to decode %s:%s transactions: %v", ms.MarketName, ms.Source, err)
			return nil, fmt.Errorf("failed to decode transactions for %s:%s: %w", ms.MarketName, ms.Source, err)
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
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}

	ctx := context.Background()
	for _, result := range results {
		key := fmt.Sprintf("usdtirr:%s", result.Source)
		value, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result for %s: %w", result.Source, err)
		}
		if err := rdb.Set(ctx, key, value, time.Second*USDTIRR_TTL).Err(); err != nil {
			return fmt.Errorf("failed to store USDTIRR price for %s: %w", result.Source, err)
		}
	}
	return nil
}
