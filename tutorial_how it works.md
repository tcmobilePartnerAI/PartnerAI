# AI彼女 go側の仕組み

## ざっくりいうと
```
func main() {
	slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))

	// (1)Slackからのメッセージを受信
	fmt.Printf("%s\n", "Connecting to Slack...")
	rtm := slackClient.NewRTM() //RTM is "Real Time Messaging"
	go rtm.ManageConnection()
	fmt.Printf("%s\n", "... Connected.")

	for msg := range rtm.IncomingEvents {
    	// (2)Slackからのメッセージを判定
    	ev := pickUpSlackMessageEvent(msg)
    	if ev != nil {
	    	fmt.Printf("ev : %s\n\n", ev)
      		// (3)Wit.aiへメッセージを転送して、解析結果を受信
      		messageResponse := sendAndReceiveWithWit(ev)
      		if messageResponse != nil {
	    		fmt.Printf("messageResponse : %s\n\n", messageResponse)
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
slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))

// (1)Slackからのメッセージを受信
fmt.Printf("%s\n", "Connecting to Slack...")
rtm := slackClient.NewRTM() //RTM is "Real Time Messaging"
go rtm.ManageConnection()
fmt.Printf("%s\n", "... Connected.")

for msg := range rtm.IncomingEvents {
	中略
}
```
この部分です。  
BOTにメッセージを送信したらfor文の中の処理が行われます

### (2)受け取ったらどんなメッセージか判定して
```
// (2)
func pickUpSlackMessageEvent(msg slack.RTMEvent) *slack.MessageEvent{
	switch ev := msg.Data.(type) {
	case *slack.MessageEvent:
	  fmt.Printf("ev.BotID : %s\n\n", ev.BotID)
    if len(ev.BotID) == 0 {
      return ev
     }
    return nil
  default:
    return nil
	}
}
```
この部分です。  
slackから受信するイベントにはいろいろな種類がありますが、  
`MessagkeEvent`というのがメッセージを受信した時のイベントタイプとなります。  
ここではそれ以外のタイプだった場合は何もしないように制御しています

### (3)wit.aiに解析を依頼して

```
// (3)
func sendAndReceiveWithWit(ev *slack.MessageEvent) *wit.MessageResponse{
	result, err := witClient.Message(ev.Msg.Text)
	if err != nil {
		return nil
	}
  return result
}
```
この部分です  
wit.aiにリクエストを投げているだけです。
もしエラーがあった場合は何も返さないようにしています。

### (4)結果からbotにしゃべらせるメッセージを作成して

```
// (4)
func createReplyMessge(ev *slack.MessageEvent, result *wit.MessageResponse) slack.MsgOption {
	var (
    intent          interface{}
    food            interface{}
    where           interface{}
	)

	for key, entityList := range result.Entities {
		for _, entity := range entityList {
      if key == "intent" {
        intent = entity.Value
      }
      if key == "food" {
        food = entity.Value
      }
      if key == "where" {
        where = entity.Value
      }
		}
  }

	text := slack.MsgOptionText("¯\\_(o_o)_/¯", false)

	switch intent {
	case "want eat":
		text = slack.MsgOptionText("私はあんまり" + food.(string) + "すきじゃない。。。", false)
	case "want go":
		text = slack.MsgOptionText("へぇ、" + where.(string) + "行きたいんだ。一人で行けば。", false)
	}

  return text
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
	text = slack.MsgOptionText("私はあんまり" + food.(string) + "すきじゃない。。。", false)
case "want go":
	text = slack.MsgOptionText("へぇ、" + where.(string) + "行きたいんだ。一人で行けば。", false)
}
```
この部分を
```
switch intent {
case "want eat":
	text = slack.MsgOptionText("私はあんまり" + food.(string) + "すきじゃない。。。", false)
case "want go":
	text = slack.MsgOptionText("へぇ、" + where.(string) + "行きたいんだ。一人で行けば。", false)
case "invite go":
	text = slack.MsgOptionText("いいね、" + where.(string) + "連れてって！", false)
}
```

としてください  
それでslackで  
「明日一緒に遊園地行こうよ」  
と言ったら  
「いいね、遊園地連れてって！」 
って言ってくれるかと思います
