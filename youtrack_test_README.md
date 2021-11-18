```json
[
  {
    "project": "TW",
    "summary": "2021년 %M월 %WM주차 주간 보고서 작성",
    "property": {
      "목표 버전": ["%CURRENT_WEEK_VERSION", "SingleVersionIssueCustomField"],
      "담당자": ["ffdd270", "User"],
      "유형": ["기능", "SingleEnumIssueCustomField"],
      "우선순위": ["READY TO FIRE", "SingleEnumIssueCustomField"]
    },
    "cron": {
      "day_of_week": 5,
      "hour": 0
    }
  }
]
```
## project 
어떤 프로젝트 읽을 건지 적기 

## summary
제목 

## property 
여기서부터는 커스텀 프로퍼티임. YouTrack에 적힌 거 보고 알아서 보기. 0이 value, 1이 type. key는 Youtrack과 동일하게.

## cron
이 일을 얼마동안 반복할 것인가? 해당하는 조건을 구한 다음, 그 조건에 모두 맞는데, 아직 실행이 안 되었을 때, 단 한번만  실행함. 

### day_of_week 
[0-6], 일요일부터 시작. 몇주차?

### hour 
[0-23], 몇 시에?


## 특수 키워드들 

### %M
오늘이 몇월인지 

### %MW 
오늘이 이번달의 몇 주차인지 

### %CURRENT_WEEK_VERSION 
현재 목표버전을 알아서 추츨해서 여기 넣어줌 
