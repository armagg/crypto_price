package db

import (
	"context"
	"crypto_price/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
  
  func CreatMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
  
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
	  return nil, err
	}
    
	err = client.Ping(ctx, nil)
	if err != nil {
	  return nil, err
	}
  
	return client, nil
  }



func GetKucoinSymbolsFromDB() ([]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
	  data := config.GetConfigs()
    client, err := CreatMongoClient(ctx, data["MONGO_HOST"])
    
    defer client.Disconnect(ctx)

    collection := client.Database(data["MARKET_DATABASE"]).Collection(data["CONFIG_COLLECTION"])
    
    var results []bson.M
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
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