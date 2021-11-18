package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type JSONYoutrackRequestCron struct {
	DayOfWeek int `json:"day_of_week"`
	Hour int `json:"hour"`
}

type JSONYoutrackRequestProperty  map[string][]string // String - String

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

type ParserCurrentWeekVersion struct {
	ReplaceValue string
}
func (p ParserCurrentWeekVersion) valid(testStr string) bool {
	return strings.Contains( testStr, "%CURRENT_WEEK_VERSION" )
}

func (p ParserCurrentWeekVersion) replace(testStr string) string {
	return strings.Replace( testStr, "%CURRENT_WEEK_VERSION", p.ReplaceValue, -1 )
}

func parsingYouTrackRequest( request * JSONYoutrackRequest, targetVersionId string ) {
	parsers := make([]Parser, 3)

	parserMonth := new(ParserMonth)
	parserMonth.ReplaceMonth = "11"
	parsers[0] = parserMonth

	parserWeekOfMonth := new(ParserWeekOfMonth)
	parserWeekOfMonth.ReplaceWeekMonth = "3"
	parsers[1] = parserWeekOfMonth

	parserCurrentWeekVersion := new(ParserCurrentWeekVersion)
	parserCurrentWeekVersion.ReplaceValue = targetVersionId
	parsers[2] = parserCurrentWeekVersion


	// 파싱 대상. Porperty 와 summary.

	request.Project = requestShortNameToYoutrackId( os.Args[2], request.Project )

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
			if parser.valid(destStr[0]) {
				destStr[0] = parser.replace(destStr[0])
			}
		}

		request.Property[key] = destStr
	}
}

func getYouTrackRequests(targetVersionId string) []JSONYoutrackRequest {
	byteValue, _ := getByteValue("youtrack_test.json")
	var infos []JSONYoutrackRequest
	json.Unmarshal( byteValue, &infos )

	for key, info := range infos {
		parsingYouTrackRequest(&info, targetVersionId)
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

type YoutrackCustomFieldVersionBundle struct {
	Id string `json:"id"`
	Type  string `json:"Type"`
}

func makeVersionBundle( versionId string ) YoutrackCustomFieldVersionBundle {
	return YoutrackCustomFieldVersionBundle{ versionId, "VersionBundleElement" }
}

type YoutrackCustomFieldRequest struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
	Type string `json:"$type"`
}

func makeSingleIssueCustomField( name string, versionId string ) YoutrackCustomFieldRequest {
	value := makeVersionBundle(versionId)

	return YoutrackCustomFieldRequest{ name, value, "SingleVersionIssueCustomField" }

}

type YoutrackIssueRequest struct {
	Project struct{ Id string `json:"id"` } `json:"project"`
	Summary string `json:"summary"`
	Description string `json:"description"`
	CustomFields []interface{} `json:"customFields"`
}



func requestCreateAnIssue( youtrackApiKey string, jsonYoutrackRequest JSONYoutrackRequest ) {
	var request YoutrackIssueRequest

	request.Project.Id = jsonYoutrackRequest.Project
	request.Summary = jsonYoutrackRequest.Summary
	request.Description = ""

	for propertyKey, property := range jsonYoutrackRequest.Property {
		valueString := property[0]
		typeString := property[1]

		if typeString == "SingleVersionIssueCustomField" {
			request.CustomFields = append( request.CustomFields, makeSingleIssueCustomField( propertyKey, valueString  ) )
		}
	}


	doc, _ := json.Marshal(request)
	result := string(doc)

	response := requestPOSTToYoutrack( "issues", youtrackApiKey, result )

	print(result)
	print(string(response))

	//print("youtrackApiKey?", youtrackApiKey, "\tprojectId", projectId, "\ttargetVersionId", targetVersionId, "\tsummary", summary)
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