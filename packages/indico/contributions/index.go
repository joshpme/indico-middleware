package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func connect() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("ap-southeast-2"),
		Endpoint:    aws.String("https://syd1.digitaloceanspaces.com"),
		Credentials: credentials.NewStaticCredentials(os.Getenv("SPACES_KEY"), os.Getenv("SPACES_SECRET"), ""),
	})
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

func upload(space *s3.S3, name string, contents string) error {
	_, err := space.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("indico"),
		Key:    aws.String(name),
		Body:   strings.NewReader(contents),
		ACL:    aws.String("private"),
	})
	return err
}

func download(space *s3.S3, name string) (string, error) {
	resp, err := space.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("indico"),
		Key:    aws.String(name),
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getIds(data map[string]interface{}) []float64 {
	var ids []float64
	results := data["results"].([]interface{})
	for _, conference := range results {
		conferenceMap := conference.(map[string]interface{})
		sessions := conferenceMap["sessions"].([]interface{})
		for _, confSession := range sessions {
			sessionMap := confSession.(map[string]interface{})
			contributions := sessionMap["contributions"].([]interface{})
			for _, entry := range contributions {
				entryMap := entry.(map[string]interface{})
				if _, ok := entryMap["code"]; ok {
					if floatVal, ok := entryMap["db_id"].(float64); ok {
						ids = append(ids, floatVal)
					}
				}
			}
		}
	}
	return ids
}

func fetch(event string, id float64) (string, error) {
	url := fmt.Sprintf("https://indico.jacow.org/event/%s/contributions/%.0f.json", event, id)
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

func Main(in Request) (*Response, error) {
	event := "41" // Replace with your event ID

	space, err := connect()
	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Error connecting to S3: %s", err.Error()),
		}, nil
	}

	sessions, err := download(space, "sessions/"+event+".json")
	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Error downloading sessions: %s", err.Error()),
		}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(sessions), &data); err != nil {
		return &Response{
			Body: fmt.Sprintf("Error decoding JSON: %s", err.Error()),
		}, nil
	}

	ids := getIds(data)

	var wg sync.WaitGroup
	maxWorkers := 8
	idsChan := make(chan float64, maxWorkers)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idsChan {
				contribution, err := fetch(event, id)

				if err != nil {
					fmt.Printf("Error fetching contribution: %s\n", err.Error())
					continue
				}

				err = upload(space, fmt.Sprintf("contributions/%s/%.0f.json", event, id), contribution)
				if err != nil {
					fmt.Printf("Error uploading contribution: %s\n", err.Error())
				}
			}
		}()
	}

	for _, id := range ids {
		idsChan <- id
	}

	close(idsChan)

	wg.Wait()

	return &Response{
		Body: fmt.Sprintf("Event ID: %s: %d Contributions downloaded", event, len(ids)),
	}, nil
}
