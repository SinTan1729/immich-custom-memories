package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
)

func cleanUpMemories(client *http.Client, config *config) error {
	fmt.Println("Cleaning up older memories.")
	body, err := callRequest(client, "GET", config, "/api/memories", nil)
	if err != nil {
		return fmt.Errorf("Error deleting old memories: %s\n", err)
	}

	var memories []memory
	if err = json.Unmarshal(body, &memories); err != nil {
		return fmt.Errorf("Error deleting old memories: %s\n", err)
	}
	for _, memory := range memories {
		id := strings.Trim(memory.Id, `"`)
		_, err := callRequest(client, "DELETE", config, "/api/memories/"+id, nil)
		if err != nil {
			return err
		}
	}

	fmt.Println(" ", len(memories), "memories cleaned up.")
	return nil
}

func generateMemories(client *http.Client, allImages *map[int][]searchResult, config *config, date *date) error {
	err := cleanUpMemories(client, config)
	if err != nil {
		return err
	}

	showAt := time.Date(date.year, date.month, date.day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	hideAt := time.Date(date.year, date.month, date.day, 23, 59, 59, 999999999, time.UTC).Format(time.RFC3339Nano)

	fmt.Println("----------\nAdding new memories.")
	for year, images := range *allImages {
		assets := make([]string, len(images))
		for i, image := range images {
			assets[i] = image.Id
		}
		data := memoryEntry{
			AssetIDs: assets,
			Data:     memoryEntryData{Year: year},
			MemoryAt: time.Date(year, date.month, date.day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			ShowAt:   showAt,
			HideAt:   hideAt,
			Type:     "on_this_day",
		}
		jsonData, _ := json.Marshal(data)

		_, err := callRequest(client, "POST", config, "/api/memories", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("Error creating memories for year %d: %s", year, err)
		}
		if len(images) == 1 {
			fmt.Printf("  Created memory for year %d with 1 entry.\n", year)
		} else {
			fmt.Printf("  Created memory for year %d with %d entries.\n", year, len(images))
		}
	}

	return nil
}

func chooseImages(items *[]searchResult, n int) []searchResult {
	if n <= 0 {
		return nil
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
	// then with more people, then with tags, then randomly break ties
	slices.SortFunc(*items, func(a, b searchResult) int {
		return cmp.Or(
			cmpBool(a.IsFavorite, b.IsFavorite),
			cmp.Compare(b.numFavPeople, a.numFavPeople),
			cmp.Compare(b.numPeople, a.numPeople),
			cmpBool(a.HasTags, b.HasTags),
			cmp.Compare(a.sortSeed, b.sortSeed),
		)
	})

	n = min(n, len(*items))
	return (*items)[:n]
}
