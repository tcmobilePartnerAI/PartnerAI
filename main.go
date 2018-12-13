package main

import (
  "os"
  "fmt"

  "github.com/nlopes/slack"
  "github.com/christianrondeau/go-wit"
)

// Wit.aiの分析結果の閾値
const confidenceThreshold = 0.5

var (
  slackClient *slack.Client
  witClient   *wit.Client
)

func main() {
  // 環境変数から各サービスに接続するためのトークン情報を取得
  slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
  witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))

  // (1)Slackからのメッセージを受信
  rtm := slackClient.NewRTM() //RTM is "Real Time Messaging"
  go rtm.ManageConnection()

  for msg := range rtm.IncomingEvents {
    // (2)Slackからのメッセージを判定
    if ev, ok := pickUpSlackMessageEvent(msg); ok {
      fmt.Printf("[Slackからのメッセージ内容]\n%v\n\n", ev)

      // (3)Wit.aiへメッセージを転送して、解析結果を受信
      if messageResponse, err := witClient.Message(ev.Msg.Text); err == nil {
        fmt.Printf("[Wit.aiからの返答内容]\n%v\n\n", messageResponse)

        // (4)Slackへ返信するメッセージを作成
        message := createReplyMessge(ev, messageResponse)

        // (5)Slackへ返信
        go replyToSlack(ev, message)
      }
    }
  }
}

// (2)
func pickUpSlackMessageEvent(msg slack.RTMEvent) (*slack.MessageEvent, bool){
  switch ev := msg.Data.(type) {
  case *slack.ConnectedEvent:
    fmt.Printf("Connected to Slack : %v\n", ev.Info)
  case *slack.MessageEvent:
    // SlackのBotが色々メッセージ送ってるのは無視します
    if len(ev.BotID) == 0 {
      // 複数の戻り値を返せます！
      return ev, true
     }
  }

  // 複数の戻り値を返せます！
  return nil, false
}

// (4)
func createReplyMessge(ev *slack.MessageEvent, result *wit.MessageResponse) (text slack.MsgOption) {
  var (
    intent interface{}
  )

  topEntityConfidences :=  make(map[string]float64)
  topEntityValues :=  make(map[string]interface{})

  for key, entityList := range result.Entities {
    topEntityConfidences[key] = 0
    topEntityValues[key] = nil
    for _, entity := range entityList {
      if key == "intent" {
        intent = entity.Value
      }

      if entity.Confidence > confidenceThreshold && entity.Confidence > topEntityConfidences[key] {
        topEntityConfidences[key] = entity.Confidence
        topEntityValues[key] = entity.Value
      }
    }
  }

  text = slack.MsgOptionText("¯\\_(o_o)_/¯", false)

  switch intent {
  case "want eat":
    text = slack.MsgOptionText("私はあんまり。。。", false)
    if v, ok := topEntityValues["food"]; v != nil && ok {
      text = slack.MsgOptionText("私はあんまり" + v.(string) + "好きじゃない。。。", false)
    } else if v, ok := topEntityValues["when"]; v != nil && ok {
      text = slack.MsgOptionText("私、" + v.(string) + "は忙しい。。。", false)
    }
  case "want go":
    text = slack.MsgOptionText("一人で行けば。", false)
    if v, ok := topEntityValues["where"]; v != nil && ok {
      text = slack.MsgOptionText("へぇ、" + v.(string) + "行きたいんだ。一人で行けば。", false)
    } else if v, ok := topEntityValues["when"]; v != nil && ok {
      text = slack.MsgOptionText("私、" + v.(string) + "は忙しい。。。", false)
    }
  // ここに新しいインテントのcase文を追加してください
  }

  // 名前付き戻り値を使うと"return text"と書かなくもて、textを返してくれます
  return
}

// (5)
func replyToSlack(ev *slack.MessageEvent, message slack.MsgOption) {
  params := slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
    AsUser: true,
  })
  slackClient.PostMessage(ev.User, message, params)
}
