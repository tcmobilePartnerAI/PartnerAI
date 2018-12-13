# AI彼女 go側の仕組み

## ざっくりいうと

```
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
```

- (1)slackからイベントを受信できるようにして
- (2)受け取ったらどんなメッセージか判定して
- (3)wit.aiに解析を依頼して
- (4)結果からbotにしゃべらせるメッセージを作成して
- (5)botにしゃべらせる

↑がgo側の実装となってます  
各処理のざっくりとした概要は下記

### (1)slackからイベントを受信できるようにして

```
  // 環境変数から各サービスに接続するためのトークン情報を取得
  slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
  witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))

  // (1)Slackからのメッセージを受信
  rtm := slackClient.NewRTM() //RTM is "Real Time Messaging"
  go rtm.ManageConnection()

  for msg := range rtm.IncomingEvents {
	  中略
  }
```

この部分です。  
BOTにメッセージを送信したらfor文の中の処理が行われます

### (2)受け取ったらどんなメッセージか判定して

```
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
```

この部分です。  
slackから受信するイベントにはいろいろな種類がありますが、  
`MessagkeEvent`というのがメッセージを受信した時のイベントタイプとなります。  
ここではそれ以外のタイプだった場合は何もしないように制御しています

### (3)wit.aiに解析を依頼して

```
      // (3)Wit.aiへメッセージを転送して、解析結果を受信
      if messageResponse, err := witClient.Message(ev.Msg.Text); err == nil {
```

この部分です  
wit.aiにリクエストを投げているだけです。
もしエラーがあった場合は何も返さないようにしています。

### (4)結果からbotにしゃべらせるメッセージを作成して

```
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
	  中略
  // ここに新しいインテントのcase文を追加してください 
  }

  // 名前付き戻り値を使うと"return text"と書かなくもて、textを返してくれます
  return
}
```

この部分です  
witは解析結果として連想配列のようなものを返します  
わかりづらいと思うので具体例を挙げて説明します。  
例えば、wit.aiに「明日海に行きたい」と送ったら
```
key:"intent"
value:"want go"

key:"where"
value:"海"
```
みたいな情報が解析結果としてが返ってきます  
上記を各対応した`interface{}`に詰めてるのが↑のfor文の処理です  
for文の後で一旦デフォルトのメッセージを設定しておいて  
intentの種類で返答のメッセージを出し分けています  

### (5)botにしゃべらせる

```
// (5)
func replyToSlack(ev *slack.MessageEvent, message slack.MsgOption) {
  params := slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{
    AsUser: true,
  })
  slackClient.PostMessage(ev.User, message, params)
}
```

です  
引き数の`message`に(4)で設定したメッセージが格納されています。
ここではメッセージを送信してきたユーザー(`ev.User`)にメッセージを返信しています。  
また`AsUser: true`で作成したbotユーザーとしての変身をしています

## 実装例

例えば  
「明日一緒に遊園地行こうよ」  
と言ったら  
「いいね、遊園地連れてって！」  
と言ってもらいたいとします  

まずはwit.aiのintentに
`invite go`なんて感じのintentを登録します。
それでwit.ai側で上記の文章を送信したときに
```
key:"intent"
value:"invite go"

key:"where"
value:"遊園地"
```
を返してくれるようにうまいこと学習させてください  
witの学習方法は別途冨永君あたりが説明してくれるかと思います。  

で
```
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
```
この部分の
`  // ここに新しいインテントのcase文を追加してください`
に
```
  case "invite go":
    if v, ok := topEntityValues["where"]; v != nil && ok {
      text = slack.MsgOptionText("いいね、" + v.(string) + "連れてって！", false)
    }
```
を追加してください。

それでslackで  
「明日一緒に遊園地行こうよ」  
と言ったら  
「いいね、遊園地連れてって！」 
って言ってくれるかと思います
