package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Configs struct {
	Title         string `yaml:"title"`
	FilePath      string `yaml:"filePath"`
	RegexPattern  string `yaml:"regexPattern"`
	MaxLines      int    `yaml:"maxLines"`
	LeftPadding   int    `yaml:"leftPadding"`
	TopPadding    int    `yaml:"topPadding"`
	RightPadding  int    `yaml:"rightPadding"`
	BottomPadding int    `yaml:"bottomPadding"`
}

func main() {

	configFile := "config.yml"

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config.yml file: %v", err)
	}

	configData := make(map[string]Configs)
	e := yaml.Unmarshal(yamlFile, &configData)
	if e != nil {
		log.Fatalf("Error while performing Unmarshal: %v", e)
	}

	graphsData := make(map[string]Graphs)

	i := 0
	for k, v := range configData {
		i++
		fmt.Printf("Plotting: %s[\"%s\"]\n", k, v.Title)
		var gm Graphs

		// TODO: add checks to verify input

		// get file size and set seek to its current size for first time read
		// program will read from the current position and not from beginning
		file, err := os.Open(v.FilePath)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
		fileStat, err := file.Stat()
		if err != nil {
			log.Fatalf("failed to get stats of file: %v", err)
		}

		gm.GraphId = "graph" + strconv.Itoa(i)
		gm.ConfigName = k
		gm.Title = v.Title
		gm.StartFrom = fileStat.Size()
		gm.Values = []float64{0, 0}
		gm.LeftPadding = v.LeftPadding
		gm.TopPadding = v.TopPadding
		gm.RightPadding = v.RightPadding
		gm.BottomPadding = v.BottomPadding
		graphsData[k] = gm

		file.Close()
	}

	DrawGraph(configData, graphsData)

}
