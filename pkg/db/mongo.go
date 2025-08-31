package db

import (
	"context"
	"crypto_price/pkg/config"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient     *mongo.Client
	mongoClientOnce sync.Once
	mongoClientErr  error
)

// GetMongoClient returns a singleton MongoDB client with connection pooling
func GetMongoClient() (*mongo.Client, error) {
	mongoClientOnce.Do(func() {
		cfg := config.GetConfigs()

		// Configure connection pool
		clientOptions := options.Client().ApplyURI(cfg.MongoHost).
			SetMaxPoolSize(10).     
			SetMinPoolSize(2).
			SetMaxConnIdleTime(30 * time.Minute).
			SetServerSelectionTimeout(5 * time.Second).
			SetConnectTimeout(10 * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			mongoClientErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
			return
		}

		// Ping to verify connection
		err = client.Ping(ctx, nil)
		if err != nil {
			mongoClientErr = fmt.Errorf("failed to ping MongoDB: %w", err)
			return
		}

		mongoClient = client
		log.Println("MongoDB connection pool initialized successfully")
	})

	return mongoClient, mongoClientErr
}


func CreateMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}



func GetKucoinSymbolsFromDB() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := config.GetConfigs()
	client, err := GetMongoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB client: %w", err)
	}

	collection := client.Database(cfg.MarketDatabase).Collection(cfg.ConfigCollection)

	var results []bson.M
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	var symbols []string
	for _, result := range results {
		symbol, ok := result["kucoin_symbol"].(string)
		if ok {
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}