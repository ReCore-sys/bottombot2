package stocks

import (
	"fmt"
	"os"

	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/wcharczuk/go-chart"
)

func CreateGraph(prices []map[string]float64) {
	tickerprices := make(map[string][]float64)
	for _, price := range prices {
		for k, v := range price {
			tickerprices[k] = append(tickerprices[k], v)
		}
	}
	var line []chart.Series

	for ticker := range tickerprices {
		var intlist []float64 // Create a list of all numbers up to the length of the longest ticker
		for i := 0; i < len(tickerprices[ticker]); {
			intlist = append(intlist, float64(i))
			i++
		}
		line = append(line, chart.ContinuousSeries{
			XValues: intlist,
			YValues: tickerprices[ticker],
			Name:    ticker,
			Style: chart.Style{
				Show: true,
			},
		})
	}
	graph := chart.Chart{
		TitleStyle: chart.Style{
			Show: true,
		},
		ColorPalette: nil,
		Width:        1000,
		Height:       400,
		DPI:          0,
		Background: chart.Style{
			Show: true,
			Padding: chart.Box{
				Left: 100,
			},
		},
		YAxis: chart.YAxis{
			Name: "Price",
			Style: chart.Style{
				Show: true,
			},
			NameStyle: chart.Style{
				Show: true,
			},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("$%.2f", v)
			},
		},
		Title:  "Stock prices",
		Series: line,
	}
	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}
	f, err := os.OpenFile("graph.png", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logging.Log(err)
	}
	err = graph.Render(chart.PNG, f)
	if err != nil {
		logging.Log(err)
	}
}
