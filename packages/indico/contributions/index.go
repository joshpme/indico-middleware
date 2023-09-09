package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type MongoConference struct {
	ID  int       `bson:"_id"`
	End time.Time `bson:"end"`
}

type MongoContribution struct {
	ID int `bson:"_id"`
}

type IndicoCustomField struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

type IndicoAffiliationLink struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	CountryName string `json:"country_name"`
	CountryCode string `json:"country_code"`
	Postcode    string `json:"postcode"`
}

type IndicoDetailedPerson struct {
	ID              int                   `json:"person_id"`
	FirstName       string                `json:"first_name"`
	LastName        string                `json:"last_name"`
	Email           string                `json:"email"`
	IsSpeaker       bool                  `json:"is_speaker"`
	AuthorType      string                `json:"author_type"`
	Affiliation     string                `json:"affiliation"`
	AffiliationLink IndicoAffiliationLink `json:"affiliation_link"`
}

type IndicoType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type IndicoDetailedContribution struct {
	AbstractId   int                    `json:"abstract_id,omitempty"`
	CustomFields []IndicoCustomField    `json:"custom_fields"`
	Persons      []IndicoDetailedPerson `json:"persons"`
	Type         IndicoType             `json:"type"`
}

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

type MongoAffiliationLink struct {
	ID          int    `bson:"id"`
	Name        string `bson:"name"`
	City        string `bson:"city"`
	CountryName string `bson:"country_name"`
	CountryCode string `bson:"country_code"`
	Postcode    string `bson:"postcode"`
}

type MongoPerson struct {
	ID              int                  `bson:"person_id"`
	FirstName       string               `bson:"first_name"`
	LastName        string               `bson:"last_name"`
	Email           string               `bson:"email"`
	IsSpeaker       bool                 `bson:"is_speaker"`
	AuthorType      string               `bson:"author_type"`
	Affiliation     string               `bson:"affiliation"`
	AffiliationLink MongoAffiliationLink `bson:"affiliation_link"`
}

type DetailedMongoContribution struct {
	AbstractID       int           `bson:"abstract_id,omitempty"`
	Persons          []MongoPerson `bson:"persons"`
	IsDuplicate      bool          `bson:"is_duplicate"`
	ContributionType string        `bson:"contribution_type"`
	FundingAgency    string        `bson:"funding_agency"`
	Footnotes        string        `bson:"footnotes"`
}

func fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("INDICO_AUTH"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func currentConferences() ([]int, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))

	client, connectErr := mongo.Connect(context.Background(), clientOptions)
	if connectErr != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", connectErr.Error())
	}

	collection := client.Database("author-title").Collection("conferences")

	// Print the ID of all items in this collection
	cursor, findError := collection.Find(context.Background(), bson.D{})
	if findError != nil {
		return nil, fmt.Errorf("error finding conferences: %s", findError.Error())
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, context.Background())

	var ids []int
	for cursor.Next(context.Background()) {
		var conference MongoConference

		if decodeErr := cursor.Decode(&conference); decodeErr != nil {
			return nil, fmt.Errorf("error decoding conference: %s", decodeErr.Error())
		}

		if conference.End.After(time.Now()) {
			ids = append(ids, conference.ID)
		}
	}

	return ids, nil
}

func getCurrentContributions(collection mongo.Collection, conferenceId int) ([]MongoContribution, error) {
	cursor, findError := collection.Find(context.Background(), bson.D{{"conferenceId", conferenceId}})
	if findError != nil {
		return nil, fmt.Errorf("error finding contributions: %s", findError.Error())
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, context.Background())
	var contributions []MongoContribution
	for cursor.Next(context.Background()) {
		var contribution MongoContribution
		if decodeErr := cursor.Decode(&contribution); decodeErr != nil {
			return nil, fmt.Errorf("error decoding conference: %s", decodeErr.Error())
		}
		contributions = append(contributions, contribution)
	}

	return contributions, nil
}

func indicoPersonToMongoPerson(entry IndicoDetailedPerson) MongoPerson {
	return MongoPerson{
		ID:          entry.ID,
		FirstName:   entry.FirstName,
		LastName:    entry.LastName,
		Email:       entry.Email,
		IsSpeaker:   entry.IsSpeaker,
		AuthorType:  entry.AuthorType,
		Affiliation: entry.Affiliation,
		AffiliationLink: MongoAffiliationLink{
			ID:          entry.AffiliationLink.ID,
			Name:        entry.AffiliationLink.Name,
			City:        entry.AffiliationLink.City,
			CountryName: entry.AffiliationLink.CountryName,
			CountryCode: entry.AffiliationLink.CountryCode,
			Postcode:    entry.AffiliationLink.Postcode,
		},
	}
}

func indicoDetailedContributionToMongoContribution(entry IndicoDetailedContribution) DetailedMongoContribution {
	var persons []MongoPerson
	for _, person := range entry.Persons {
		persons = append(persons, indicoPersonToMongoPerson(person))
	}

	isDuplicate := false
	for _, customField := range entry.CustomFields {
		if customField.Name == "duplicate_of" && customField.Value != "" {
			isDuplicate = true
		}
	}

	fundingAgency := ""
	for _, customField := range entry.CustomFields {
		if customField.Name == "Funding Agency" && customField.Value != "" {
			fundingAgency = customField.Value
		}
	}

	footNotes := ""
	for _, customField := range entry.CustomFields {
		if customField.Name == "Footnotes" && customField.Value != "" {
			footNotes = customField.Value
		}
	}

	return DetailedMongoContribution{
		AbstractID:       entry.AbstractId,
		Persons:          persons,
		IsDuplicate:      isDuplicate,
		ContributionType: entry.Type.Name,
		FundingAgency:    fundingAgency,
		Footnotes:        footNotes,
	}
}

func fetchAndUpdateDetails(conferenceId int, contributionId int, collection mongo.Collection) error {
	detailsContributionContent, err := fetch(fmt.Sprintf("https://indico.jacow.org/event/%d/contributions/%d.json", conferenceId, contributionId))
	if err != nil {
		return err
	}
	var detailedContribution IndicoDetailedContribution
	if err := json.Unmarshal([]byte(detailsContributionContent), &detailedContribution); err != nil {
		return fmt.Errorf("unable to parse json: %s", err.Error())
	}
	_, err = collection.UpdateOne(context.Background(), bson.D{{"_id", contributionId}}, bson.D{
		{"$set", indicoDetailedContributionToMongoContribution(detailedContribution)},
	})
	if err != nil {
		return err
	}

	return nil
}

func fetchConferenceContributions(conferenceId int) error {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))
	client, connectErr := mongo.Connect(context.Background(), clientOptions)
	if connectErr != nil {
		return fmt.Errorf("error connecting to MongoDB: %s", connectErr.Error())
	}
	collection := client.Database("author-title").Collection("contributions")
	contributions, err := getCurrentContributions(*collection, conferenceId)

	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	maxWorkers := 8
	idsChan := make(chan int, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idsChan {
				err := fetchAndUpdateDetails(conferenceId, id, *collection)
				if err != nil {
					fmt.Printf("error fetching contribution details: %s", err.Error())
				}
			}
		}()
	}

	for _, contribution := range contributions {
		idsChan <- contribution.ID
	}
	close(idsChan)

	return nil
}

func Main(in Request) (*Response, error) {

	ids, err := currentConferences()
	if err != nil {
		return nil, fmt.Errorf("error finding current conferences: %s", err.Error())
	}

	for _, id := range ids {
		err := fetchConferenceContributions(id)

		if err != nil {
			return nil, fmt.Errorf("error fetching conference contributions: %s", err.Error())
		}
	}

	return &Response{
		Body: fmt.Sprintf("Updated details for %d conferences", len(ids)),
	}, nil
}
