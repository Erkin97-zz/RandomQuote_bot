package main

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	strip "github.com/grokify/html-strip-tags-go"
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

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BUTTER"))
	if err != nil {
		log.Printf("Failed to connect with bot. Error: %v\n", err)
		return
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if strings.Contains(update.Message.Text, "quote") { // give quote
			if quoteMessage := getQuote(); quoteMessage != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, quoteMessage)
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)
			}
		}
	}
}
