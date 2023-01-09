package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
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
	var title string
	var fileToRead string
	var rePattern string
	var maxLinesToRead int
	var leftPadding int
	var topPadding int
	var rightPadding int
	var bottomPadding int

	configFile := "config.yml"
	var startFrom int64 = 0

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config.yml file: %v", err)
	}

	configData := make(map[string]Configs)
	e := yaml.Unmarshal(yamlFile, &configData)
	if e != nil {
		log.Fatalf("Error while performing Unmarshal: %v", e)
	}

	for k, v := range configData {
		fmt.Printf("Plotting: %s[\"%s\"]\n", k, v.Title)
		title = v.Title
		fileToRead = v.FilePath
		rePattern = v.RegexPattern
		maxLinesToRead = v.MaxLines
		leftPadding = v.LeftPadding
		topPadding = v.TopPadding
		rightPadding = v.RightPadding
		bottomPadding = v.BottomPadding
		break // plotting only single graph for now - multiple plots for future
	}

	// get file size and set seek to its current size for first time read
	// program will read from the current position and not from beginning
	file, err := os.Open(fileToRead)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.Fatalf("failed to get stats of file: %v", err)
	}
	startFrom = fileStat.Size()
	file.Close()

	// initialize termui
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	// give some initial entries for the plot - in case if no data found on first read
	var plotData = []float64{0, 0}

	drawGraph := func(valsToPlot []float64) {
		p := widgets.NewPlot()
		p.Title = title
		p.Data = make([][]float64, 1)
		p.Data[0] = valsToPlot
		p.SetRect(leftPadding, topPadding, rightPadding, bottomPadding)
		p.TitleStyle.Fg = ui.ColorRed
		p.AxesColor = ui.ColorWhite
		p.LineColors[0] = ui.ColorGreen
		p.PlotType = widgets.LineChart

		ui.Render(p)
	}

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-ticker:
			// read from file
			data, currentPos, err := ReadFile(fileToRead, maxLinesToRead, startFrom, rePattern)
			// set startFrom for next iteration of read
			startFrom = currentPos
			if errors.Is(err, os.ErrNotExist) {
				// TODO: ignore if file does not exist, for now
			} else if err != nil {
				log.Fatalf("Error: %v", err)
			}

			// append data for plotting if matches were found
			if len(data) > 0 {
				for i := 0; i < len(data); i++ {
					plotData = append(plotData, data[i])
				}
				// reduce data size by removing old data to fit into the plot area for visibility
				if len(plotData) > rightPadding-10 {
					plotData = plotData[len(plotData)-(rightPadding-10):]
				}
			}

			drawGraph(plotData)
		}
	}
}
