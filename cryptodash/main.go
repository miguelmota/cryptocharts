package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	ui "github.com/gizak/termui"
	coinApi "github.com/miguelmota/go-coinmarketcap"
)

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func Render(coin string, dateRange string, color string) error {
	if coin == "" {
		coin = "bitcoin"
	}

	if dateRange == "" {
		dateRange = "7d"
	}

	if color == "" {
		color = "green"
	}

	primaryColor := ui.ColorGreen

	if color == "green" {
		primaryColor = ui.ColorGreen
	} else if color == "cyan" || color == "blue" {
		primaryColor = ui.ColorCyan
	} else if color == "magenta" || color == "pink" {
		primaryColor = ui.ColorMagenta
	} else if color == "white" {
		primaryColor = ui.ColorWhite
	} else if color == "red" {
		primaryColor = ui.ColorRed
	} else if color == "yellow" {
		primaryColor = ui.ColorYellow
	}

	var (
		oneMinute int64 = 60
		oneHour   int64 = oneMinute * 60
		oneDay    int64 = oneHour * 24
		oneWeek   int64 = oneDay * 7
		oneMonth  int64 = oneDay * 30
		oneYear   int64 = oneDay * 365
	)

	now := time.Now()
	secs := now.Unix()
	start := secs - oneDay
	end := secs

	dateNumber, err := strconv.ParseInt(dateRange[0:len(dateRange)-1], 10, 64)

	if err != nil {
		dateNumber = 30
	}

	dateType := dateRange[len(dateRange)-1:]

	if dateType == "n" {
		start = secs - (oneMinute * dateNumber)
	} else if dateType == "h" {
		start = secs - (oneHour * dateNumber)
	} else if dateType == "d" {
		start = secs - (oneDay * dateNumber)
	} else if dateType == "w" {
		start = secs - (oneWeek * dateNumber)
	} else if dateType == "m" {
		start = secs - (oneMonth * dateNumber)
	} else if dateType == "y" {
		start = secs - (oneYear * dateNumber)
	} else {
		dateType = "d"
	}

	coinInfo, err := coinApi.GetCoinData(coin)

	if err != nil {
		return err
	}

	graphData, err := coinApi.GetCoinGraphData(coin, start, end)

	if err != nil {
		return err
	}

	sinps := (func() []float64 {
		n := len(graphData.PriceUsd)
		ps := make([]float64, n)
		for i := range graphData.PriceUsd {
			ps[i] = graphData.PriceUsd[i][1]
		}
		return ps
	})()

	lc1 := ui.NewLineChart()
	lc1.Data = sinps
	lc1.Width = 100
	lc1.Height = 16
	lc1.X = 0
	lc1.Y = 7
	lc1.AxesColor = primaryColor
	lc1.LineColor = primaryColor | ui.AttrBold
	lc1.BorderFg = primaryColor
	lc1.BorderLabel = fmt.Sprintf("%s %s: %d%s", coinInfo.Symbol, "Price History", dateNumber, strings.ToUpper(dateType))
	lc1.BorderLabelFg = primaryColor

	par0 := ui.NewPar(fmt.Sprintf("%.2f%%", coinInfo.PercentChange1h))
	par0.Height = 3
	par0.Width = 20
	par0.Y = 1
	par0.TextFgColor = ui.ColorGreen
	par0.BorderLabel = "% Change (1H)"
	par0.BorderLabelFg = ui.ColorGreen
	par0.BorderFg = ui.ColorGreen
	if coinInfo.PercentChange1h < 0 {
		par0.TextFgColor = ui.ColorRed
		par0.BorderFg = ui.ColorRed
		par0.BorderLabelFg = ui.ColorRed
	}

	par1 := ui.NewPar(fmt.Sprintf("%.2f%%", coinInfo.PercentChange24h))
	par1.Height = 3
	par1.Width = 20
	par1.Y = 1
	par1.TextFgColor = ui.ColorGreen
	par1.BorderLabel = "% Change (24H)"
	par1.BorderFg = ui.ColorGreen
	if coinInfo.PercentChange24h < 0 {
		par1.TextFgColor = ui.ColorRed
		par1.BorderFg = ui.ColorRed
		par1.BorderLabelFg = ui.ColorRed
	}

	par2 := ui.NewPar(fmt.Sprintf("%.2f%%", coinInfo.PercentChange7d))
	par2.Height = 3
	par2.Width = 20
	par2.Y = 1
	par2.TextFgColor = ui.ColorGreen
	par2.BorderLabel = "% Change (7D)"
	par2.BorderFg = ui.ColorGreen
	if coinInfo.PercentChange7d < 0 {
		par2.TextFgColor = ui.ColorRed
		par2.BorderFg = ui.ColorRed
		par2.BorderLabelFg = ui.ColorRed
	}

	par3 := ui.NewPar(fmt.Sprintf("%s", coinInfo.Name))
	par3.Height = 3
	par3.Width = 20
	par3.Y = 1
	par3.TextFgColor = ui.ColorWhite
	par3.BorderLabel = "Name"
	par3.BorderLabelFg = primaryColor
	par3.BorderFg = primaryColor

	par4 := ui.NewPar(fmt.Sprintf("$%s", humanize.Commaf(coinInfo.PriceUsd)))
	par4.Height = 3
	par4.Width = 20
	par4.Y = 1
	par4.TextFgColor = ui.ColorWhite
	par4.BorderLabel = "Price (USD)"
	par4.BorderLabelFg = primaryColor
	par4.BorderFg = primaryColor

	par5 := ui.NewPar(fmt.Sprintf("%s", coinInfo.Symbol))
	par5.Height = 3
	par5.Width = 20
	par5.Y = 1
	par5.TextFgColor = ui.ColorWhite
	par5.BorderLabel = "Symbol"
	par5.BorderLabelFg = primaryColor
	par5.BorderFg = primaryColor

	par6 := ui.NewPar(humanize.Comma(int64(coinInfo.Rank)))
	par6.Height = 3
	par6.Width = 20
	par6.Y = 1
	par6.TextFgColor = ui.ColorWhite
	par6.BorderLabel = "Rank"
	par6.BorderLabelFg = primaryColor
	par6.BorderFg = primaryColor

	par7 := ui.NewPar(fmt.Sprintf("$%s", humanize.Commaf(coinInfo.MarketCapUsd)))
	par7.Height = 3
	par7.Width = 20
	par7.Y = 1
	par7.TextFgColor = ui.ColorWhite
	par7.BorderLabel = "Market Cap"
	par7.BorderLabelFg = primaryColor
	par7.BorderFg = primaryColor

	par8 := ui.NewPar(fmt.Sprintf("$%s", humanize.Commaf(coinInfo.Usd24hVolume)))
	par8.Height = 3
	par8.Width = 20
	par8.Y = 1
	par8.TextFgColor = ui.ColorWhite
	par8.BorderLabel = "Volume (24H)"
	par8.BorderLabelFg = primaryColor
	par8.BorderFg = primaryColor

	par9 := ui.NewPar(fmt.Sprintf("%s %s", humanize.Commaf(coinInfo.AvailableSupply), coinInfo.Symbol))
	par9.Height = 3
	par9.Width = 20
	par9.Y = 1
	par9.TextFgColor = ui.ColorWhite
	par9.BorderLabel = "Available Supply"
	par9.BorderLabelFg = primaryColor
	par9.BorderFg = primaryColor

	par10 := ui.NewPar(fmt.Sprintf("%s %s", humanize.Commaf(coinInfo.TotalSupply), coinInfo.Symbol))
	par10.Height = 3
	par10.Width = 20
	par10.Y = 1
	par10.TextFgColor = ui.ColorWhite
	par10.BorderLabel = "Total Supply"
	par10.BorderLabelFg = primaryColor
	par10.BorderFg = primaryColor

	unix, err := strconv.ParseInt(coinInfo.LastUpdated, 10, 64)

	if err != nil {
		return err
	}

	par11 := ui.NewPar(time.Unix(unix, 0).Format("15:04:05 Jan 02"))
	par11.Height = 3
	par11.Width = 20
	par11.Y = 1
	par11.TextFgColor = ui.ColorWhite
	par11.BorderLabel = "Last Updated"
	par11.BorderLabelFg = primaryColor
	par11.BorderFg = primaryColor

	// reset
	ui.Body.Rows = ui.Body.Rows[:0]

	// add grid rows and columns
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(2, 0, par3),
			ui.NewCol(2, 0, par5),
			ui.NewCol(2, 0, par4),
			ui.NewCol(2, 0, par0),
			ui.NewCol(2, 0, par1),
			ui.NewCol(2, 0, par2),
		),
		ui.NewRow(
			ui.NewCol(2, 0, par6),
			ui.NewCol(2, 0, par7),
			ui.NewCol(2, 0, par8),
			ui.NewCol(2, 0, par9),
			ui.NewCol(2, 0, par10),
			ui.NewCol(2, 0, par11),
		),
		ui.NewRow(
			ui.NewCol(12, 0, lc1),
		),
	)

	// calculate layout
	ui.Body.Align()

	// render to terminal
	ui.Render(ui.Body)

	return nil
}

func main() {
	coin := ""
	dateRange := ""
	color := ""

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {
		coin = argsWithoutProg[0]
	}

	if len(argsWithoutProg) > 1 {
		dateRange = argsWithoutProg[1]
	}

	if len(argsWithoutProg) > 2 {
		color = argsWithoutProg[2]
	}

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	err = Render(coin, dateRange, color)

	if err != nil {
		panic(err)
	}

	// re-adjust grid on window resize
	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	// quit on Ctrl-c
	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})

	// refresh every minute
	ticker := time.NewTicker(60 * time.Second)

	// routine
	go func() {
	RESTART:
		for range ticker.C {
			err := Render(coin, dateRange, color)
			if err != nil {
				goto RESTART
			}
		}
	}()

	ui.Loop()
}
