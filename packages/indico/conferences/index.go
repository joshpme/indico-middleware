package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type MongoConference struct {
	ID   int    `bson:"_id"`
	Name string `bson:"name"`
}
type Request struct {
	Conference string `json:"conference"`
	Code       string `json:"code"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(in Request) (*Response, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))
	client, connectErr := mongo.Connect(context.Background(), clientOptions)
	if connectErr != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", connectErr.Error())
	}
	collection := client.Database("author-title").Collection("conferences")
	cursor, findError := collection.Find(context.Background(), bson.D{})
	if findError != nil {
		return nil, fmt.Errorf("error finding conferences: %s", findError.Error())
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, context.Background())
	var conferences []MongoConference
	if err := cursor.All(context.Background(), &conferences); err != nil {
		return nil, fmt.Errorf("error decoding documents: %s", err.Error())
	}
	jsonBytes, err := json.Marshal(conferences)
	if err != nil {
		return nil, fmt.Errorf("error marshalling documents: %s", err.Error())
	}
	return &Response{
		Body: string(jsonBytes),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
