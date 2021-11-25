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


type NotionRequestPropertyLink struct {
	Url string `json:"url"`
	Type string `json:"type"`
}

func newNotionRequestPropertyLink(urlText string) *NotionRequestPropertyLink {
	link := new(NotionRequestPropertyLink)
	link.Type = "url"
	link.Url = urlText

	return link
}

type NotionRequestPropertyText struct {
	Text struct {
		Content string `json:"content"`
		Link *NotionRequestPropertyLink `json:"link"`
	} `json:"text"`

	Type string `json:"type"`
}

func newNotionRequestPropertyText( text string, link *string ) NotionRequestPropertyText {
	textProperty := NotionRequestPropertyText{ Type: "text" }
	textProperty.Text.Content = text

	if link != nil {
		textProperty.Text.Link = newNotionRequestPropertyLink( *link )
	}

	return textProperty
}


type NotionRequestPropertiesTitleData struct {
	Title [] NotionRequestPropertyText `json:"title"`
	Type string `json:"type"`
}

type NotionRequestPropertyOption struct {
	Name string `json:"name"`
}

type NotionRequestPropertiesOptionData struct {
	Select NotionRequestPropertyOption `json:"select"`
}

type NotionRequestPropertiesDateData struct {
	Date struct {
		Start string `json:"start"`
		End   *string `json:"end"`
	} `json:"date"`

	Type string `json:"type"`
}

type NotionRequestBlockObject struct {
	Object string `json:"object"`
	Type string `json:"type"`
}

func newNotionRequestBlockObject(typeString string) NotionRequestBlockObject {
	block := NotionRequestBlockObject{ Type: typeString, Object: "block"}
	return block
}

type NotionRequestTextData struct {
	Text []NotionRequestPropertyText `json:"text"`
}

type NotionRequestHeadingOneBlock struct {
	NotionRequestBlockObject
	Heading1 NotionRequestTextData `json:"heading_1"`
}

func newNotionRequestHeadingOneBlock( headingText string ) NotionRequestHeadingOneBlock {
	blockObject := newNotionRequestBlockObject("heading_1")
	headingOne := NotionRequestHeadingOneBlock{ }
	headingOne.Object = blockObject.Object
	headingOne.Type = blockObject.Type

	headingOne.Heading1.Text = make([]NotionRequestPropertyText, 1)
	headingOne.Heading1.Text[0] = newNotionRequestPropertyText(headingText, nil)
	return headingOne
}

type NotionRequestParagraphBlock struct {
	NotionRequestBlockObject
	Paragraph NotionRequestTextData `json:"paragraph"`
}

func newNotionRequestParagraphBlock( text string ) NotionRequestParagraphBlock {
	blockObject := newNotionRequestBlockObject("paragraph")
	paragraph := NotionRequestParagraphBlock{}
	paragraph.Object = blockObject.Object
	paragraph.Type = blockObject.Type

	paragraph.Paragraph.Text =  make([]NotionRequestPropertyText, 1)
	paragraph.Paragraph.Text[0] = newNotionRequestPropertyText(text, nil)
	return paragraph
}

func newNotionRequestParagraphBlockWithLink( text string, link * string ) NotionRequestParagraphBlock {
	paragraph := newNotionRequestParagraphBlock(text)
	paragraph.Paragraph.Text[0] = newNotionRequestPropertyText( text, link )
	return paragraph
}


func (block * NotionRequestParagraphBlock) addText(text string, link * string) {
	textProperty := newNotionRequestPropertyText(text, link)
	block.Paragraph.Text = append(block.Paragraph.Text, textProperty)
}

type NotionProperties = map[string]interface{}

type NotionRequest struct {
	Parent NotionRequestParent `json:"parent"`
	Properties NotionProperties `json:"properties"`
	Children []interface{} `json:"children"`
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
	titleProperty := NotionRequestPropertiesTitleData{}
	titleProperty.Title = make([]NotionRequestPropertyText, 1)
	titleProperty.Title[0] = newNotionRequestPropertyText("Type을 양산해봤어요", nil)

	request.Properties["title"] = titleProperty

	doc, _ := json.Marshal(request)

	return string(doc)
}

func main() {
	if len(os.Args) != 4 && (os.Args[1] == "create" || os.Args[1] == "retrieve") {
		fmt.Print("USAGE : command API_KEY DATABASE_ID\n")
		return
	}

	if os.Args[1] == "youtrack_cron" {
		currentVersionId := requestCurrentYoutrackVersion( os.Args[2] )
		youtrackRequest := getYouTrackRequests(os.Args[2], currentVersionId)

		requestCreateAnIssue( os.Args[2], youtrackRequest[0] )
	}

	if os.Args[1] == "create" {
		makeNotionDocumentByJsonData()
	}

	if os.Args[1] == "retrieve" {
		notionApiKey := os.Args[2]
		notionDatabaseId := os.Args[3]

		sendNotionRetrieveRequest( notionApiKey, notionDatabaseId )
	}

	// startDaemon();
}

// JSONRequestPropertyDate GoLang에서는 대문자로 스타트 시켜야 한다 =ㅁ =
type JSONRequestType struct {
	Type string
}

type JSONRequestTextType struct {
	JSONRequestType
	Text string
}

type JsonRequestTextListDataType struct {
	Text string
	Link *string
}

type JsonRequestTextListType struct {
	JSONRequestType
	Text []JsonRequestTextListDataType
}
type JSONRequestPropertyDate struct {
	JSONRequestType
	Start string
	End *string // nullable
}

type JSONRequestProperties struct {
	Title string
	Date JSONRequestPropertyDate
}

type JSONTemplate struct {
	Properties map[string]map[string]interface{}
	Text []map[string]interface{}
}

type JSONData struct {
	Title string
	Text *struct {
		Idx int
		Text string
		Link *string
	}
}

func createRequestFromJsonRequest( databaseId string, jsonTemplate JSONTemplate, jsonData JSONData) string {
	req := new(NotionRequest)

	req.Parent.DatabaseId = databaseId
	req.Properties = NotionProperties{}

	for key, info := range jsonTemplate.Properties {
		typeString := info["type"].(string)
		jsonRequestType := JSONRequestType{typeString}

		if typeString == "title" {
			text := jsonData.Title
			titleInfo := JSONRequestTextType{ jsonRequestType, text }

			titleProperty := NotionRequestPropertiesTitleData{}
			titleProperty.Type = "title"
			titleProperty.Title = make([]NotionRequestPropertyText, 1)
			titleProperty.Title[0] = newNotionRequestPropertyText(titleInfo.Text, nil)

			req.Properties[key] = titleProperty
		} else if typeString == "date"  {
			start := info["start"].(string)
			var end *string = nil

			if info["end"] != nil {
				end = info["end"].(*string)
			}

			dateInfo := JSONRequestPropertyDate{ jsonRequestType, start, end }

			dateProperty := NotionRequestPropertiesDateData{}
			dateProperty.Type = "date"
			dateProperty.Date.Start = dateInfo.Start
			dateProperty.Date.End = dateInfo.End

			req.Properties[key] = dateProperty
		} else if typeString == "select" {
			name := info["name"].(string)

			optionProperty := NotionRequestPropertiesOptionData{}
			optionProperty.Select.Name = name

			req.Properties[key] = optionProperty

		}
	}

	childrenSize := len(jsonTemplate.Text)
	req.Children = make([]interface{}, childrenSize)

	for index, info := range jsonTemplate.Text {
		typeString := info["type"].(string)
		jsonRequestType := JSONRequestType{typeString}
		var result interface{}

		if jsonRequestType.Type == "h1" {
			text := info["text"].(string)
			result = newNotionRequestHeadingOneBlock( text )
		}else if jsonRequestType.Type == "text" {
			log.Print(info)

			text := info["text"].([]interface{})
			textFirstElement := text[0].(map[string]interface{})



			firstTextString := textFirstElement["text"].(string)
			var firstLinkString *string = nil

			if  textFirstElement["link"] != nil {
				str := textFirstElement["link"].(string)
				firstLinkString = &str
			}

			if (jsonData.Text != nil) && (jsonData.Text.Idx == index) {
				firstTextString = jsonData.Text.Text
				firstLinkString = jsonData.Text.Link
			}

			paragraphBlock := newNotionRequestParagraphBlockWithLink( firstTextString, firstLinkString )

			for index, textInfo := range text {
				if index == 0 {
					continue
				}

				textElement := textInfo.(map[string]interface{})

				textString := textElement["text"].(string)
				var linkString *string = nil

				if  textElement["link"] != nil {
					str := textElement["link"].(string)
					linkString = &str
				}

				paragraphBlock.addText( textString, linkString )
			}

			result = paragraphBlock
		}

		req.Children[index] = result
	}


	doc, _ := json.Marshal(req)

	return string(doc)
}

func getByteValue(fileName string) ([]byte, error) {
	data, err := os.Open( fileName )

	if err != nil{
		log.Print(err)
	}

	byteValue,err := ioutil.ReadAll(data)

	return byteValue, err
}

func makeNotionDocumentByJsonData() {
	if len(os.Args) != 4 {
		fmt.Print("USAGE : API_KEY\n")
		return
	}

	notionApiKey := os.Args[2]
	notionDatabaseId := os.Args[3]

	byteValue, _ := getByteValue("json_template.json")
	var template JSONTemplate
	json.Unmarshal( byteValue, &template )

	byteValue, _ = getByteValue("test.json")
	var infos []JSONData
	json.Unmarshal( byteValue, &infos )

	for _, info := range infos {
		reqStr := createRequestFromJsonRequest( notionDatabaseId, template, info )
		log.Print(reqStr)

		sendNotionPageCreateRequest( notionApiKey, notionDatabaseId, reqStr )
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

	sendNotionPageCreateRequest( notionApiKey, notionDatabaseId, requestStr )
}

func sendNotionPageCreateRequest(notionApiKey string, notionDatabaseId string, requestStr string) {
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

func sendNotionRetrieveRequest(notionApiKey string, notionDatabaseId string) {
	fmt.Print(notionApiKey + "\n" + notionDatabaseId + "\n")

	requestClient := http.Client{}
	req, _ := http.NewRequest("GET", "https://api.notion.com/v1/databases/" + notionDatabaseId, nil )
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