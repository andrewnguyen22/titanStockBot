package main

import (
	"flag"
	"fmt"
	"github.com/paked/messenger"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	Welcome = "Welcome to titan fitness stock bot! " +
		"This bot will check the stock status on certain titan fitness products and notify you of stock changes. " +
		"Type the product name to subscribe." +
		"You may subscribe to the following options:\n"
	InvalidResponse = "Sorry! That request is not valid. " +
		"Type the product name to subscribe." +
		"You may subscribe to the following products:\n"
)

var (
	WelcomePrompt         = Welcome + entries.String()
	InvalidResponsePrompt = InvalidResponse + entries.String()
)

var (
	verify = flag.Bool("should-verify", false, "Whether or not the app should verify itself")
	host   = flag.String("host", "0.0.0.0", "The host used to serve the messenger bot")
)

func StartMessengerServer() {
	appSecret := os.Getenv("app-secret")
	pageToken := os.Getenv("page-token")
	verifyToken := os.Getenv("verify_token")
	flag.Parse()
	if verifyToken == "" || appSecret == "" || pageToken == "" {
		fmt.Println("missing environment variables")
		fmt.Println("appSecret", appSecret)
		fmt.Println("pageToken", pageToken)
		fmt.Println("verifyToken", verifyToken)
		os.Exit(-1)
	}
	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      *verify,
		AppSecret:   appSecret,
		VerifyToken: verifyToken,
		Token:       pageToken,
	})
	// setup greeting
	client.GreetingSetting(WelcomePrompt)
	// Setup a handler to be triggered when a message is received
	client.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		text := strings.ToLower(strings.TrimSpace(m.Text))
		if entry, ok := entries[text]; ok {
			// subscribe the user
			u := User{UserID: m.Sender.ID}
			entry.Subs[u.String()] = struct{}{}
			entries[text] = entry
			// get the status message
			if e, ok := entries[text]; !ok {
				fmt.Println("critical error", text, "found in subscriptions, but not entries")
				err := r.Text("uh oh, there's a problem with the bot! try again later :)", messenger.ResponseType)
				if err != nil {
					fmt.Println("Messaging error: ", err.Error())
				}
			} else {
				err := r.Text(e.StatusMsg(), messenger.ResponseType)
				if err != nil {
					fmt.Println("Messaging error: ", err.Error())
				}
			}
		} else {
			// invalid query
			err := r.Text(InvalidResponsePrompt, messenger.ResponseType)
			if err != nil {
				fmt.Println("Messaging error: ", err.Error())
			}
		}
	})
	port := os.Getenv("PORT")
	addr := fmt.Sprintf("%s:%d", *host, port)
	fmt.Println("Serving messenger bot on", addr)
	log.Fatal(http.ListenAndServe(addr, client.Handler()))
}

func StockAlertMessage(key string) {
	appSecret := os.Getenv("app-secret")
	pageToken := os.Getenv("page-token")
	verifyToken := os.Getenv("verify_token")
	if verifyToken == "" || appSecret == "" || pageToken == "" {
		fmt.Println("missing arguments")
		fmt.Println("appSecret", appSecret)
		fmt.Println("pageToken", pageToken)
		fmt.Println("verifyToken", verifyToken)
		os.Exit(-1)
	}
	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      *verify,
		AppSecret:   appSecret,
		VerifyToken: verifyToken,
		Token:       pageToken,
	})
	// get the entry
	e := entries[key]
	for uID := range e.Subs {
		// convert the uID to int64
		i, err := strconv.ParseInt(uID, 10, 64)
		if err != nil {
			fmt.Println("error converting uid to int:", uID, err)
			continue
		}
		// send
		err = client.Send(messenger.Recipient{ID: i}, e.StatusMsg(), messenger.NonPromotionalSubscriptionType)
		if err != nil {
			fmt.Println("error sending the message to:", uID, err)
			continue
		}
	}
}
