package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	EntriesDB = "entries.json"
)

var (
	entries = make(Entries, 0)
)

func init() {
	DownloadFileFromS3()
	err := entries.FromJSONFile()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	for key, entry := range entries {
		if entry.Subs == nil {
			entry.Subs = make(map[string]struct{}) // map[userIDString]Empty
			entries[key] = entry
		}
	}
	ScrapeAllEntries(entries)
}

func main() {
	go StartMessengerServer()
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		os.Kill, //nolint
		os.Interrupt)
	defer func() {
		sig := <-signalChannel
		fmt.Println(fmt.Sprintf("Exit signal %s received\n", sig))
		UploadFileToS3()
		os.Exit(3)
	}()
	PeriodicallyCheckTitanFitness(time.Minute * 1)
}

var count = 0

func PeriodicallyCheckTitanFitness(duration time.Duration) {
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			ScrapeAllEntries(entries)
			err := entries.ToJSONFile()
			if err != nil {
				panic(err)
			}
			if count == 30 {
				count = 0
				UploadFileToS3()
			}
			count++
		}
	}
}
