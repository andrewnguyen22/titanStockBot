package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

const (
	FacebookApi     = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"
	Subscribed      = " you are now subscribed to this product. If it changes, you will be notified.\n "
	Donate          = "If this bot helped you, consider donating to me :) @ paypal.me/titanfitnessbot!"
	InvalidResponse = "Sorry! That request is not valid. " +
		"Type the product name to subscribe." +
		"You may subscribe to the following products:\n"
)

type Callback struct {
	Object string `json:"object,omitempty"`
	Entry  []struct {
		ID        string      `json:"id,omitempty"`
		Time      int         `json:"time,omitempty"`
		Messaging []Messaging `json:"messaging,omitempty"`
	} `json:"entry,omitempty"`
}

type Messaging struct {
	Sender    User    `json:"sender,omitempty"`
	Recipient User    `json:"recipient,omitempty"`
	Timestamp int     `json:"timestamp,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type User struct {
	UserID string `json:"id,omitempty"`
}

type Message struct {
	MID        string `json:"mid,omitempty"`
	Text       string `json:"text,omitempty"`
	QuickReply *struct {
		Payload string `json:"payload,omitempty"`
	} `json:"quick_reply,omitempty"`
	Attachments *[]Attachment `json:"attachments,omitempty"`
	Attachment  *Attachment   `json:"attachment,omitempty"`
}

type Attachment struct {
	Type    string  `json:"type,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

type Response struct {
	Recipient User    `json:"recipient,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type Payload struct {
	URL string `json:"url,omitempty"`
}

func VerificationEndpoint(w http.ResponseWriter, r *http.Request) {
	challenge := r.URL.Query().Get("hub.challenge")
	token := r.URL.Query().Get("hub.verify_token")

	if token == os.Getenv("verify_token") {
		w.WriteHeader(200)
		_, err := w.Write([]byte(challenge))
		if err != nil {
			fmt.Println("json decoding error in message endpoint", err)
			return
		}
	} else {
		w.WriteHeader(404)
		_, err := w.Write([]byte("Error, wrong validation token"))
		if err != nil {
			fmt.Println("json decoding error in message endpoint", err)
			return
		}
	}
}

func ProcessMessage(m Messaging) {
	text := strings.ToLower(strings.TrimSpace(m.Message.Text))
	fmt.Println("message received: ", text)
	entry, ok := entries[text]
	if ok {
		// subscribe the user
		u := User{UserID: m.Sender.UserID}
		entry.Subs[u.UserID] = struct{}{}
		entries[text] = entry
		text = entry.StatusMsg() + Subscribed + Donate
	} else {
		// invalid query
		text = InvalidResponse + entries.String()
	}
	response := Response{
		Recipient: User{
			UserID: m.Sender.UserID,
		},
		Message: Message{
			Text: text,
		},
	}
	sendMessage(response)
}

func sendMessage(msg Response) {
	client := &http.Client{}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(&msg)
	url := fmt.Sprintf(FacebookApi, os.Getenv("page-token"))
	req, err := http.NewRequest("POST", url, body)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	defer resp.Body.Close()
}

func MessagesEndpoint(w http.ResponseWriter, r *http.Request) {
	var callback Callback
	err := json.NewDecoder(r.Body).Decode(&callback)
	if err != nil {
		fmt.Println("json decoding error in message endpoint", err)
		return
	}
	if callback.Object == "page" {
		for _, entry := range callback.Entry {
			for _, event := range entry.Messaging {
				ProcessMessage(event)
			}
		}
		w.WriteHeader(200)
		_, err := w.Write([]byte("Got your message"))
		if err != nil {
			fmt.Println("json decoding error in message endpoint", err)
			return
		}
	} else {
		w.WriteHeader(404)
		_, err := w.Write([]byte("Message not supported"))
		if err != nil {
			fmt.Println("json decoding error in message endpoint", err)
			return
		}
	}
}

func StartMessengerServer() {
	r := mux.NewRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.HandleFunc("/", VerificationEndpoint).Methods("GET")
	r.HandleFunc("/", MessagesEndpoint).Methods("POST")
	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatal(err)
	}
}

func StockAlertMessage(key string) {
	// get the entry
	e := entries[key]
	for uID := range e.Subs {
		r := Response{
			Recipient: User{UserID: uID},
			Message: Message{
				Text: e.StatusMsg() + Subscribed + Donate,
			},
		}
		sendMessage(r)
	}
}
