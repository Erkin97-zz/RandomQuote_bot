/*
simple webhook server for my bot
this spaghetti code just for testing golang power
*/

//TODO Refactor code for error handling and make readable structure

package main

import (
	"fmt"
	"bufio"
//	"io/ioutil"
	"html"
	"log"
	"path/filepath"
	"github.com/grokify/html-strip-tags-go"
	"net/http"
	"os"
	"encoding/json"
	"github.com/line/line-bot-sdk-go/linebot"
	"strings"
)

func doQuote() string {
	resp, err := http.Get("http://quotesondesign.com/wp-json/posts?filter[orderby]=rand&filter[posts_per_page]=1")
	if err != nil {
		log.Fatalln(err)
	}
	var quoteData []interface{}
	err = json.NewDecoder(resp.Body).Decode(&quoteData)
	if err != nil {
		log.Fatalln(err)
		return "Sorry no quotes :|"
	}
	fmt.Printf("Your quote data %v", quoteData)
	quoteContent := quoteData[0].(map[string]interface{})["content"].(string)
	quoteContent = strip.StripTags(quoteContent)
	quoteContent = html.UnescapeString(quoteContent)
	quoteContent = strings.TrimSuffix(quoteContent, "\n")
	quoteContent = strings.TrimSuffix(quoteContent, " ")
	quoteAuthor := quoteData[0].(map[string]interface{})["title"].(string)
	message := "\"" + quoteContent + "\"\n" + quoteAuthor
	return message
	
}
func parseText(webhookData map[string]interface{}) string {
	events := webhookData["events"].([]interface{})
	event := events[0].(map[string]interface{})
	messages := event["message"].(map[string]interface{})
	text := messages["text"].(string)
	return text
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	webhookData := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&webhookData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// log webhook payload
	fmt.Println("got webhook payload: ")
	for k, v := range webhookData {
		fmt.Printf("%s : %v\n", k, v)
	}
	text := parseText(webhookData)
	fmt.Printf("Your message is: %s\n", text)
	var replyMessage string
	if(text == "/quote") {
		replyMessage = doQuote()
	} else if(text == "/help" || text == "/h") {
		replyMessage = "Hello.\nMy name is Butter. I am an intelligent bot, " +
					"but I was created for the simple tasks. Currently, I support following commands:\n" +
					"/quote - give a random quote"
	} else {
		return
	}
	// get secrets
	var channel_secret, channel_token string
	file_path, err := filepath.Abs("./.bot_configs")
	file, err := os.Open(file_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
		if (channel_secret == "") {
			channel_secret = scanner.Text()
		} else {
			channel_token = scanner.Text()
		}
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
	// reply
	bot, err := linebot.New(channel_secret, channel_token)
	if err != nil {
		fmt.Println("Failed to create bot channel\n")
		return
	}
	events := webhookData["events"].([]interface{})
	event := events[0].(map[string]interface{})
	replyToken := event["replyToken"].(string)
	if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyMessage)).Do(); 
	err != nil {
		fmt.Println("Failed to reply\n")
	}
}

func main() {
    log.Println("server started")
	http.HandleFunc("/", handleWebhook)
	log.Fatal(http.ListenAndServe(":8000", nil))
}