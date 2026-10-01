// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"bytes"
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

func callRequest(client *http.Client, t string, config *config, path string, data io.Reader) ([]byte, error) {
	req, err := http.NewRequest(t, config.ServerUrl+path, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if !slices.Contains([]int{200, 201, 204}, resp.StatusCode) {
		return nil, errors.New(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return body, nil

}

func getYearImages(client *http.Client, config *config, date *date, personIds []string, tagIds []string) ([]searchResult, error) {
	earliestZone, _ := time.LoadLocation("Etc/GMT-14")
	lastZone, _ := time.LoadLocation("Etc/GMT+12")
	data := searchParams{
		Filter: filter{
			Type: eqFilterEnum{Eq: "IMAGE"},
			TakenAt: dateFilter{
				After:  time.Date(date.year, date.month, date.day, 0, 0, 0, 0, earliestZone),
				Before: time.Date(date.year, date.month, date.day, 11, 59, 59, 999999999, lastZone),
			},
			ExcludePeople: noneFilter{None: personIds},
			ExcludeTags:   noneFilter{None: tagIds},
			IsOffline:     eqFilterBool{Eq: false},
			Visibility:    eqFilterEnum{Eq: "timeline"},
		},
		WithPeople: true,
		WithExif:   true,
	}

	var cursor string
	var images []searchResult
	for true {
		data.Cursor = cursor
		jsonData, _ := json.Marshal(data)
		body, err := callRequest(client, "POST", config, "/api/search/metadata", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("Error fetching images: %s\n", err)
		}

		var res searchResponse
		if err = json.Unmarshal(body, &res); err != nil {
			return nil, err
		}

		validItems := make([]searchResult, len(res.Assets.Items))
		for i, item := range res.Assets.Items {
			if !item.HasMetadata {
				fmt.Println("Skipping bad item:", item.Id)
				continue
			}

			item.numFavPeople = 0
			item.numPeople = 0
			for _, person := range item.People {
				item.numPeople += 1
				if person.IsFavorite {
					item.numFavPeople += 1
				}
			}
			item.sortSeed = rand.Uint32()
			validItems[i] = item
		}

		var filteredItems []searchResult
		for _, item := range validItems {
			if item.LocalDateTime.Day() == date.day {
				filteredItems = append(filteredItems, item)
			}
		}

		images = append(images, filteredItems...)
		if res.NextCursor != "" {
			cursor = res.NextCursor
		} else {
			break
		}
	}

	return images, nil
}

func getTagIds(client *http.Client, config *config) ([]string, error) {
	body, err := callRequest(client, "GET", config, "/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("Error fetching IDs for tags: %s", err)
	}

	var res []tag
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var out []string
	for _, in := range config.ExcludedTags {
		for _, tag := range res {
			if strings.Contains(tag.Value, in) {
				out = append(out, tag.Id)
			}
			if tag.Id == in {
				out = append(out, tag.Id)
				break
			}
		}
	}

	fmt.Println("Mapped tags to IDs.")
	return out, nil
}

func getPersonIds(client *http.Client, config *config) ([]string, error) {
	page := 1
	var people []person

	for true {
		body, err := callRequest(client, "GET", config, fmt.Sprintf("/api/people?page=%d", page), nil)
		if err != nil {
			return nil, fmt.Errorf("Error fetching IDs for people: %s", err)
		}

		var res personResponse
		if err = json.Unmarshal(body, &res); err != nil {
			return nil, err
		}
		people = append(people, res.People...)
		if !res.HasNextPage {
			break
		}
		page += 1
	}

	var out []string
	for _, in := range config.ExcludedTags {
		for _, p := range people {
			if p.Name == in || p.Id == in {
				out = append(out, p.Id)
				break
			}
		}
	}

	fmt.Println("Mapped people to IDs.")
	return out, nil
}
