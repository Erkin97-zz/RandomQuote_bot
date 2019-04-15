package main

import (
	"bufio"
	"encoding/json"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/line/line-bot-sdk-go/linebot"
)

func getQuote() string {
	resp, err := http.Get("http://quotesondesign.com/wp-json/posts?filter[orderby]=rand&filter[posts_per_page]=1")
	if err != nil {
		log.Printf("Failed to request quote. Error: %v\n", err)
		return ""
	}

	var quoteData []interface{}
	err = json.NewDecoder(resp.Body).Decode(&quoteData)
	if err != nil {
		log.Printf("Failed to request quote. Error: %v\n", err)
		return ""
	}

	log.Printf("Your quote data %v\n", quoteData)

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
	log.Println("got webhook payload: ")
	for k, v := range webhookData {
		log.Printf("%s : %v\n", k, v)
	}
	text := parseText(webhookData)
	log.Printf("Your message is: %s\n", text)

	var replyMessage string
	if text == "/quote" {
		replyMessage = getQuote()
	} else if text == "/help" || text == "/h" {
		replyMessage = "Hello.\nMy name is Butter. I am an intelligent bot, " +
			"but I was created for the simple tasks. Currently, I support following commands:\n" +
			"/quote - give a random quote"
	} else {
		return
	}

	// get secrets
	var channelSecret, channelToken string
	filePath, err := filepath.Abs("./.bot_configs")
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if channelSecret == "" {
			channelSecret = scanner.Text()
		} else {
			channelSecret = scanner.Text()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// reply
	bot, err := linebot.New(channelSecret, channelToken)
	if err != nil {
		log.Println("Failed to create bot channel")
		return
	}
	events := webhookData["events"].([]interface{})
	event := events[0].(map[string]interface{})
	replyToken := event["replyToken"].(string)
	if _, err := bot.ReplyMessage(replyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
		log.Println("Failed to reply")
	}
}

func main() {
	log.Println("server started")
	http.HandleFunc("/", handleWebhook)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
