/*
simple webhook server for my bot
this spaghetti code just for testing golang power
*/
package main

import (
	"fmt"
//	"io"
	"log"
	"net/http"
//	"os"
	"encoding/json"
)

func doQuote() string {
    return "some quote"
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
}

func main() {
    log.Println("server started")
	http.HandleFunc("/", handleWebhook)
	log.Fatal(http.ListenAndServe(":8000", nil))
}