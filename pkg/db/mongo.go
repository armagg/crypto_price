package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
  
  func CreatMongoClient(ctx context.Context, uri string, dbName string) (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(uri)
  
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
	  return nil, err
	}
  
	db := client.Database(dbName)
  
	err = client.Ping(ctx, nil)
	if err != nil {
	  return nil, err
	}
  
	return db, nil
  }