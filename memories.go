package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func cleanUpMemories(client *http.Client, config *config) error {
	fmt.Println("Cleaning up older memories.")
	req, err := http.NewRequest("GET", config.ServerUrl+"/api/memories/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Error fetching images: " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var memories []memory
	err = json.Unmarshal(body, &memories)
	if err != nil {
		return err
	}
	for _, memory := range memories {
		id := strings.Trim(memory.Id, `"`)

		req, err := http.NewRequest("DELETE", config.ServerUrl+"/api/memories/"+id, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-API-Key", config.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 204 {
			return errors.New("Error deleting old memories: " + id + " : " + resp.Status)
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

		req, err := http.NewRequest("POST", config.ServerUrl+"/api/memories", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", config.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
			defer resp.Body.Close()
			return errors.New("Error creating memories: " + strconv.Itoa(year) + " : " + resp.Status)
		}
		if len(images) == 1 {
			fmt.Printf("  Created memory for year %d with 1 entry.\n", year)
		} else {
			fmt.Printf("  Created memory for year %d with %d entries.\n", year, len(images))
		}
	}

	return nil
}
