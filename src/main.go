package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type NotionRequestParent struct {
	DatabaseId string `json:"database_id"`
}

type NotionRequestPropertiesTitle struct {
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

type NotionRequestPropertiesTitles struct {
	Title [] NotionRequestPropertiesTitle `json:"title"`
}

type NotionRequestProperties  struct {
	Title NotionRequestPropertiesTitles `json:"title"`
}

type NotionRequest struct {
	Parent NotionRequestParent `json:"parent"`
	Properties NotionRequestProperties `json:"properties"`
}

// 위 Type들은
/*{
	"parent": {
		"database_id": "$NOTION_DATABASE_ID"
	},
	"properties": {
		"title": {
			"title": [{
				"text": {
					"content": "Yurts in Big Sur, California"
				}
			}]
		}
	}
}
*/



func createRequest(databaseId string) string {
	request := new(NotionRequest) // key가 string, value는 {}. 뭐던.

	request.Parent.DatabaseId = databaseId
	request.Properties.Title.Title = make([]NotionRequestPropertiesTitle, 1)
	request.Properties.Title.Title[0].Text.Content = "Type을 양산해봤어요"

	doc, _ := json.Marshal(request)

	return string(doc)
}

func main() {
	cntxt := &daemon.Context{
		PidFileName: "sample.pid",
		PidFilePerm: 0644,
		LogFileName: "sample.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        os.Args,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	startNotionCron()
}

func startNotionCron() {
	for {
		Nsecs := rand.Intn(3000)
		time.Sleep(time.Millisecond * time.Duration(Nsecs))
		t := time.Now()

		if t.Second() == 00 {
			log.Print("EXECUTED!!!!")
			sendNotionAPI()
		}
	}
}

func sendNotionAPI() {

	if len(os.Args) != 3 {
		fmt.Print("USAGE : API_KEY\n")
		return
	}

	notionApiKey := os.Args[1]
	notionDatabaseId := os.Args[2]

	fmt.Print(notionApiKey + "\n" + notionDatabaseId + "\n")

	requestStr := createRequest(notionDatabaseId)

	requestClient := http.Client{}
	req, _ := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBufferString(requestStr) )
	req.Header.Add("Authorization", "Bearer " + notionApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Notion-Version", "2021-08-16")

	resp, _ := requestClient.Do(req)

	defer resp.Body.Close() // 함수가 끝날 떄 불림. 우와. 짱 신기하다.
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(data))
}