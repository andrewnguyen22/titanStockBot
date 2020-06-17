package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	InStock StockStatus = iota + 1
	OutOfStock
)

type StockStatus int

func (ss StockStatus) String() string {
	if ss == InStock {
		return "In Stock"
	} else if ss == OutOfStock {
		return "Out Of Stock"
	} else {
		panic(fmt.Sprintf("invalid stock status: %d", ss))
	}
}

type Entry struct {
	Name      string      `json:"key"`
	URL       string      `json:"url"`
	TimeStamp time.Time   `json:"time"`
	Status    StockStatus `json:"stock-status"`
	Subs      Subscribers `json:"subscribers"`
}

func (e Entry) StatusMsg() string {
	location, _ := time.LoadLocation("EST")
	if e.Status == InStock {
		return fmt.Sprintf("%s is now in stock! As of %s.\nGet it at %s", e.Name, e.TimeStamp.In(location), e.URL)
	}
	return fmt.Sprintf("%s is out of stock... As of %s", e.Name, e.TimeStamp.In(location))
}

type Entries map[string]Entry

func (e *Entries) ToJSONFile() error {
	return ToJSON(EntriesDB, e)
}

func (e *Entries) FromJSONFile() error {
	return FromJSON(EntriesDB, e)
}

func (e *Entries) String() (s string) {
	for name := range *e {
		s += fmt.Sprintf("%s,\n", name)
	}
	return
}

func (e *Entries) AddSubscription(key string, user User) {
	// get the sub
	entry := (*e)[key]
	// add the user to the sub with empty structure
	entry.Subs[user.UserID] = struct{}{}
	// set the sub
	(*e)[key] = entry
}

func (e *Entries) RemoveSubscription(key string, u User) {
	// get the sub
	entry := (*e)[key]
	// delete the user from the sub
	delete(entry.Subs, u.UserID)
	// set the sub
	(*e)[key] = entry
}

type Subscribers map[string]struct{}

func FromJSON(f string, i interface{}) error {
	plan, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	err = json.Unmarshal(plan, &i)
	if err != nil {
		return err
	}
	return nil
}

func ToJSON(f string, i interface{}) error {
	file, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(f, file, 0644)
	if err != nil {
		return err
	}
	return nil
}
