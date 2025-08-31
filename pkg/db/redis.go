package db

import (
	"context"
	"crypto_price/pkg/config"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient     *redis.Client
	redisClientOnce sync.Once
	redisClientErr  error
)

func GetRedisClient() (*redis.Client, error) {
	redisClientOnce.Do(func() {
		cfg := config.GetConfigs()
		redisClient = redis.NewClient(&redis.Options{
			Addr:         cfg.RedisHost,
			Password:     "",
			DB:           0,
			PoolSize:     10,
			MinIdleConns: 2,  
			MaxConnAge:   30 * time.Minute,
			IdleTimeout:  5 * time.Minute,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			redisClientErr = fmt.Errorf("failed to connect to Redis: %w", err)
			return
		}

		log.Println("Redis connection pool initialized successfully")
	})

	return redisClient, redisClientErr
}

// CreateRedisClient is kept for backward compatibility but marked as deprecated
// Use GetRedisClient() instead for better performance
func CreateRedisClient() *redis.Client {
	client, err := GetRedisClient()
	if err != nil {
		log.Printf("Redis connection error: %v", err)
		return nil
	}
	return client
}


// Redis expiration constants
const (
	RedisLongTermExpiration  = 10 * time.Minute
	RedisShortTermExpiration = 20 * time.Second
)

func StorePricesInRedis(client *redis.Client, prices map[string]float64, source string) error {
	if client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	now := time.Now()
	ctx := context.Background()
	for symbol, price := range prices {
		// Long-term key with expiration
		longTermKey := fmt.Sprintf("%s:%s:long", source, symbol)
		longTermTimeKey := fmt.Sprintf("%s:%s:long:time", source, symbol)
		// Short-term key with expiration
		shortTermKey := fmt.Sprintf("%s:%s:short", source, symbol)
		shortTermTimeKey := fmt.Sprintf("%s:%s:short:time", source, symbol)

		// Use a transaction to set both keys atomically
		_, err := client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, longTermKey, price, RedisLongTermExpiration)
			pipe.Set(ctx, shortTermKey, price, RedisShortTermExpiration)
			pipe.Set(ctx, longTermTimeKey, now.Unix(), RedisLongTermExpiration)
			pipe.Set(ctx, shortTermTimeKey, now.Unix(), RedisShortTermExpiration)
			return nil
		})

		if err != nil {
			log.Printf("Error storing price for %s:%s: %v", source, symbol, err)
			return fmt.Errorf("failed to store price for %s:%s: %w", source, symbol, err)
		}
	}

	return nil
}