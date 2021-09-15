package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

	//HTTP

	/*
	resp, err := http.Get("https://google.com")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close() // 함수가 끝날 떄 불림. 우와. 짱 신기하다.

	// 결과 출력
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s\n", string(data))
	*/

}
