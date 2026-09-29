// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import "time"

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
	Year int `json:"year"`
}

type memory struct {
	Id string `json:"id"`
}

// Searching

type searchResult struct {
	Id            string    `json:"id"`
	LocalDateTime time.Time `json:"localDateTime"`
	People        []person  `json:"people"`
	peopleIDs     []string  `json:""`
	peopleNames   []string
	HasMetadata   bool `json:"hasMetadata"`
	IsFavorite    bool `json:"isFavorite"`
	IsOffline     bool `json:"isOffline"`
	HasTags       bool
	Visibility    string `json:"visibility"`
}
type person struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	IsFavorite bool   `json:"isFavorite"`
}

type searchParams struct {
	Filter     filter `json:"filter"`
	WithPeople bool   `json:"withPeople"`
	WithExif   bool   `json:"withExif"`
}
type filter struct {
	Type    typeFilter `json:"type"`
	TakenAt dateFilter `json:"takenAt"`
}
type typeFilter struct {
	Eq string `json:"eq"`
}
type dateFilter struct {
	After  time.Time `json:"gte"`
	Before time.Time `json:"lt"`
}

type searchResponse struct {
	Assets struct {
		Items []searchResult `json:"items"`
	} `json:"assets"`
}

type tagResponse struct {
	Tags []tag `json:"tags"`
}
type tag struct {
	Value string `json:"value"`
}

type rankedResult struct {
	item         searchResult
	isFavorite   bool
	numFavPeople int
	numPeople    int
	hasTags      bool
	random       uint64
}
