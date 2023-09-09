package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strconv"
)

type MongoContribution struct {
	ID               int                   `bson:"_id"`
	Code             string                `bson:"code"`
	Title            string                `bson:"title"`
	Description      string                `bson:"description"`
	Presenters       *[]MongoPerson        `bson:"presenters,omitempty"`
	Authors          *[]MongoPerson        `bson:"authors,omitempty"`
	ConferenceId     int                   `bson:"conferenceId"`
	AbstractID       int                   `bson:"abstract_id,omitempty"`
	Persons          []MongoDetailedPerson `bson:"persons"`
	IsDuplicate      bool                  `bson:"is_duplicate"`
	ContributionType string                `bson:"contribution_type"`
}

type MongoPerson struct {
	FirstName    string `bson:"firstName"`
	FamilyName   string `bson:"familyName"`
	Affiliation  string `bson:"affiliation"`
	DisplayOrder int    `bson:"displayOrder"`
	//Email        string `bson:"email"` PII
}

type MongoAffiliationLink struct {
	ID          int    `bson:"id"`
	Name        string `bson:"name"`
	City        string `bson:"city"`
	CountryName string `bson:"country_name"`
	CountryCode string `bson:"country_code"`
	Postcode    string `bson:"postcode"`
}

type MongoDetailedPerson struct {
	ID              int                  `bson:"person_id"`
	FirstName       string               `bson:"first_name"`
	LastName        string               `bson:"last_name"`
	// Email           string               `bson:"email"` PII
	IsSpeaker       bool                 `bson:"is_speaker"`
	AuthorType      string               `bson:"author_type"`
	Affiliation     string               `bson:"affiliation"`
	AffiliationLink MongoAffiliationLink `bson:"affiliation_link"`
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

	// Convert in.Conference to int
	conferenceId, err := strconv.Atoi(in.Conference)
	if err != nil {
		return nil, fmt.Errorf("error converting conference id to int: %s", err.Error())
	}

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))

	client, connectErr := mongo.Connect(context.Background(), clientOptions)
	if connectErr != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", connectErr.Error())
	}

	collection := client.Database("author-title").Collection("contributions")

	cursor, findError := collection.Find(context.Background(), bson.D{
		{"conferenceId", conferenceId},
		{"code", in.Code},
	})

	if findError != nil {
		return nil, fmt.Errorf("error finding documents: %s", findError.Error())
	}

	var contributions []MongoContribution
	if err := cursor.All(context.Background(), &contributions); err != nil {
		return nil, fmt.Errorf("error decoding documents: %s", err.Error())
	}

	// Output as Json
	jsonBytes, err := json.Marshal(contributions)
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
