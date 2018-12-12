# 目的

- AIを利用した彼女botを作成する。

# 概要

- AI  
Wit.ai
- bot  
Slack
- 準備  
  - Wit.aiでいい感じ学習させたプロジェクトを用意しておく
  - Slack Appでbotを用意しておく
  - Go言語の開発環境を用意しておく
- 仕組み  
  1. Slackでbotに話しかけると
  1. Wit.aiにWeb APIを利用して話しかけられたメッセージ内容を送って
  1. それに対するレスポンスを受け取ったら、Slackに返答のメッセージをbotとして投稿する  
ということをする処理をGoで実現する。

# 準備
## Wit.ai
- URL  
https://wit.ai/
- 手順
  1. アカウント作る
  1. アプリ作る
  1. 学習... がんばる
  1. アプリの画面で、"Settings"メニュークリック
  1. "Server Access Token"をコピーしてどこかに覚えておく

## Slack

- URL  
https://slack.com/
- 手順
  1. アカウント作る
  1. Workspace作る
  1. Apps作る
    1. 下記URLへアクセス
    https://ワークスペース名.slack.com/apps/manage
    1. 上部の"Search App Directory"に`bots`と入力して検索
    1. "Bots"を選択
    1. "Add Configuration"をクリック
    1. BotのUsernameに任意の名前を入力
    1. "Add bot integration"をクリック
    1. Botの設定画面が表示されるので"API Token"をコピーしてどこかに覚えておく

## Git

- 今回使用するGoのパッケージを取得するために必要
- 手順
  1. 下記からDL  
  https://git-scm.com/download/win
  1. DLしたインストーラー(exe)を実行
  1. インストーラーの仰せのままにインストール

## Go  
### Windows10想定
- 参考  
https://qiita.com/yoskeoka/items/0dcc62a07bf5eb48dc4b

- 手順
  1. コマンド プロンプト を管理者として起動
  1. Chocolateyインストール　　
```
@"%SystemRoot%\System32\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -InputFormat None -ExecutionPolicy Bypass -Command "iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))" && SET "PATH=%PATH%;%ALLUSERSPROFILE%\chocolatey\bin"
```
※参考：1分くらいかかった。
  1. Goインストール
```
choco install golang
```
※参考：所要時間2,3分程度だが、ネットからDLを行っているので、当日やってもらうのは微妙。
  1. PATHを通す
    1. Winキーを押下
    2. `env`と入力
    3. "最も一致する検索結果"に表示された、"環境変数を編集"をクリック
    4. ユーザーの環境変数の一覧から"Path"を選択状態にして、"編集"をクリック  
    ※この時点ですでに`%GOPATH%\bin`が設定済みなら、後の作業不要。
    5. "新規"をクリック
    6. `%GOPATH%\bin`を入力
    7. "OK"をクリックして、環境変数名の編集ウィンドウを閉じる
    8. "OK"をクリックして、環境変数ウィンドウを閉じる
  1. コマンドプロンプトを起動  
  ※管理者権限不要
  1. goコマンドのバージョンを確認する
```
go version
```
下記のように表示されたらOK  
※バージョンはその時点の最新となります。
```
go version go1.10.2 windows/amd64
```
  1. 今回必要なパッケージを導入
```
go get -u github.com/nlopes/slack
go get -u github.com/christianrondeau/go-wit
```

### mac想定
- 参考  
https://qiita.com/balius_1064/items/ac7dff5ef10eaf69996f
https://qiita.com/Noah0x00/items/63e024f9b5a27276401b

- 手順
  1. ターミナル を起動
  1. homebrewインストール　
```
https://qiita.com/balius_1064/items/ac7dff5ef10eaf69996f
```
  1. goインストール　
```
https://qiita.com/Noah0x00/items/63e024f9b5a27276401b
```

    1. goコマンドのバージョンを確認する
```
go version
```
下記のように表示されたらOK  
※バージョンはその時点の最新となります。
```
go version go1.10.2 windows/amd64
```
    2. 今回必要なパッケージを導入
```
go get -u github.com/nlopes/slack
go get -u github.com/christianrondeau/go-wit
```
---

# 実行

- 実際の勉強会の手順ではなく、このRepositoryのソースを動かす手順を記します。

- 手順
  1. 作業用のフォルダ作成  
  例)C:\Users\xxx\work
  1. Git Bash起動
  1. Cloneする
```
cd
cd work
git clone https://github.com/tcmobilePartnerAI/PartnerAI.git
```
  開発用は以下になります。（今回は使用しません）
  git clone git@gitlab.com:n.matsushige/talk-with-wit.ai.git


  1. Cloneできてるか確認
```
ls -lR talk_with_witai
```
ソースが取得できてればOK
  1. コマンドプロンプト起動
  1. 環境変数設定
* windowsの場合
```
set PATH=%PATH%C:\Program Files\Git\bin;
set SLACK_ACCESS_TOKEN=SSSSS
set WIT_AI_ACCESS_TOKEN=WWWWW
```
* macの場合(動作確認はしていないので、もし違うようなら修正お願いします。)
```
export SLACK_ACCESS_TOKEN="SSSSS"
export WIT_AI_ACCESS_TOKEN="WWWWW"
```

SSSSS : Slackのbotの"API Token"
WWWWW : Wit.aiの"Server Access Token"
  1. 実行
```
go run main.go
```
※処理始まるまで時間かかるので、10秒心の中で数えてください。
  1. Slackでbotに話しかけてください。  
  なにか返答があるはず。。。
  1. 実際には、Wit.aiで学習時に設定したIntentを、Goの処理中でWit.aiのAPIからの戻り値を判定する処理に入れ込まないといけないです。


# 参考
## Wit.ai
https://wit.ai
## Slack Go package
https://github.com/nlopes/slack
## packagemain
- youtube  
https://www.youtube.com/watch?v=zkB_c3cgtd0
- Code  
https://github.com/plutov/packagemain/tree/master/09-slack-bot
