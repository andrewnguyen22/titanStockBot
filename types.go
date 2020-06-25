package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
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
		fmt.Println(fmt.Sprintf("invalid stock status: %d", ss))
		return "invalid stock status"
	}
}

type Entry struct {
	Name          string      `json:"key"`
	URL           string      `json:"url"`
	TimeStamp     time.Time   `json:"time"`
	Status        StockStatus `json:"stock-status"`
	Subs          Subscribers `json:"subscribers"`
	Notifications int         `json:"notifications"`
}

func (e Entry) StatusMsg() string {
	if e.Status == InStock {
		return fmt.Sprintf("As of %d minutes ago, %s is now in stock!\nGet it at %s", int(time.Since(e.TimeStamp).Round(time.Minute).Minutes()), e.Name, e.URL)
	}
	return fmt.Sprintf("As of %d minutes ago, %s is out of stock", int(time.Since(e.TimeStamp).Round(time.Minute).Minutes()), e.Name)
}

type Entries map[string]Entry

func (e *Entries) ToJSONFile() error {
	return ToJSON(EntriesDB, e)
}

func (e *Entries) FromJSONFile() error {
	return FromJSON(EntriesDB, e)
}

func (e *Entries) String() (s string) {
	l := len(*e)
	sortedEntries := make([]string, l)
	i := 0
	for key := range (*e) {
		sortedEntries[i] = key
		i++
	}
	sort.Strings(sortedEntries)
	for i, name := range sortedEntries {
		if i == l-1 {
			s += fmt.Sprintf("%s", name)
			continue
		}
		s += fmt.Sprintf("%s,\n", name)
	}
	return
}

func (e *Entries) ClearNotifications() {
	fmt.Println("clearing notifications...")
	for key, entry := range *e {
		fmt.Println("clearing notification for ", key)
		// clear notifications to 0
		entry.Notifications = 0
		// reset entry
		(*e)[key] = entry
		fmt.Println("notification check")
		fmt.Println((*e)[key])
	}
}

func (e *Entries) AddSubscription(key string, user User) {
	// get the sub
	entry := (*e)[key]
	// add the user to the sub with empty structure
	entry.Subs[user.UserID] = struct{}{}
	// set the sub
	(*e)[key] = entry
}

func (e *Entries) Unsubscribe(u User, key string) error {
	// get the sub
	entry, ok := (*e)[key]
	if !ok {
		return fmt.Errorf("item: %s is not a valid option to unsubscribe from", key)
	}
	// delete the user from the sub
	delete(entry.Subs, u.UserID)
	// set the sub
	(*e)[key] = entry
	// return no error
	return nil
}

func (e *Entries) UnsubscribeAll(u User) {
	for key := range *e {
		// get the sub
		entry := (*e)[key]
		// delete the user from the sub
		delete(entry.Subs, u.UserID)
		// set the sub
		(*e)[key] = entry
	}
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
