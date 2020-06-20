package main

import (
	"os"
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
			if count == 60 {
				count = 0
				UploadFileToS3()
			}
			count++
		}
	}
}
