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
	GraphId       string
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
	GraphId    string
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

		p := widgets.NewPlot()
		q := widgets.NewPlot()

		for k := range graph {

			switch graph[k].GraphId {
			case "graph1":
				p.Title = graph[k].Title
				p.Data = make([][]float64, 1)
				p.Data[0] = graph[k].Values
				p.SetRect(graph[k].LeftPadding, graph[k].TopPadding, graph[k].RightPadding, graph[k].BottomPadding)
				p.TitleStyle.Fg = ui.ColorGreen
				p.AxesColor = ui.ColorWhite
				p.LineColors[0] = ui.ColorRed
				p.PlotType = widgets.LineChart

			case "graph2":
				q.Title = graph[k].Title
				q.Data = make([][]float64, 1)
				q.Data[0] = graph[k].Values
				q.SetRect(graph[k].LeftPadding, graph[k].TopPadding, graph[k].RightPadding, graph[k].BottomPadding)
				q.TitleStyle.Fg = ui.ColorGreen
				q.AxesColor = ui.ColorWhite
				q.LineColors[0] = ui.ColorRed
				q.PlotType = widgets.LineChart
			}
		}

		ui.Render(p, q)
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
				go ReadFile(graphsMap[k].GraphId, v.FilePath, v.MaxLines, graphsMap[k].StartFrom, v.RegexPattern, k, ch)
			}

			for range configData {
				dataResp := <-ch

				var graph Graphs
				id := dataResp.GraphId

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

				graph.GraphId = id
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
