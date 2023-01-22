package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const _url = "https://api.ironfish.network/faucet_transactions/status"
const _telegram_api = "https://api.telegram.org/bot"
const _wait_minutes = 3

const _bot_token = "<bot_token>"
const _chat_id = "<chat_id>"

var COMPLETED int
var MSG_ID int64 = 0

func main() {
	fmt.Println("Started. Time: ", time.Now())

	cl := new(http.Client)

	for i := 0; ; i++ {
		HandleCycle(cl, i)

		fmt.Println(i)

		time.Sleep(_wait_minutes * time.Minute)
	}
}

func HandleCycle(cl *http.Client, i int) {
	// first cycle
	if i == 0 {
		COMPLETED = GetResponse(cl, _url)
		SendNewMessage(cl, _bot_token, _chat_id)
		return
	}

	// get response from faucet status
	completedNew := GetResponse(cl, _url)

	var faucet bool
	if completedNew >= COMPLETED+_wait_minutes {
		faucet = true
	} else {
		faucet = false
	}

	UpdateMessage(cl, _bot_token, _chat_id, faucet)

	COMPLETED = completedNew
}

func UpdateMessage(cl *http.Client, bot, chat string, faucet bool) {
	emojiW := "✅"
	emojiNW := "🚫"
	url := _telegram_api
	bot_token := bot
	chat_token := chat
	funcn := "/editMessageText?chat_id="

	textW := fmt.Sprintf("#ironfish\n\nFaucet is working %s", emojiW)
	textNW := fmt.Sprintf("#ironfish\n\nFaucet is not working %s", emojiNW)

	var text string
	if faucet {
		text = textW
	} else {
		text = textNW
	}

	textJson := []byte(fmt.Sprintf(`{
		"text": "%s"
	}`, text))
	jsonReader := bytes.NewBuffer(textJson)

	resp, err := cl.Post(url+bot_token+funcn+chat_token+"&message_id="+
		fmt.Sprint(MSG_ID),
		"application/json", jsonReader)
	if err != nil {
		fmt.Println("Couldn't send message in telegram", err)
	}
	defer resp.Body.Close()

	response, _ := io.ReadAll(resp.Body)
	fmt.Println("telegram response: ", string(response))
}

func GetResponse(cl *http.Client, url string) int {
	var resp *http.Response
	errCount := 0

	for {
		var err error
		resp, err = cl.Get(_url)
		if err != nil {
			fmt.Println(errCount, "@ error getting response from faucet: ", err)
			errCount++
			time.Sleep(3 * time.Second)
			if errCount >= 2 {
				time.Sleep(_wait_minutes)
				errCount = 0
			}
			continue
		}
		break
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	faucetResponse := FaucetResp{}
	err := json.Unmarshal(responseBody, &faucetResponse)
	if err != nil || faucetResponse.Completed == 0 {
		fmt.Println("error parsing from []byte to FaucetResp struct: ", err)
		fmt.Println("trying recursive step")
		resp := GetResponse(cl, url)
		return resp
	}

	fmt.Println(faucetResponse.Completed)

	return faucetResponse.Completed
}

type FaucetResp struct {
	Completed int
	Running   int
	Pending   int
}

func SendNewMessage(cl *http.Client, bot, chat string) {
	faucet := false
	emojiW := "✅"
	emojiNW := "🚫"
	url := _telegram_api
	bot_token := bot
	chat_token := chat
	funcn := "/sendMessage?chat_id="
	textW := fmt.Sprintf("#ironfish\n\nFaucet is working %s", emojiW)
	textNW := fmt.Sprintf("#ironfish\n\nFaucet is not working %s", emojiNW)

	var text string
	if faucet {
		text = textW
	} else {
		text = textNW
	}

	textJson := []byte(fmt.Sprintf(`{
		"text": "%s"
	}`, text))
	jsonReader := bytes.NewBuffer(textJson)

	resp, err := cl.Post(url+bot_token+funcn+chat_token, "application/json", jsonReader)
	if err != nil {
		fmt.Println("Couldn't send message in telegram", err)
	}
	defer resp.Body.Close()

	response, _ := io.ReadAll(resp.Body)
	var msgResponse SendMessageResponse
	json.Unmarshal(response, &msgResponse)
	MSG_ID = msgResponse.Result.Message_id
	fmt.Println("Message Id: ", MSG_ID)
}

type SendMessageResponse struct {
	Ok     bool
	Result struct {
		Message_id int64
	}
}
