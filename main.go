// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	ServerUrl      string   `json:"serverUrl"`
	APIKey         string   `json:"apiKey"`
	ExcludedPeople []string `json:"excludedPeople"`
	ExcludedTags   []string `json:"excludedTags"`
}

func main() {
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
	fmt.Println(config)

	client := &http.Client{}
	allImages, err := getAllImages(client, &config)
	if err != nil {
		log.Fatalln(err)
	}
	images := filterPeople(&allImages, &config)
	images, err = filterTags(client, &images, &config)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(images)
}
