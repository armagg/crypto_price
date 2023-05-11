package db

import (
	"context"
	"crypto_price/pkg/config"
	"time"

	"github.com/go-redis/redis/v8"
)

func CreatRedisClient() *redis.Client {
	config := config.GetConfigs()
	client := redis.NewClient(&redis.Options{
		Addr:     config["redis_host"], // Replace with the address of your Redis instance
		Password: "",               // Set if your Redis instance requires authentication
		DB:       0,                // Specify the Redis database number to use
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
