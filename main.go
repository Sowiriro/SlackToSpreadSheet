package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/sheets/v4"
)

func main() {
	var err error
	spreadsheetId := ""

	//まずは時間を取ってくる
	readRange := "Time"
	timeSheet, err := readSheet(spreadsheetId, readRange)
	if err != nil {
		log.Print("failed read time sheet")
	}

	// log.Print(timeSheet)
	var timeString string
	//timeが取れているのかの確認コード
	if len(timeSheet.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("")
		for _, row := range timeSheet.Values {
			// Print columns A and C, which correspond to indices 0 and ２.
			timeString = row[2].(string)
			fmt.Printf("%s\n", row[2])
		}
	}
	// あとで使えるようにしておきたい
	log.Print(timeString)

	// slackから持ってくる
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("%v", err)
	}
	token := os.Getenv("TOKEN")
	channelID := os.Getenv("CHANNEL_ID")

	callSlack := getCallSlackClient(token, channelID)
	returnHistories, timeStamps, err := callSlack.callGetConversationHistories(GetConversationHistoriesParams{StartTimeStamp: timeString})
	log.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	if err != nil {
		log.Fatalf("%v", err)
	}

	//並列処理したらもう少し早くなるかも
	var ReturnValues []*ReturnValue
	for _, timeStamp := range timeStamps[:len(timeStamps)] {
		returnReply, err := callSlack.callGetConversationReplies(GetConversationRepliesParams{TimeStamp: timeStamp})
		ReturnValues = append(ReturnValues, returnReply...)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	var vr sheets.ValueRange
	// var timeStamp string

	//持ってきたDataをスプレッドシートに記録
	for _, i := range returnHistories {

		userName := getUserName(token, i.User)

		// 変更後のUserの名前
		fmt.Print("\n", userName, i.URL, i.Timestamp)

		myval := []interface{}{userName, i.URL, i.Timestamp}
		vr.Values = append(vr.Values, myval)
	}

	addRange := "Data!A2:C"

	err = addSheet(spreadsheetId, addRange, vr)
	if err != nil {
		log.Fatalf("Unable to retrieve data from")
	}

	// 時間の更新
	t := time.Now()
	unixTime := t.Unix()
	var timeValueRange sheets.ValueRange
	timeValue := []interface{}{"", "", unixTime}
	timeValueRange.Values = append(timeValueRange.Values, timeValue)
	updateRange := "Time!A2:C"
	err = updateSheet(spreadsheetId, updateRange, timeValueRange)
	if err != nil {
		log.Fatalf("Unable to update data from")
	}

}
