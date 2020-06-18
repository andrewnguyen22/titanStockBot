package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
	"time"
)

func ScrapeAllEntries(entries Entries) {
	for _, entry := range entries {
		fmt.Println("checking stock status of: ", entry.Name)
		stockS, err := scrapeTitanURL(entry.Name, entry.URL)
		if err != nil {
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
			fmt.Println("sending stock alert message")
			StockAlertMessage(entry.Name)
		}
	}
}

func scrapeTitanURL(name, url string) (ss StockStatus, err error) {
	// init collector
	c := colly.NewCollector()
	// custom logic for t3 page <thanks titan :)>
	if name == "t3 tall rack" {
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			fmt.Println("in t3 tall...")
			optionTxt:=strings.ToLower(strings.TrimSpace(e.Text))
			fmt.Println("option text: ", optionTxt)
			// Print link
			if strings.Contains(optionTxt, "tall") {
				fmt.Println("t3 tall stock check... ", ss.String())
				ss = stockCheck(e)
				return
			}
		})
	} else if name == "t3 short rack" {
		// check for option
		c.OnHTML("option", func(e *colly.HTMLElement) {
			fmt.Println("in t3 short...")
			optionTxt:=strings.ToLower(strings.TrimSpace(e.Text))
			fmt.Println("option text: ", optionTxt)
			// Print link
			if strings.Contains(optionTxt, "short") {
				fmt.Println("t3 short stock check... ", ss.String())
				ss = stockCheck(e)
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
	fmt.Println("visiting url: ", url)
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
