package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
	"time"
)

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
			StockAlertMessage(entry.Name)
		}
	}
}

func scrapeTitanURL(name, url string) (ss StockStatus, err error) {
	var tallDepthEnabled, tallHeightEnabled, shortDepthEnabled, shortHeightEnabled = false, false, false, false
	// init collector
	c := colly.NewCollector()
	// custom logic for t3 page tall
	if strings.Contains(strings.ToLower(name), "t3 tall rack") {
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			optionTxt := strings.ToLower(strings.TrimSpace(e.Text))
			// Print link
			if strings.Contains(optionTxt, "tall") {
				// height check on t3 tall
				if stockCheck(e) == InStock {
					// if short is enabled
					tallHeightEnabled = true
				}
			} else if strings.Contains(optionTxt, "24") || strings.Contains(optionTxt, "36") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					tallDepthEnabled = true
				}
			}
			// check to see if both are enabled
			if tallHeightEnabled && tallDepthEnabled {
				// set them to false (may be unnecessary)
				tallHeightEnabled, tallDepthEnabled = false, false
				// return as in stock
				ss = InStock
				return
			}
		})
		// custom logic for t3 short
	} else if strings.Contains(strings.ToLower(name), "t3 short rack") {
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			optionTxt := strings.ToLower(strings.TrimSpace(e.Text))
			// Print link
			if strings.Contains(optionTxt, "short") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					shortHeightEnabled = true
				}
			} else if strings.Contains(optionTxt, "24") || strings.Contains(optionTxt, "36") {
				// height check on t3 short
				if stockCheck(e) == InStock {
					// if short is enabled
					shortDepthEnabled = true
				}
			}
			// check to see if both are enabled
			if shortDepthEnabled && shortHeightEnabled {
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
