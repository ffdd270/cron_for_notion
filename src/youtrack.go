package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	valid(testStr string, typeStr string) bool
	replace(testStr string, additionalInfo string) string
}


type ParserMonth struct {
	ReplaceMonth string
}

func (p ParserMonth) valid(testStr string, _ string) bool {
	return strings.Contains( testStr, "%M" )
}


func (p ParserMonth) replace(testStr string, additionalInfo string) string {
	return strings.Replace( testStr, "%M", p.ReplaceMonth, -1 )
}


// %MW하면 버그라서(.. )
type ParserWeekOfMonth struct {
	ReplaceWeekMonth string
}
func (p ParserWeekOfMonth) valid(testStr string, _ string) bool {
	return strings.Contains( testStr, "%WM" )
}


func (p ParserWeekOfMonth) replace(testStr string, _ string) string {
	return strings.Replace( testStr, "%WM", p.ReplaceWeekMonth, -1 )
}

type ParserCurrentWeekVersion struct {
	ReplaceValue string
}
func (p ParserCurrentWeekVersion) valid(testStr string, chk string) bool {
	return strings.Contains( testStr, "%CURRENT_WEEK_VERSION" ) && strings.Contains( chk,  "SingleVersionIssueCustomField" )
}

func (p ParserCurrentWeekVersion) replace(testStr string, _ string) string {
	return strings.Replace( testStr, "%CURRENT_WEEK_VERSION", p.ReplaceValue, -1 )
}

type ParserReplaceUserName struct {
	isRequestYouTrack bool
	youtrackId string

	YoutrackApiKey string
}


func parserUserNameReplaceCreate( youtrackApiKey string ) * ParserReplaceUserName  {
	ptr := new(ParserReplaceUserName)
	ptr.isRequestYouTrack = false
	ptr.youtrackId = ""
	ptr.YoutrackApiKey = youtrackApiKey

	return ptr
}

func (p ParserReplaceUserName) valid(_ string, typeStr string) bool {
	return strings.Contains(typeStr, "SingleUserIssueCustomField")
}

func (p * ParserReplaceUserName) replace(testStr string, _ string) string {
	if !p.isRequestYouTrack {
		result := requestCurrentUserId( p.YoutrackApiKey, testStr )
		p.youtrackId = result
		p.isRequestYouTrack = true
	}

	return p.youtrackId
}


type ParserReplaceEnumIssue struct {
	isRequestYouTrack bool
	youtrackId string
	InstanceList []YoutrackCustomFieldInstanceValue
	YoutrackApiKey string
}

func parserEnumIssueCreate( youtrackApiKey string) * ParserReplaceEnumIssue  {
	ptr := new(ParserReplaceEnumIssue)
	ptr.isRequestYouTrack = false
	ptr.youtrackId = ""
	ptr.YoutrackApiKey = youtrackApiKey

	return ptr
}

func (p ParserReplaceEnumIssue) valid(_ string, typeStr string) bool {
	return strings.Contains( typeStr, "SingleEnumIssueCustomField" )
}

func (p * ParserReplaceEnumIssue) replace(testStr string, propertyKey string) string {
	print("MUSH ROOM HUNTER ", propertyKey, "\n")
	if !p.isRequestYouTrack {
		result := getEnumId( p.YoutrackApiKey, p.InstanceList, testStr )
		p.youtrackId = result
		p.isRequestYouTrack = true
	}

	return p.youtrackId
}

type YoutrackEnumBundleResponse struct {
	Values []YoutrackCustomField `json:"values"`
	Name string `json:"name"`
	Id string `json:"id"`
}

type YoutrackEnumBundleResponses []YoutrackEnumBundleResponse

func getEnumId( youtrackApiKey string, enumList []YoutrackCustomFieldInstanceValue, findString string ) string {
	res := requestGETToYoutrack( "admin/customFieldSettings/bundles/enum?fields=name,id,values(name,id,localizedName)", youtrackApiKey )
	var responseObject YoutrackEnumBundleResponses
	json.Unmarshal( res, &responseObject )

	enumId := enumList[0].Bundle.Id
	println("enumId?", enumId)
	for _, enumBundle := range responseObject {
		if enumBundle.Id != enumId {
			continue
		}
		for _, enumValue := range enumBundle.Values {
			print("enumValue ? ", enumValue.Name, "enumValue? ", enumValue.Id)

			if ( enumValue.LocalizedName != nil && *enumValue.LocalizedName == findString )  || enumValue.Name == findString  {
				return enumValue.Id
			}
		}
	}

	println("NOT FIND ENUM : ", enumId, findString)
	return ""
}

func parsingYouTrackRequest( request * JSONYoutrackRequest, youtrackApiKey string, targetVersionId string ) {
	parsers := make([]Parser, 5)

	nowDate := time.Now()
	nowDay := nowDate.Day()
	firstDate := time.Date( nowDate.Year(), nowDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	firstDateWeekday := int(firstDate.Weekday())
	firstWeekCount := 7 - firstDateWeekday // firstDateWeekday => 0-6, 첫 주차는 얼마짜리?
	weekOfMonth := 1

	if firstWeekCount < nowDay {
		calcRemainValue := nowDay - firstDateWeekday
		weekOfMonth += int(math.Floor(float64(calcRemainValue / 7)))
	}

	parserMonth := new(ParserMonth)
	parserMonth.ReplaceMonth = strconv.Itoa(int(nowDate.Month()))
	parsers[0] = parserMonth

	parserWeekOfMonth := new(ParserWeekOfMonth)
	parserWeekOfMonth.ReplaceWeekMonth = strconv.Itoa(weekOfMonth)
	parsers[1] = parserWeekOfMonth

	parserCurrentWeekVersion := new(ParserCurrentWeekVersion)
	parserCurrentWeekVersion.ReplaceValue = targetVersionId
	parsers[2] = parserCurrentWeekVersion

	parserReplaceUserName := parserUserNameReplaceCreate( youtrackApiKey )
	parsers[3] = parserReplaceUserName

	nameMap := getNameToCustomFieldMap( youtrackApiKey )
	parserReplaceEnumIssue := parserEnumIssueCreate( youtrackApiKey )
	parsers[4] = parserReplaceEnumIssue

	// 파싱 대상. Porperty 와 summary.

	request.Project = requestShortNameToYoutrackId( os.Args[2], request.Project )

	for _, parser := range parsers {
		if parser.valid(request.Summary, "") {
			request.Summary = parser.replace(request.Summary, "")
		}
	}

	// Property
	for key, value := range request.Property {
		destStr := value

		for _, parser := range parsers {
			if parser.valid(destStr[0], destStr[1]) {
				if destStr[1] == "SingleEnumIssueCustomField" {
					parser.(*ParserReplaceEnumIssue).InstanceList = nameMap[key].Instances
				}

				destStr[0] = parser.replace(destStr[0], key)
			}
		}

		request.Property[key] = destStr
	}
}

func getYouTrackRequests(youtrackApiKey string, targetVersionId string) []JSONYoutrackRequest {
	byteValue, _ := getByteValue("youtrack_test.json")
	var infos []JSONYoutrackRequest
	json.Unmarshal( byteValue, &infos )

	for key, info := range infos {
		parsingYouTrackRequest(&info, youtrackApiKey, targetVersionId)
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

type YoutrackUserResponse struct {
	Login string `json:"login"`
	Id string `json:"id"`
	Type string `json:"$type"`
}

type YoutrackUserResponses []YoutrackUserResponse


func requestCurrentUserId( youtrackApiKey string, loginId string ) string {
	ret := requestGETToYoutrack( "users?fields=login,id", youtrackApiKey )
	var users YoutrackUserResponses
	json.Unmarshal( ret, &users )
	println(string(ret))

	for _, user := range users {
		if user.Login == loginId {
			return user.Id
		}
	}
	return ""
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

type YoutrackCustomFieldValue struct {
	Id * string `json:"id"`
	Name *string `json:"name"`
	Type  string `json:"$type"`
}

func makeVersionBundle( versionId string ) YoutrackCustomFieldValue {
	return YoutrackCustomFieldValue{&versionId, nil,"VersionBundleElement" }
}

func makeUserBundle( userId string  ) YoutrackCustomFieldValue {
	return YoutrackCustomFieldValue{ &userId, nil,"User" }
}

func makeSingleEnumBundle( id string ) YoutrackCustomFieldValue {
	return YoutrackCustomFieldValue{ &id, nil, "EnumBundleElement" }
}

type YoutrackCustomFieldRequest struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
	Type string `json:"$type"`
}

func makeSingleIssueCustomFieldVersion( name string, versionId string ) YoutrackCustomFieldRequest {
	value := makeVersionBundle(versionId)

	return YoutrackCustomFieldRequest{ name, value, "SingleVersionIssueCustomField" }
}

func makeSingleIssueCustomFieldUser( name string, userId string ) YoutrackCustomFieldRequest {
	value := makeUserBundle(userId)
	return YoutrackCustomFieldRequest{ name, value, "SingleUserIssueCustomField" }
}

func makeSingleEnumField( id string, valueName string ) YoutrackCustomFieldRequest {
	value := makeSingleEnumBundle( valueName )
	return YoutrackCustomFieldRequest{  id, value, "SingleEnumIssueCustomField" }
}

type YoutrackIssueRequest struct {
	Project struct{ Id string `json:"id"` } `json:"project"`
	Summary string `json:"summary"`
	Description string `json:"description"`
	CustomFields []interface{} `json:"customFields"`
}

type YoutrackCustomFieldInstanceValue struct{
	Bundle struct {
		Name string `json:"Name"`
		Id string `json:"id"`
	} `json:"bundle"`

	Id string
}

type YoutrackCustomField struct {
	Id string `json:"id"`
	Name string `json:"name"`
	LocalizedName *string `json:"localizedName"`
	Instances []YoutrackCustomFieldInstanceValue
}
type YoutrackCustomFields []YoutrackCustomField

func requestCustomFields( youtrackApiKey string ) YoutrackCustomFields {
	var customFields YoutrackCustomFields
	customFieldsRes := requestGETToYoutrack( "admin/customFieldSettings/customFields?fields=id,name,localizedName,instances(id,bundle(id,name))", youtrackApiKey)
	json.Unmarshal( customFieldsRes, &customFields )
	return customFields
}

func getNameToCustomFieldMap( youtrackApiKey string ) map[string]YoutrackCustomField {
	rtn := make(map[string]YoutrackCustomField)
	fields := requestCustomFields( youtrackApiKey )

	for _, field := range fields {
		if field.LocalizedName != nil {
			rtn[ *field.LocalizedName ] = field
		}

		rtn[ field.Name ] = field
 	}
	return rtn
}



func requestCreateAnIssue( youtrackApiKey string, jsonYoutrackRequest JSONYoutrackRequest ) {
	nameMap := getNameToCustomFieldMap( youtrackApiKey )

	var request YoutrackIssueRequest

	request.Project.Id = jsonYoutrackRequest.Project
	request.Summary = jsonYoutrackRequest.Summary
	request.Description = ""

	for propertyKey, property := range jsonYoutrackRequest.Property {
		valueString := property[0]
		typeString := property[1]

		customField, ok := nameMap[propertyKey]
		if ok == false {
			continue
		}

		if typeString == "SingleVersionIssueCustomField" {
			request.CustomFields = append( request.CustomFields, makeSingleIssueCustomFieldVersion( customField.Name, valueString  ) )
		} else if typeString == "SingleUserIssueCustomField" {
			request.CustomFields = append( request.CustomFields, makeSingleIssueCustomFieldUser( customField.Name, valueString  ) )
		} else if typeString == "SingleEnumIssueCustomField" {
			request.CustomFields = append( request.CustomFields, makeSingleEnumField( customField.Name, valueString ) )
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