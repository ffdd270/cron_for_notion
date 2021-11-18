package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type JSONYoutrackRequestCron struct {
	DayOfWeek int `json:"day_of_week"`
	Hour int `json:"hour"`
}

type JSONYoutrackRequestProperty  map[string]string // String - String

type JSONYoutrackRequest struct {
	Project string
	Summary string
	Property JSONYoutrackRequestProperty
	Cron JSONYoutrackRequestCron
}

type Parser interface {
	valid(testStr string) bool
	replace(testStr string) string
}


type ParserMonth struct {
	ReplaceMonth string
}

func (p ParserMonth) valid(testStr string) bool {
	return strings.Contains( testStr, "%M" )
}


func (p ParserMonth) replace(testStr string) string {
	return strings.Replace( testStr, "%M", p.ReplaceMonth, -1 )
}


// %MW하면 버그라서(.. )
type ParserWeekOfMonth struct {
	ReplaceWeekMonth string
}
func (p ParserWeekOfMonth) valid(testStr string) bool {
	return strings.Contains( testStr, "%WM" )
}


func (p ParserWeekOfMonth) replace(testStr string) string {
	return strings.Replace( testStr, "%WM", p.ReplaceWeekMonth, -1 )
}


func parsingYouTrackRequest( request * JSONYoutrackRequest ) {
	parsers := make([]Parser, 2)

	parserMonth := new(ParserMonth)
	parserMonth.ReplaceMonth = "11"
	parsers[0] = parserMonth

	parserWeekOfMonth := new(ParserWeekOfMonth)
	parserWeekOfMonth.ReplaceWeekMonth = "3"
	parsers[1] = parserWeekOfMonth

	// 파싱 대상. Porperty 와 summary.


	for _, parser := range parsers {
		if parser.valid(request.Summary) {
			request.Summary = parser.replace(request.Summary)
			log.Print("CHK ")
		}
	}


	// Property
	for key, value := range request.Property {
		destStr := value

		for _, parser := range parsers {
			if parser.valid(destStr) {
				destStr = parser.replace(destStr)
			}
		}

		request.Property[key] = destStr
	}
}

func getYouTrackRequests() []JSONYoutrackRequest {
	byteValue, _ := getByteValue("youtrack_test.json")
	var infos []JSONYoutrackRequest
	json.Unmarshal( byteValue, &infos )

	for key, info := range infos {
		parsingYouTrackRequest(&info)
		infos[key] = info
	}

	return infos
}


type YoutrackSprintResponse struct {
	Start int `json:"start"`
	Finish int `json:"finish"`
	Name string `json:"name"`
	Id string `json:"id"`
	Type string `json:"$type"`
}

type YoutrackAgileResponse struct {
	CurrentSprint YoutrackSprintResponse
	Type string  `json:"type"`
}

type YoutrackAgileResponses []YoutrackAgileResponse


type YoutrackVersionValue struct {
	ReleaseDate int `json:"releaseDate"`
	Name string `json:"name"`
	Id string `json:"id"`
}

type YoutrackVersionResponse struct {
	Values []YoutrackVersionValue `json:"values"`
	Name string `json:"name"`
	Id string `json:"id"`
}

type YoutrackVersionResponsesValue []YoutrackVersionResponse

func requestCurrentYoutrackVersion( youtrackApiKey string ) string {
	ret := requestGETToYoutrack( "agiles?fields=currentSprint(start,finish,name,id)", youtrackApiKey )

	var infos YoutrackAgileResponses
	json.Unmarshal( ret, &infos )
	currentSprintName := infos[0].CurrentSprint.Name

	ret = requestGETToYoutrack("admin/customFieldSettings/bundles/version?fields=id,name,fieldType(presentation,id),values(id,name,releaseDate)", youtrackApiKey)
	var versionInfos YoutrackVersionResponsesValue
	json.Unmarshal(ret, &versionInfos)

	versionInfoValues := versionInfos[0].Values

	currentVersionId := ""

	for _, versionInfo := range versionInfoValues {
		if versionInfo.Name == currentSprintName {
			currentVersionId = versionInfo.Id
			break
		}
	}

	return currentVersionId
}

type YoutrackProjectResponse struct {
	ShortName string `json:"shortName"`
	Name string `json:"name"`
	Id string `json:"id"`
}

type YoutrackProjectResponses []YoutrackProjectResponse

func requestShortNameToYoutrackId( youtrackApiKey string, projectName string ) string {
	ret := requestGETToYoutrack("admin/projects?fields=id,name,shortName", youtrackApiKey)
	var projectInfos YoutrackProjectResponses
	json.Unmarshal(ret, &projectInfos)

	for _, projectInfo := range projectInfos {
		if projectInfo.ShortName == projectName {
			return projectInfo.Id
		}
	}

	return ""
}

func requestCreateAnIssue( youtrackApiKey string, projectId string, targetVersionId string, summary string ) {
	print("youtrackApiKey?", youtrackApiKey, "\tprojectId", projectId, "\ttargetVersionId", targetVersionId, "\tsummary", summary)
}


func requestGETToYoutrack(youtrackApi string, youtrackApiKey string) []byte {
	requestClient := http.Client{}
	req, _ := http.NewRequest("GET", "https://kuroneko-lab.myjetbrains.com/youtrack/api/" + youtrackApi, nil )
	req.Header.Add("Authorization", "Bearer " + youtrackApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	resp, _ := requestClient.Do(req)

	defer resp.Body.Close() // 함수가 끝날 떄 불림. 우와. 짱 신기하다.

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return data
}


func requestPOSTToYoutrack(youtrackApi string, youtrackApiKey string, requestStr string) []byte {
	requestClient := http.Client{}
	req, _ := http.NewRequest("POST", "https://kuroneko-lab.myjetbrains.com/youtrack/api/" + youtrackApi, bytes.NewBufferString(requestStr) )
	req.Header.Add("Authorization", "Bearer " + youtrackApiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	resp, _ := requestClient.Do(req)

	defer resp.Body.Close() // 함수가 끝날 떄 불림. 우와. 짱 신기하다.

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return data
}