package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
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

	jsonData, err := gabs.ParseJSON(body)
	if err != nil {
		return err
	}
	for _, memory := range jsonData.Children() {
		id := strings.Trim(memory.Path("id").String(), `"`)
		fmt.Println(memory.StringIndent(" ", " "))

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

	fmt.Println(" ", len(jsonData.Children()), "memories cleaned up.")
	return nil
}

func generateMemories(client *http.Client, allImages *map[int][]searchResult, config *config, date *date) error {
	if config.CleanupDaily {
		err := cleanUpMemories(client, config)
		if err != nil {
			return err
		}
	}

	fmt.Println("Adding new memories.")
	for year, images := range *allImages {
		jsonData := `{"assetIds":[`
		for _, image := range images {
			jsonData += `"` + image.id + `",`
		}
		jsonData = jsonData[:len(jsonData)-1]
		jsonData += fmt.Sprintf(`],"data":{"year":%d},"memoryAt":"`, year)
		memoryTime := time.Date(year, date.month, date.day, 0, 0, 0, 0, time.UTC)
		jsonData += memoryTime.Format(time.RFC3339) + `","type":"on_this_day"}`

		req, err := http.NewRequest("POST", config.ServerUrl+"/api/memories", bytes.NewBufferString(jsonData))
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
		fmt.Println("  Created memory for year", year, "with", len(images), "entries.")
	}

	return nil
}
