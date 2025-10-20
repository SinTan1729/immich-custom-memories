// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	ServerUrl      string   `json:"serverUrl"`
	APIKey         string   `json:"apiKey"`
	ExcludedPeople []string `json:"excludedPeople"`
	ExcludedTags   []string `json:"excludedTags"`
	EarliestYear   int      `json:"earliestYear"`
}

type date struct {
	year  int
	month time.Month
	day   int
}

func main() {
	now := time.Now()

	fmt.Println("Immich Custom Memories Album")
	fmt.Println("https://github.com/SinTan1729/immich-custom-memories-album")
	fmt.Println("----------")
	configFile, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var config config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}
	if config.EarliestYear == 0 {
		config.EarliestYear = 2010
	}
	configPrint, _ := json.MarshalIndent(&config, " ", " ")
	fmt.Println("Starting processing memories at", now.Format(time.RFC850))
	fmt.Println("Using config:\n", string(configPrint))
	fmt.Println("----------")

	date := date{now.Year(), now.Month(), now.Day()}
	client := &http.Client{}
	allImages := make(map[int][]searchResult)
	for year := now.Year() - 1; year >= config.EarliestYear; year-- {
		fmt.Println("Processing year:", year)
		date.year = year
		yearImages, err := getYearImages(client, &config, &date)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("  Got %d images for the date.\n", len(yearImages))
		images := filterPeople(&yearImages, &config)
		fmt.Printf("  After filtering by people, %d images remaining.\n", len(images))
		images, err = filterTags(client, &images, &config)
		fmt.Printf("  After filtering by tags, %d images remaining.\n", len(images))
		if err != nil {
			log.Fatalln(err)
		}
		if len(images) > 0 {
			allImages[year] = images
		}
	}

	err = generateMemories(client, &allImages, &config, &date)
	if err != nil {
		log.Fatalln(err)
	}
}
