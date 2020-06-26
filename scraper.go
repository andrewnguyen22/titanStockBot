package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
	"time"
)

const NotificationLimit = 2

func ScrapeAllEntries(entries Entries) {
	for _, entry := range entries {
		stockS, err := scrapeTitanURL(entry.Name, entry.URL)
		if err != nil || stockS == 0 {
			fmt.Println(err)
			continue
		}
		// temp save old status
		oldStatus := entry.Status
		// update stock status
		entry.Status = stockS
		// update timestamp
		entry.TimeStamp = time.Now()
		// update entry in mapping
		entries[entry.Name] = entry
		// if there was a change,
		if oldStatus != stockS {
			fmt.Println("sending stock alert message for:", entry.Name)
			if entry.Notifications < NotificationLimit {
				// update number of notifications
				entry.Notifications++
				// update entry in mapping
				entries[entry.Name] = entry
				StockAlertMessage(entry.Name)
			} else {
				fmt.Println(entry.Name, "reached the notification limit")
			}
		}
	}
}

func scrapeTitanURL(name, url string) (ss StockStatus, err error) {
	tallDepthEnabled, tallHeightEnabled, shortDepthEnabled, shortHeightEnabled := false, false, false, false
	// init collector
	c := colly.NewCollector()
	// custom logic for t3 page tall
	if strings.Contains(strings.ToLower(name), "t3 tall rack") {
		ss = OutOfStock
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			optionTxt := strings.ToLower(strings.TrimSpace(e.Text))
			// Print link
			if strings.Contains(optionTxt, "tall") {
				// height check on t3 tall
				if stockCheck(e) == InStock {
					// if short is enabled
					fmt.Println("setting tall height enabled as true")
					tallHeightEnabled = true
				}
			} else if strings.Contains(optionTxt, "24") || strings.Contains(optionTxt, "36") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					fmt.Println("setting tall depth enabled as true")
					tallDepthEnabled = true
				}
			}
			// check to see if both are enabled
			if tallHeightEnabled && tallDepthEnabled {
				// set them to false (may be unnecessary)
				fmt.Println("tall height and depth enabled!!")
				tallHeightEnabled, tallDepthEnabled = false, false
				// return as in stock
				ss = InStock
				return
			}
		})
		// custom logic for t3 short
	} else if strings.Contains(strings.ToLower(name), "t3 short rack") {
		ss = OutOfStock
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			optionTxt := strings.ToLower(strings.TrimSpace(e.Text))
			// Print link
			if strings.Contains(optionTxt, "short") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					fmt.Println("short height enabled!!")
					shortHeightEnabled = true
				}
			} else if strings.Contains(optionTxt, "24") || strings.Contains(optionTxt, "36") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					shortDepthEnabled = true
					fmt.Println("short depth enabled!!")
				}
			}
			// check to see if both are enabled
			if shortDepthEnabled && shortHeightEnabled {
				fmt.Println("short height and depth enabled!")
				// set them to false (may be unnecessary)
				shortHeightEnabled, shortDepthEnabled = false, false
				// return as in stock
				ss = InStock
				return
			}
		})
	} else {
		c.OnHTML("button", func(e *colly.HTMLElement) {
			text := strings.ToLower(strings.TrimSpace(e.Text))
			if strings.Contains(text, "cart") || strings.Contains(text, "order") {
				ss = stockCheck(e)
				return
			}
		})
	}
	err = c.Visit(url)
	return
}

func stockCheck(e *colly.HTMLElement) StockStatus {
	if strings.Contains(strings.ToLower(fmt.Sprintf("%v", e)), "disabled") {
		return OutOfStock
	} else {
		return InStock
	}
}
