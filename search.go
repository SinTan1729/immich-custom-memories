// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"slices"
	"strings"
	"time"
)

type searchResult struct {
	Id            string    `json:"id"`
	LocalDateTime time.Time `json:"localDateTime"`
	People        []person  `json:"people"`
	peopleIDs     []string  `json:""`
	peopleNames   []string
	HasMetadata   bool   `json:"hasMetadata"`
	IsFavorite    bool   `json:"isFavorite"`
	IsOffline     bool   `json:"isOffline"`
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
	item      searchResult
	favorite  bool
	favPeople int
	people    int
	random    uint64
}

func getYearImages(client *http.Client, config *config, date *date) ([]searchResult, error) {
	earliestZone, _ := time.LoadLocation("Etc/GMT-14")
	lastZone, _ := time.LoadLocation("Etc/GMT+12")
	data := searchParams{
		Filter: filter{
			Type: typeFilter{Eq: "IMAGE"},
			TakenAt: dateFilter{
				After:  time.Date(date.year, date.month, date.day, 0, 0, 0, 0, earliestZone),
				Before: time.Date(date.year, date.month, date.day, 11, 59, 59, 999999999, lastZone),
			},
		},
		WithPeople: true,
		WithExif:   true,
	}
	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", config.ServerUrl+"/api/search/metadata", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Error fetching images: " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res searchResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	validItems := make([]searchResult, len(res.Assets.Items))
	for i, item := range res.Assets.Items {
		isVisible := item.Visibility == "timeline"
		if !item.HasMetadata || item.IsOffline || !isVisible {
			fmt.Println("Skipping bad item:", item.Id)
			continue
		}

		for _, person := range item.People {
			item.peopleIDs = append(item.peopleIDs, person.Id)
			item.peopleNames = append(item.peopleIDs, person.Name)
		}
		validItems[i] = item
	}

	var filteredItems []searchResult
	for _, item := range validItems {
		if item.LocalDateTime.Day() == date.day {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

func filterPeople(items *[]searchResult, config *config) []searchResult {
	if len(config.ExcludedPeople) == 0 {
		return *items
	}
	var filteredItems []searchResult
	for _, item := range *items {
		includeItem := true
		for _, person := range config.ExcludedPeople {
			if slices.Contains(item.peopleIDs, person) || slices.Contains(item.peopleNames, person) {
				includeItem = false
				break
			}
		}

		if includeItem {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems
}

func filterTags(client *http.Client, items *[]searchResult, config *config) ([]searchResult, error) {
	if len(config.ExcludedTags) == 0 {
		return *items, nil
	}
	var filteredItems []searchResult
	for _, item := range *items {
		req, err := http.NewRequest("GET", config.ServerUrl+"/api/assets/"+item.Id, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-API-Key", config.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("Error fetching image info: " + item.Id + " : " + resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		includeItem := true
		var res tagResponse
		err = json.Unmarshal(body, &res)
		if err != nil {
			return nil, err
		}
		for _, tag := range res.Tags {
			filterFunc := func(excludedTag string) bool {
				return tag.Value == excludedTag ||
					strings.HasSuffix(tag.Value, "/"+excludedTag) ||
					strings.HasPrefix(tag.Value, excludedTag+"/") ||
					strings.Contains(tag.Value, "/"+excludedTag+"/")
			}
			if slices.ContainsFunc(config.ExcludedTags, filterFunc) {
				includeItem = false
				break
			}
		}

		if includeItem {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems, nil
}

func chooseImages(items *[]searchResult, n int) []searchResult {
	if n <= 0 {
		return nil
	}

	ranked := make([]rankedResult, len(*items))
	for i, item := range *items {
		favPeople := 0
		for _, p := range item.People {
			if p.IsFavorite {
				favPeople++
			}
		}

		ranked[i] = rankedResult{
			item:      item,
			favorite:  item.IsFavorite,
			favPeople: favPeople,
			people:    len(item.peopleIDs),
			random:    rand.Uint64(),
		}
	}

	cmpBool := func(a, b bool) int {
		if a == b {
			return 0
		}
		if a {
			return -1
		}
		return 1
	}
	// Favorite images first, then with more favorite people,
	// then with more people, then randomly break ties
	slices.SortFunc(ranked, func(a, b rankedResult) int {
		return cmp.Or(
			cmpBool(a.favorite, b.favorite),
			cmp.Compare(b.favPeople, a.favPeople),
			cmp.Compare(b.people, a.people),
			cmp.Compare(a.random, b.random),
		)
	})

	n = min(n, len(ranked))
	result := make([]searchResult, n)
	for i := range n {
		result[i] = ranked[i].item
	}
	return result
}
