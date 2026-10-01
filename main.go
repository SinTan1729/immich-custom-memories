// SPDX-FileCopyrightText: 2025 Sayantan Santra <sayantan.santra689@gmail.com>
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var Version = "unknown"

func main() {
	if len(os.Args) > 1 && slices.Contains([]string{"--version", "-V"}, os.Args[1]) {
		fmt.Println(Version)
		return
	}

	now := time.Now()

	fmt.Println("Immich Custom Memories Album")
	fmt.Println("https://github.com/SinTan1729/immich-custom-memories-album")
	fmt.Println("----------")

	var configPath string
	var dry_run bool
	flag.StringVar(&configPath, "config", "", "Path for the config file.")
	flag.BoolVar(&dry_run, "dry_run", false, "Do a dry run.")
	flag.Parse()
	if dry_run {
		fmt.Println("Doing a dry run...")
	}
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatal("Error getting the config directory: ", err)
		}
		configPath = filepath.Join(configDir, "/immich-custom-memories/config.json")
	}

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	var config config
	if err = json.Unmarshal(configFile, &config); err != nil {
		log.Fatal("Error reading config: ", err)
	}
	if config.NoOfYears == 0 {
		config.NoOfYears = 10
	}
	if config.MaxMemorySize == 0 {
		config.MaxMemorySize = 10
	}
	configPrint, _ := json.MarshalIndent(&config, " ", " ")
	fmt.Println("Starting processing memories at", now.Format(time.RFC850))
	fmt.Println("Using config:\n", string(configPrint))
	fmt.Println("----------")

	date := date{now.Year(), now.Month(), now.Day()}
	client := &http.Client{}
	allImages := make(map[int][]searchResult)
	totalMemories := 0

	var tagIds, personIds []string
	if tagIds, err = getTagIds(client, &config); err != nil {
		log.Fatalln(err)
	}
	if personIds, err = getPersonIds(client, &config); err != nil {
		log.Fatalln(err)
	}

	curYear := now.Year()
	for year := curYear - 1; year >= curYear-config.NoOfYears; year-- {
		fmt.Println("Processing year:", year)
		date.year = year
		images, err := getYearImages(client, &config, &date, personIds, tagIds)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("  Got %d images for the date.\n", len(images))
		if len(images) > config.MaxMemorySize {
			fmt.Printf("  Choosing %d images for the memory based on heuristics.\n", config.MaxMemorySize)
			images = chooseImages(&images, config.MaxMemorySize)
		}
		if len(images) > 0 {
			allImages[year] = images
			totalMemories += 1
		}
	}

	date.year = curYear
	if dry_run {
		fmt.Printf("[Dry run] Will create %d memories.\n", totalMemories)
	} else {
		if err = generateMemories(client, &allImages, &config, &date); err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Total created memories: %d\n", totalMemories)
	}
}
