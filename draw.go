package main

import (
	"errors"
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Graphs struct {
	ConfigName    string
	Title         string
	StartFrom     int64
	Values        []float64
	LeftPadding   int
	TopPadding    int
	RightPadding  int
	BottomPadding int
}

type DataResp struct {
	ConfigName string
	Data       []float64
	Position   int64
	FileName   string
	Error      error
}

// func DrawGraph(fileToRead string, title string, padding []int, maxLinesToRead int, startFrom int64, rePattern string) {
func DrawGraph(configData map[string]Configs, graphsMap map[string]Graphs) {

	ch := make(chan DataResp)

	// initialize termui
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(1 * time.Second).C

	// call plot on ticker
	plot := func(graph map[string]Graphs) {
		drawables := make([]ui.Drawable, 0, len(graph))
		for _, g := range graph {
			p := widgets.NewPlot()
			p.Title = g.Title
			p.Data = make([][]float64, 1)
			p.Data[0] = g.Values
			p.SetRect(g.LeftPadding, g.TopPadding, g.RightPadding, g.BottomPadding)
			p.TitleStyle.Fg = ui.ColorGreen
			p.AxesColor = ui.ColorWhite
			p.LineColors[0] = ui.ColorRed
			p.PlotType = widgets.LineChart
			drawables = append(drawables, p)
		}
		ui.Render(drawables...)
	}

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-ticker:
			for k, v := range configData {
				go ReadFile(v.FilePath, v.MaxLines, graphsMap[k].StartFrom, v.RegexPattern, k, ch)
			}

			for range configData {
				dataResp := <-ch

				var graph Graphs

				if errors.Is(dataResp.Error, os.ErrNotExist) {
					// TODO: ignore if file does not exist, for now
				} else if dataResp.Error != nil {
					log.Fatalf("Error: %v", dataResp.Error)
				}

				// reset StartFrom to the data position read so far
				graph.StartFrom = dataResp.Position

				// append data for plotting if matches were found
				graph.Values = graphsMap[dataResp.ConfigName].Values
				if len(dataResp.Data) > 0 {
					graph.Values = append(graph.Values, dataResp.Data...)

					// reduce data size by removing old data to fit into the plot area for visibility
					visiblePoints := (graphsMap[dataResp.ConfigName].RightPadding - 8) - graphsMap[dataResp.ConfigName].LeftPadding
					if len(graph.Values) > visiblePoints {
						graph.Values = graph.Values[(len(graph.Values) - visiblePoints):]
					}
				}

				graph.Title = graphsMap[dataResp.ConfigName].Title
				graph.LeftPadding = graphsMap[dataResp.ConfigName].LeftPadding
				graph.TopPadding = graphsMap[dataResp.ConfigName].TopPadding
				graph.RightPadding = graphsMap[dataResp.ConfigName].RightPadding
				graph.BottomPadding = graphsMap[dataResp.ConfigName].BottomPadding
				graphsMap[dataResp.ConfigName] = graph
			}

			plot(graphsMap)
		}
	}
}
