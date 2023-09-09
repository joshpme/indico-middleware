package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

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

type Conference struct {
	id       int
	name     string
	start    time.Time
	end      time.Time
	location string
	category string
}

func parseDate(dateStr map[string]interface{}) (time.Time, error) {
	date, ok := dateStr["date"].(string)
	if !ok {
		return time.Time{}, errors.New("date is not a string")
	}
	return time.Parse("2006-01-02", date)
}

func getConferences(data map[string]interface{}) ([]Conference, error) {
	var conferences []Conference
	results := data["results"].([]interface{})
	for _, conference := range results {

		conferenceMap := conference.(map[string]interface{})

		visibility := conferenceMap["visibility"].(map[string]interface{})

		if visibility["name"] == "Nowhere" {
			continue
		}

		conferenceIdStr, ok := conferenceMap["id"].(string)
		if !ok {
			return nil, errors.New("conference code is not an int")
		}
		conferenceId, err := strconv.Atoi(conferenceIdStr)
		if err != nil {
			return nil, err
		}

		start, err := parseDate(conferenceMap["startDate"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		end, err := parseDate(conferenceMap["endDate"].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		newConference := Conference{
			id:       conferenceId,
			name:     conferenceMap["title"].(string),
			start:    start,
			end:      end,
			location: conferenceMap["location"].(string),
			category: conferenceMap["category"].(string),
		}
		conferences = append(conferences, newConference)
	}
	return conferences, nil
}

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(in Request) (*Response, error) {
	payload, err := fetch(fmt.Sprintf("https://indico.jacow.org/export/categ/%d.json", 2))
	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Error downloading sessions: %s", err.Error()),
		}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return &Response{
			Body: fmt.Sprintf("Error decoding JSON: %s", err.Error()),
		}, nil
	}

	conferences, err := getConferences(data)

	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Failed to parse events payload: %s", err.Error()),
		}, nil
	}

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Error connecting to MongoDB: %s", err.Error()),
		}, nil
	}

	collection := client.Database("author-title").Collection("conferences")

	for _, conference := range conferences {
		filter := bson.D{{"_id", conference.id}}
		update := bson.D{
			{"$set", bson.D{
				{"name", conference.name},
				{"start", conference.start},
				{"end", conference.end},
				{"location", conference.location},
				{"category", conference.category},
			}},
		}

		_, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return &Response{
				Body: fmt.Sprintf("Error performing upsert: %s", err.Error()),
			}, nil
		}
	}

	return &Response{
		Body: fmt.Sprintf("%d Contributions downloaded", len(conferences)),
	}, nil
}
