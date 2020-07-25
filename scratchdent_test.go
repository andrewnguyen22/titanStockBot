package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"os"
	"strings"
	"testing"
)

func TestURLIDs(t *testing.T) {
	err := entries.FromJSONFile()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	for key, entry := range entries {
		if entry.Subs == nil {
			entry.Subs = make(map[string]struct{}) // map[userIDString]Empty
			entries[key] = entry
		}
		if !strings.Contains(entry.Name, "t3 tall rack") && !strings.Contains(entry.Name, "t3 short rack") && !strings.Contains(entry.Name, "scratch and dent") {
			ss := strings.Split(entry.URL, "/")
			id := ss[len(ss)-1][:len(ss[len(ss)-1])-5]
			fmt.Println(id)
		}
	}
}

func TestScratchAndDent(t *testing.T) {
	// init collector
	c := colly.NewCollector()
	c.OnHTML(".product", func(e *colly.HTMLElement) {
		name := e.ChildText(".gtm-product-list")
		fmt.Println("NAME: ", name)
		res := e.Attr("data-pid")
		fmt.Println("ID: " + res)
		text := e.ChildText(".price")
		fmt.Println("PRICE: " + text)
		l := e.ChildAttrs(".gtm-product-list", "href")
		var link string
		if len(l) != 0 {
			link = "http://titan.fitness" + l[0]
		} else {
			link = "http://titan.fitness/scratch-and-dent"
		}
		fmt.Println("Link: ", link)
	})
	err := c.Visit("https://www.titan.fitness/scratch-and-dent/")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestConverter(t *testing.T) {
	err := entries.FromJSONFile()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	for key, entry := range entries {
		if entry.Subs == nil {
			entry.Subs = make(map[string]struct{}) // map[userIDString]Empty
			entries[key] = entry
		}
		entry.Name += " scratch and dent"
		entry.URL = "http://titan.fitness/scratch-and-dent"
		entry.Status = OutOfStock
		entries[key+" scratch and dent"] = entry
	}
	err = entries.ToJSONFile()
	if err != nil {
		panic(err)
	}
}
