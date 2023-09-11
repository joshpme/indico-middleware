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
	ID        int    `bson:"person_id"`
	FirstName string `bson:"first_name"`
	LastName  string `bson:"last_name"`
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

type GeneratorOrganisation struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Zipcode  string `json:"zipcode"`
}

type GeneratorAuthor struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Affiliations []int  `json:"affiliations"`
}

type GeneratorPayload struct {
	Title         string                        `json:"title"`
	Authors       map[int]GeneratorAuthor       `json:"authors"`
	Organisations map[int]GeneratorOrganisation `json:"organisations"`
}

func getAuthorsAndOrganisations(mongoPersons []MongoPerson, affiliations []MongoAffiliationLink) (map[int]GeneratorAuthor, map[int]GeneratorOrganisation) {

	authors := make(map[int]GeneratorAuthor)
	uniqueOrganisations := make(map[int]GeneratorOrganisation)
	organisationCount := 0
	for _, mongoPerson := range mongoPersons {
		var position int = -1
		for index, organisation := range uniqueOrganisations {
			if organisation.Name == mongoPerson.Affiliation {
				position = index
			}
		}

		if position == -1 {
			uniqueOrganisations[organisationCount] = findAffiliationDetails(mongoPerson.Affiliation, affiliations)
			position = organisationCount
			organisationCount++
		}

		var affiliations []int
		affiliations = append(affiliations, position)

		author := GeneratorAuthor{
			FirstName:    mongoPerson.FirstName,
			LastName:     mongoPerson.FamilyName,
			Affiliations: affiliations,
		}

		if _, ok := authors[mongoPerson.DisplayOrder]; ok {
			for {
				if _, ok := authors[mongoPerson.DisplayOrder+1]; ok {
					mongoPerson.DisplayOrder++
				} else {
					break
				}
			}
			authors[mongoPerson.DisplayOrder] = author
		} else {
			authors[mongoPerson.DisplayOrder] = author
		}
	}

	return authors, uniqueOrganisations
}

func findAllAffiliationLinks(allPersons []MongoDetailedPerson) []MongoAffiliationLink {
	uniqueAffiliations := make(map[int]MongoAffiliationLink)
	for _, person := range allPersons {
		uniqueAffiliations[person.AffiliationLink.ID] = person.AffiliationLink
	}
	var allAffiliations []MongoAffiliationLink
	for _, affiliation := range uniqueAffiliations {
		allAffiliations = append(allAffiliations, affiliation)
	}
	return allAffiliations
}

func findAffiliationDetails(name string, allAffiliations []MongoAffiliationLink) GeneratorOrganisation {
	for _, affiliation := range allAffiliations {
		if affiliation.Name == name {
			return GeneratorOrganisation{
				Name:     affiliation.Name,
				Location: affiliation.City + ", " + affiliation.CountryName,
				Zipcode:  affiliation.Postcode,
			}
		}
	}
	return GeneratorOrganisation{
		Name:     name,
		Location: "",
		Zipcode:  "",
	}
}

func mongoToGeneratorPayload(contribution MongoContribution) GeneratorPayload {
	var mongoPersons []MongoPerson
	mongoPersons = append(mongoPersons, *contribution.Presenters...)
	mongoPersons = append(mongoPersons, *contribution.Authors...)
	for i := 0; i < len(mongoPersons); i++ {
		for j := 0; j < len(mongoPersons)-1; j++ {
			if mongoPersons[j].DisplayOrder > mongoPersons[j+1].DisplayOrder {
				mongoPersons[j], mongoPersons[j+1] = mongoPersons[j+1], mongoPersons[j]
			}
		}
	}
	affiliations := findAllAffiliationLinks(contribution.Persons)
	authors, uniqueOrganisations := getAuthorsAndOrganisations(mongoPersons, affiliations)
	return GeneratorPayload{
		Title:         contribution.Title,
		Authors:       authors,
		Organisations: uniqueOrganisations,
	}
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

	var contributions []MongoContribution = make([]MongoContribution, 0)
	if err := cursor.All(context.Background(), &contributions); err != nil {
		return nil, fmt.Errorf("error decoding documents: %s", err.Error())
	}

	var output []GeneratorPayload = make([]GeneratorPayload, 0)
	for _, contribution := range contributions {
		output = append(output, mongoToGeneratorPayload(contribution))
	}

	// Output as Json
	jsonBytes, err := json.Marshal(output)
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
