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
	name = strings.TrimSpace(strings.ToLower(name))
	// init collector
	c := colly.NewCollector()
	// custom logic for t3 page tall
	if strings.Contains(name, "t3 tall rack") {
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
	} else if strings.Contains(name, "t3 short rack") {
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
	} else if strings.Contains(name, "scratch and dent") {
		// sanity check length
		if len(name) < 17 {
			fmt.Println("ERROR: in scratch and dent: the name length is < 17 characters")
			return
		}
		// set stock status to out of stock
		ss = OutOfStock
		// retrieve the real item by name
		entry, found := entries[name[:16]]
		// if the url isn't found return
		if !found || entry.URL == "" {
			fmt.Println("ERROR: in scratch and dent: corresponding entry not found or URL is empty")
			return
		}
		// for each product found at scratch and dent, compare ID's
		c.OnHTML(".product", func(e *colly.HTMLElement) {
			// get the product id from the url of the corresponding product
			id := productIDFromURL(entry.URL)
			idFromPageTile := e.Attr("data-pid")
			if strings.Contains(idFromPageTile, id) {
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

func productIDFromURL(url string) string {
	splitURLArr := strings.Split(url, "/")
	// get the length of the split url array
	arrLen := len(splitURLArr)
	// get the end of the url
	endOfURL := splitURLArr[arrLen-1]
	// substring the end of the url by 5 chars to remove .html
	return endOfURL[:len(endOfURL)-5]
}
