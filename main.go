package main

import (
	"os"
	"time"
)

const (
	EntriesDB = "entries.json"
	SubsDB    = "subs.json"
)

var (
	entries = make(Entries, 0)
)

func init() {
	err := entries.FromJSONFile()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	for _, entry := range entries {
		entry.Subs =  make(map[string]struct{}) // map[userIDString]Empty
	}
	ScrapeAllEntries(entries)
}

func main() {
	//go StartMessengerServer()
	PeriodicallyCheckTitanFitness(time.Minute)
}

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
		}
	}
}
