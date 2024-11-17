package db

import (
	"context"
	"crypto_price/pkg/config"
	"time"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func CreateRedisClient() *redis.Client {
	config := config.GetConfigs()
	client := redis.NewClient(&redis.Options{
		Addr:     config["REDIS_HOST"], 
		Password: "",               
		DB:       0,                
		
	})

	// set timeout for connect 5 Seconds
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return client
}

// func CreateBroker(ctx context.Context, name string) (*Broker, error) {
// 	# create a new broker in local redis database
// }

// func GetBroker(ctx context.Context, name string) (*Broker, error) {
// 	# pass the name to the function to get the broker
// }

// func DeleteBroker(ctx context.Context, name string) error {}



func StorePricesInRedis(client *redis.Client, prices map[string]float64, source string) error {
    
	now := time.Now()
	ctx := context.Background()
    for symbol, price := range prices {
        // Long-term key with a 10-minute expiration
        longTermKey := fmt.Sprintf("%s:%s:long", source, symbol)
		longTermTimeKey := fmt.Sprintf("%s:%s:long:time", source, symbol)
        // Short-term key with a 20-second expiration
        shortTermKey := fmt.Sprintf("%s:%s:short", source, symbol)
		shortTermTimeKey := fmt.Sprintf("%s:%s:short:time", source, symbol)
        // Use a transaction to set both keys atomically
        _, err := client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
            pipe.Set(ctx, longTermKey, price, 10*time.Minute)
            pipe.Set(ctx, shortTermKey, price, 20*time.Second)
			pipe.Set(ctx, longTermTimeKey, now.Unix(), 10*time.Minute)
			pipe.Set(ctx, shortTermTimeKey, now.Unix(), 20*time.Second)
            return nil
        })

        if err != nil {
            return err
        }

        // fmt.Printf("[%s] Stored %s price from %s: %f\n", now.Format(time.RFC3339), symbol, source, price)
    }

    return nil
}