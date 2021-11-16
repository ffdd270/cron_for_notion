package main

import (
	"encoding/json"
	"log"
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

	for _, info := range infos {
		parsingYouTrackRequest(&info)
		log.Print(info)
	}

	return infos
}