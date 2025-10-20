// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type searchResult struct {
	id          string
	localTime   time.Time
	path        string
	peopleIDs   []string
	peopleNames []string
	tags        []string
}

func getYearImages(client *http.Client, config *config, date *date) ([]searchResult, error) {
	earliestZone, _ := time.LoadLocation("Etc/GMT-14")
	lastZone, _ := time.LoadLocation("Etc/GMT+12")
	takenAfter := time.Date(date.year, date.month, date.day, 0, 0, 0, 0, earliestZone)
	takenBefore := time.Date(date.year, date.month, date.day, 11, 59, 59, 999999999, lastZone)
	jsonData := fmt.Sprintf(`{"type":"IMAGE","takenAfter":"%s","takenBefore":"%s","withPeople":true}`,
		takenAfter.Format(time.RFC3339), takenBefore.Format(time.RFC3339))
	req, err := http.NewRequest("POST", config.ServerUrl+"/api/search/metadata", bytes.NewBufferString(jsonData))
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

	parsedJson, err := gabs.ParseJSON(body)
	if err != nil {
		return nil, err
	}

	items := parsedJson.Path("assets.items").Children()
	parsedItems := make([]searchResult, len(items))
	for i, item := range items {
		var parsedItem searchResult
		parsedItem.id = strings.Trim(item.Path("id").String(), `"`)
		parsedItem.path = strings.Trim(item.Path("originalPath").String(), `"`)
		parsedItem.localTime, _ = time.Parse(time.RFC3339, strings.Trim(item.Path("localDateTime").String(), `"`))
		people := item.Path("people").Children()
		parsedItem.peopleIDs = make([]string, len(people))
		parsedItem.peopleNames = make([]string, len(people))
		for j, person := range people {
			parsedItem.peopleIDs[j] = strings.Trim(person.Path("id").String(), `"`)
			parsedItem.peopleNames[j] = strings.Trim(person.Path("name").String(), `"`)
		}
		parsedItems[i] = parsedItem
	}

	var filteredItems []searchResult
	for _, item := range parsedItems {
		if item.localTime.Day() == date.day {
			filteredItems = append(filteredItems, item)
		}
	}

	return filteredItems, nil
}

func filterPeople(items *[]searchResult, config *config) []searchResult {
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
	var filteredItems []searchResult
	for _, item := range *items {
		req, err := http.NewRequest("GET", config.ServerUrl+"/api/assets/"+item.id, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-API-Key", config.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, errors.New("Error fetching image info: " + item.id + " : " + resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		includeItem := true
		parsedJSON, err := gabs.ParseJSON(body)
		if err != nil {
			return nil, err
		}
		for _, tag := range parsedJSON.Path("tags").Children() {
			tagValue := strings.Trim(tag.Path("value").String(), `"`)
			filterFunc := func(excludedTag string) bool {
				return tagValue == excludedTag ||
					strings.HasSuffix(tagValue, "/"+excludedTag) ||
					strings.HasPrefix(tagValue, excludedTag+"/") ||
					strings.Contains(tagValue, "/"+excludedTag+"/")
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
