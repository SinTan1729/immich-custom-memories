// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import "time"
import "encoding/json"

// Base

type config struct {
	ServerUrl      string   `json:"serverUrl"`
	APIKey         string   `json:"apiKey"`
	ExcludedPeople []string `json:"excludedPeople"`
	ExcludedTags   []string `json:"excludedTags"`
	NoOfYears      int      `json:"noOfYears"`
	MaxMemorySize  int      `json:"maxMemorySize"`
}

type date struct {
	year  int
	month time.Month
	day   int
}

func (d *date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.year, d.month, d.day = t.Date()
	return nil
}

// Memories

type memoryEntry struct {
	AssetIDs []string        `json:"assetIds"`
	Data     memoryEntryData `json:"data"`
	MemoryAt string          `json:"memoryAt"`
	ShowAt   string          `json:"showAt"`
	HideAt   string          `json:"hideAt"`
	Type     string          `json:"type"`
}

type memoryEntryData struct {
	Id   *string `json:"personId,omitempty"`
	Name *string `json:"personName,omitempty"`
	Year int     `json:"year"`
}

type memory struct {
	Id string `json:"id"`
}

// Searching

type searchResult struct {
	Id            string    `json:"id"`
	LocalDateTime time.Time `json:"localDateTime"`
	People        []person  `json:"people"`
	numPeople     int
	numFavPeople  int
	HasMetadata   bool `json:"hasMetadata"`
	IsFavorite    bool `json:"isFavorite"`
	HasTags       bool
	sortSeed      uint32
}

type personResponse struct {
	HasNextPage bool     `json:"hasNextPage"`
	People      []person `json:"people"`
}
type person struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	IsFavorite bool   `json:"isFavorite"`
	Birthday   *date  `json:"birthDate"`
}

type searchParams struct {
	Cursor     string `json:"cursor,omitempty"`
	Filter     filter `json:"filter"`
	WithPeople bool   `json:"withPeople"`
	WithExif   bool   `json:"withExif"`
}
type filter struct {
	Type          eqFilterEnum `json:"type"`
	Visibility    eqFilterEnum `json:"visibility"`
	IsOffline     eqFilterBool `json:"isOffline"`
	TakenAt       dateFilter   `json:"takenAt"`
	ExcludeTags   noneFilter   `json:"tagIds"`
	ExcludePeople noneFilter   `json:"personIds"`
}
type eqFilterEnum struct {
	Eq string `json:"eq"`
}
type eqFilterBool struct {
	Eq bool `json:"eq"`
}
type dateFilter struct {
	After  time.Time `json:"gte"`
	Before time.Time `json:"lt"`
}
type noneFilter struct {
	None []string `json:"none"`
}

type searchResponse struct {
	Assets struct {
		Items []searchResult `json:"items"`
	} `json:"assets"`
	NextCursor string `json:"nextCursor"`
}

type tag struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type birthdayResults struct {
	name   string
	images []searchResult
}
