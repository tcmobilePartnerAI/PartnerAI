package main

import (
	"log"
	"os"
	"fmt"

	"github.com/christianrondeau/go-wit"
	"github.com/nlopes/slack"
)

const confidenceThreshold = 0.5

var (
	slackClient   *slack.Client
	witClient     *wit.Client
)

func main() {
	slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))

	rtm := slackClient.NewRTM()
	go rtm.ManageConnection()

	fmt.Printf("%s\n", "Start...")
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.BotID) == 0 {
				go handleMessage(ev)
			}
		}
	}
}

func handleMessage(ev *slack.MessageEvent) {
	fmt.Printf("%v\n", ev)

	result, err := witClient.Message(ev.Msg.Text)
	if err != nil {
		log.Printf("unable to get wit.ai result: %v", err)
		return
	}

	var (
		topEntity    wit.MessageEntity
		topEntityKey string
	)

	for key, entityList := range result.Entities {
		for _, entity := range entityList {
			if entity.Confidence > confidenceThreshold && entity.Confidence > topEntity.Confidence {
				topEntity = entity
				topEntityKey = key
			}
		}
	}

	replyToUser(ev, topEntity, topEntityKey)
}

func replyToUser(ev *slack.MessageEvent, topEntity wit.MessageEntity, topEntityKey string) {
	params := slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
		AsUser: true,
	})
	text := slack.MsgOptionText("¯\\_(o_o)_/¯", false)

	switch topEntityKey {
	case "greetings":
		text = slack.MsgOptionText("Hello user! " + ev.Msg.Text, false)
	case "name":
		text = slack.MsgOptionText("Hello user! Who is " + ev.Msg.Text + "?", false)
	}

	slackClient.PostMessage(ev.User, text, params)
}
