package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func init() {
	// Load the environment variables from the .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}
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

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

type MongoConference struct {
	ID  int       `bson:"_id"`
	End time.Time `bson:"end"`
}

type MongoContribution struct {
	ID           int            `bson:"_id"`
	Code         string         `bson:"code"`
	Title        string         `bson:"title"`
	Description  string         `bson:"description"`
	Presenters   *[]MongoPerson `bson:"presenters,omitempty"`
	Authors      *[]MongoPerson `bson:"authors,omitempty"`
	ConferenceId int            `bson:"conferenceId"`
}

type MongoPerson struct {
	FirstName    string `bson:"firstName"`
	FamilyName   string `bson:"familyName"`
	Affiliation  string `bson:"affiliation"`
	DisplayOrder int    `bson:"displayOrder"`
	Email        string `bson:"email"`
}

type Timetable struct {
	Results map[string]map[string]map[string]TimetableSession `json:"results"`
}

type TimetableAuthor struct {
	FirstName    string        `json:"firstName"`
	FamilyName   string        `json:"familyName"`
	Affiliation  string        `json:"affiliation"`
	DisplayOrder []interface{} `json:"displayOrderKey"`
	Email        string        `json:"email"`
}

type TimetableEntry struct {
	ID          int                `json:"contributionId"`
	Code        string             `json:"code"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Presenters  *[]TimetableAuthor `json:"presenters,omitempty"`
	Authors     *[]TimetableAuthor `json:"authors,omitempty"`
}

type TimetableDate struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Tz   string `json:"tz"`
}

type TimetableSession struct {
	ID        string                    `json:"id"`
	Title     string                    `json:"title"`
	Code      string                    `json:"code"`
	StartDate TimetableDate             `json:"startDate"`
	EndDate   TimetableDate             `json:"endDate"`
	Entries   map[string]TimetableEntry `json:"entries"`
}

func timetableAuthorToMongoPerson(author TimetableAuthor) MongoPerson {
	displayOrder := 0
	if len(author.DisplayOrder) > 0 {
		if displayOrderInt, ok := author.DisplayOrder[0].(float64); ok {
			displayOrder = int(displayOrderInt)
		}
	}
	return MongoPerson{
		FirstName:    author.FirstName,
		FamilyName:   author.FamilyName,
		Affiliation:  author.Affiliation,
		DisplayOrder: displayOrder,
		Email:        author.Email,
	}
}

func timetableEntryToMongoContribution(entry TimetableEntry, conferenceId int) MongoContribution {
	var presenters *[]MongoPerson
	if entry.Presenters != nil {
		presenters = &[]MongoPerson{}
		for _, presenter := range *entry.Presenters {
			*presenters = append(*presenters, timetableAuthorToMongoPerson(presenter))
		}
	}
	var authors *[]MongoPerson
	if entry.Authors != nil {
		authors = &[]MongoPerson{}
		for _, author := range *entry.Authors {
			*authors = append(*authors, timetableAuthorToMongoPerson(author))
		}
	}

	if entry.ID == 0 {
		fmt.Printf("Conference: %d \n Entry %+v\n", conferenceId, entry)
	}

	return MongoContribution{
		ID:           entry.ID,
		Code:         entry.Code,
		Title:        entry.Title,
		Description:  entry.Description,
		Presenters:   presenters,
		Authors:      authors,
		ConferenceId: conferenceId,
	}
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
	// Iterate through the cursor to access the retrieved documents
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

func findSessions(timetableContent string) (map[int]TimetableEntry, error) {
	var timetable Timetable
	if err := json.Unmarshal([]byte(timetableContent), &timetable); err != nil {
		return nil, fmt.Errorf("unable to parse json: %s", err.Error())
	}
	entries := make(map[int]TimetableEntry)
	for _, day := range timetable.Results {
		for _, room := range day {
			for _, session := range room {
				for _, entry := range session.Entries {
					if entry.ID != 0 {
						entries[entry.ID] = entry
					}
				}
			}
		}
	}
	return entries, nil
}

func uploadTimetable(id int, collection mongo.Collection, wg *sync.WaitGroup) error {
	defer wg.Done()

	cursor, findError := collection.Find(context.Background(), bson.D{{"conferenceId", id}})

	if findError != nil {
		return fmt.Errorf("error finding contributions: %s", findError.Error())
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, context.Background())

	// Store a list of existing contributions
	existingContributions := make(map[int]MongoContribution)
	for cursor.Next(context.Background()) {
		var contribution MongoContribution
		if decodeErr := cursor.Decode(&contribution); decodeErr != nil {
			return fmt.Errorf("error decoding conference: %s", decodeErr.Error())
		}
		existingContributions[contribution.ID] = contribution
	}

	timetableContent, err := fetch(fmt.Sprintf("https://indico.jacow.org/export/timetable/%d.json", id))
	if err != nil {
		return fmt.Errorf("error fetching timetable: %s", err.Error())
	}

	entries, err := findSessions(timetableContent)

	if err != nil {
		return fmt.Errorf("error finding sessions: %s", err.Error())
	}

	var operations []mongo.WriteModel

	if len(entries) == 0 && len(existingContributions) == 0 {
		return nil
	}

	for _, entry := range entries {
		mongoContribution := timetableEntryToMongoContribution(entry, id)
		if _, found := existingContributions[mongoContribution.ID]; !found {
			operation := mongo.NewInsertOneModel().SetDocument(mongoContribution)
			operations = append(operations, operation)
		} else {
			filter := bson.D{{"_id", entry.ID}}
			update := bson.D{
				{"$set", timetableEntryToMongoContribution(entry, id)},
			}
			operation := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
			operations = append(operations, operation)
		}
	}

	for _, contribution := range existingContributions {
		if _, found := entries[contribution.ID]; !found {
			filter := bson.D{{"_id", contribution.ID}}
			operation := mongo.NewDeleteOneModel().SetFilter(filter)
			operations = append(operations, operation)
		}
	}

	bulkWriteOptions := options.BulkWrite().SetOrdered(false)
	_, err = collection.BulkWrite(context.Background(), operations, bulkWriteOptions)

	if err != nil {
		return fmt.Errorf("error bulk writing: %s", err.Error())
	}
	return nil
}

func Main(in Request) (*Response, error) {

	ids, err := currentConferences()
	if err != nil {
		return nil, fmt.Errorf("error finding current conferences: %s", err.Error())
	}

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_AUTH"))

	client, connectErr := mongo.Connect(context.Background(), clientOptions)
	if connectErr != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", connectErr.Error())
	}

	collection := client.Database("author-title").Collection("contributions")

	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		id := id
		go func() {
			err = uploadTimetable(id, *collection, &wg)
			if err != nil {
				fmt.Printf("Error uploading timetable: %s", err.Error())
			}
		}()
	}

	wg.Wait()

	return &Response{
		Body: fmt.Sprintf("%d Timetables downloaded", len(ids)),
	}, nil
}