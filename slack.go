package main

import (
	"fmt"
	"regexp"

	"github.com/slack-go/slack"
)

type CallSlack struct {
	Api       *slack.Client
	ChannelID string
}

type ReturnValue struct {
	User      string
	Timestamp string
	URL       string
}

type GetConversationHistoriesParams struct {
	ChannelID      string
	HistoryCursol  string
	StartTimeStamp string
}

type GetConversationRepliesParams struct {
	ChannelID   string
	ReplyCursol string
	TimeStamp   string
}

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("%v", err)
// 	}
// 	token := os.Getenv("TOKEN")
// 	channelID := os.Getenv("CHANNEL_ID")

// 	callSlack := getCallSlackClient(token, channelID)
// 	returnHistories, timeStamps, err := callSlack.callGetConversationHistories(GetConversationHistoriesParams{})
// 	if err != nil {
// 		log.Fatalf("%v", err)
// 	}

// 	//並列処理したらもう少し早くなるかも
// 	var ReturnValues []*ReturnValue
// 	for _, timeStamp := range timeStamps[:len(timeStamps)] {
// 		returnReply, err := callSlack.callGetConversationReplies(GetConversationRepliesParams{TimeStamp: timeStamp})
// 		ReturnValues = append(ReturnValues, returnReply...)
// 		if err != nil {
// 			log.Fatalf("%v", err)
// 		}
// 	}

// 	for _, v := range returnHistories {
// 		fmt.Printf("returnHistory:%v\n", v)
// 	}
// 	for _, v := range ReturnValues {
// 		fmt.Printf("returnValue:%v\n", v)
// 	}
// }

func getCallSlackClient(token string, channelID string) *CallSlack {
	api := slack.New(token)
	return &CallSlack{Api: api, ChannelID: channelID}
}

func (callSlack *CallSlack) callGetConversationHistories(params GetConversationHistoriesParams) ([]*ReturnValue, []string, error) {
	var returnValues []*ReturnValue
	var timeStamps []string

	historyParams := slack.GetConversationHistoryParameters{ChannelID: callSlack.ChannelID, Cursor: params.HistoryCursol, Oldest: params.StartTimeStamp}
	resp, err := callSlack.Api.GetConversationHistory(&historyParams)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(resp.Messages); i++ {
		// replyがある場合はcallGetConversationRepliesでurlの収集を行う
		if resp.Messages[i].ReplyCount > 0 {
			timeStamps = append(timeStamps, resp.Messages[i].Timestamp)
		} else {
			urls := parseURL(resp.Messages[i].Text)
			if len(urls) == 0 {
				continue
			}
			for _, url := range urls {
				returnValues = append(returnValues, &ReturnValue{User: resp.Messages[i].User, Timestamp: resp.Messages[i].Timestamp, URL: url[0]})
			}
		}
	}
	if resp.HasMore {
		nextCursor := resp.ResponseMetaData.NextCursor
		callSlack.callGetConversationHistories(GetConversationHistoriesParams{ChannelID: callSlack.ChannelID, HistoryCursol: nextCursor})
	}
	return returnValues, timeStamps, nil
}

func (callSlack *CallSlack) callGetConversationReplies(params GetConversationRepliesParams) ([]*ReturnValue, error) {
	var returnValues []*ReturnValue
	replyParams := slack.GetConversationRepliesParameters{ChannelID: callSlack.ChannelID, Timestamp: params.TimeStamp, Cursor: params.ReplyCursol}
	msg, hasMore, nextCursor, err := callSlack.Api.GetConversationReplies(&replyParams)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(msg); i++ {
		urls := parseURL(msg[i].Text)
		if len(urls) == 0 {
			continue
		}
		for _, url := range urls {
			returnValues = append(returnValues, &ReturnValue{User: msg[i].User, Timestamp: msg[i].Timestamp, URL: url[0]})
		}
	}
	if hasMore {
		callSlack.callGetConversationReplies(GetConversationRepliesParams{ChannelID: callSlack.ChannelID, ReplyCursol: nextCursor, TimeStamp: params.TimeStamp})
	}
	return returnValues, nil
}

//URLの抽出
func parseURL(text string) [][]string {
	r := regexp.MustCompile(`https?://[\w/:%#\$&\?\(\)~\.=\+\-]+`)
	res := r.FindAllStringSubmatch(text, -1)
	return res
}

// User Name の抽出
func getUserName(token string, userId string) string {
	api := slack.New(token)
	user, err := api.GetUserInfo(userId)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return user.Name
}
