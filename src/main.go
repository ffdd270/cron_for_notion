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

	Type string `json:"type"`
}

type NotionRequestPropertiesTitleData struct {
	Title [] NotionRequestPropertiesTitle `json:"title"`
	Type string `json:"type"`
}

type NotionRequestPropertiesDateData struct {
	Date struct {
		Start string `json:"start"`
		End   *string `json:"end"`
	} `json:"date"`

	Type string `json:"type"`
}

type NotionRequestProperties  struct {
	Title NotionRequestPropertiesTitleData `json:"title"`
	Date NotionRequestPropertiesDateData `json:"date"`
}


type NotionRequestTextProperty struct {

}


type NotionRequestHeadingOne struct {
	Object string `json:"object"`
	Type string `json:"type"`
	Heading1 string `json:"heading_1"`
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
	makeNotionDocumentByJsonData()
	// startDaemon();
}

// JSONRequestPropertyDate GoLang에서는 대문자로 스타트 시켜야 한다 =ㅁ =
type JSONRequestPropertyDate struct {
	Start string
	End *string // nullable
}

type JSONRequestProperties struct {
	Title string
	Date JSONRequestPropertyDate
}

type JSONRequest struct {
	Properties JSONRequestProperties
	Text string
}


func createRequestFromJsonRequest( databaseId string, jsonReq JSONRequest ) string {
	req := new(NotionRequest)

	req.Parent.DatabaseId = databaseId
	req.Properties.Title.Type = "title"
	req.Properties.Title.Title = make([]NotionRequestPropertiesTitle, 1)
	req.Properties.Title.Title[0].Text.Content = jsonReq.Properties.Title
	req.Properties.Title.Title[0].Type = "text"

	req.Properties.Date.Type = "date"
	req.Properties.Date.Date.Start = jsonReq.Properties.Date.Start
	req.Properties.Date.Date.End = jsonReq.Properties.Date.End

	doc, _ := json.Marshal(req)

	return string(doc)
}

func makeNotionDocumentByJsonData() {
	if len(os.Args) != 3 {
		fmt.Print("USAGE : API_KEY\n")
		return
	}

	notionApiKey := os.Args[1]
	notionDatabaseId := os.Args[2]


	data, err := os.Open( "test.json" )

	if err != nil{
		log.Print(err)
	}

	byteValue, _ := ioutil.ReadAll(data)

	var infos []JSONRequest
	json.Unmarshal( byteValue, &infos )

	for _, info := range infos {
		reqStr := createRequestFromJsonRequest( notionDatabaseId, info  )
		log.Print(reqStr)

		sendNotionRequest( notionApiKey, notionDatabaseId, reqStr )
	}

}


func startDaemon() {
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
	requestStr := createRequest(notionDatabaseId)

	sendNotionRequest( notionApiKey, notionDatabaseId, requestStr )
}

func sendNotionRequest(notionApiKey string, notionDatabaseId string, requestStr string) {
	fmt.Print(notionApiKey + "\n" + notionDatabaseId + "\n")


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