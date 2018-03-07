package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	ui "github.com/gizak/termui"
	table "github.com/miguelmota/cryptodash/cryptodash/table"
	cmc "github.com/miguelmota/go-coinmarketcap"
)

// RenderChartDash renders chart dash
func RenderChartDash(coin string, dateRange string, color string, lineChartHeight uint) error {
	if coin == "" {
		coin = "bitcoin"
	}

	if dateRange == "" {
		dateRange = "7d"
	}

	primaryColor := getColor(color)

	var (
		oneMinute int64 = 60
		oneHour         = oneMinute * 60
		oneDay          = oneHour * 24
		oneWeek         = oneDay * 7
		oneMonth        = oneDay * 30
		oneYear         = oneDay * 365
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

	coinInfo, err := cmc.GetCoinData(coin)

	if err != nil {
		return err
	}

	graphData, err := cmc.GetCoinGraphData(coin, start, end)

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

	if lineChartHeight == 0 {
		lineChartHeight = 20
	}

	lc1 := ui.NewLineChart()
	lc1.Data = sinps
	lc1.Width = 100
	lc1.Height = int(lineChartHeight)
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
	par9.BorderLabel = "Circulating Supply"
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

// RenderGlobalMarketDash renders global market dash
func RenderGlobalMarketDash(color string) error {
	primaryColor := getColor(color)

	marketData, err := cmc.GetMarketData()

	if err != nil {
		return err
	}

	par0 := ui.NewPar(fmt.Sprintf("$%s", humanize.Commaf(marketData.TotalMarketCapUsd)))
	par0.Height = 3
	par0.Width = 20
	par0.Y = 1
	par0.TextFgColor = ui.ColorWhite
	par0.BorderLabel = "Total Market Cap (USD)"
	par0.BorderLabelFg = primaryColor
	par0.BorderFg = primaryColor

	par1 := ui.NewPar(fmt.Sprintf("$%s", humanize.Commaf(marketData.Total24hVolumeUsd)))
	par1.Height = 3
	par1.Width = 20
	par1.Y = 1
	par1.TextFgColor = ui.ColorWhite
	par1.BorderLabel = "Total Volume (24H)"
	par1.BorderLabelFg = primaryColor
	par1.BorderFg = primaryColor

	par2 := ui.NewPar(fmt.Sprintf("%.2f%%", marketData.BitcoinPercentageOfMarketCap))
	par2.Height = 3
	par2.Width = 20
	par2.Y = 1
	par2.TextFgColor = ui.ColorWhite
	par2.BorderLabel = "% Bitcoin Dominance"
	par2.BorderLabelFg = primaryColor
	par2.BorderFg = primaryColor

	par3 := ui.NewPar(fmt.Sprintf("%s", humanize.Comma(int64(marketData.ActiveCurrencies))))
	par3.Height = 3
	par3.Width = 20
	par3.Y = 1
	par3.TextFgColor = ui.ColorWhite
	par3.BorderLabel = "Active Currencies"
	par3.BorderLabelFg = primaryColor
	par3.BorderFg = primaryColor

	par4 := ui.NewPar(fmt.Sprintf("%s", humanize.Comma(int64(marketData.ActiveAssets))))
	par4.Height = 3
	par4.Width = 20
	par4.Y = 1
	par4.TextFgColor = ui.ColorWhite
	par4.BorderLabel = "Active Assets"
	par4.BorderLabelFg = primaryColor
	par4.BorderFg = primaryColor

	par5 := ui.NewPar(fmt.Sprintf("%s", humanize.Comma(int64(marketData.ActiveMarkets))))
	par5.Height = 3
	par5.Width = 20
	par5.Y = 1
	par5.TextFgColor = ui.ColorWhite
	par5.BorderLabel = "Active Markets"
	par5.BorderLabelFg = primaryColor
	par5.BorderFg = primaryColor

	// reset
	ui.Body.Rows = ui.Body.Rows[:0]

	// add grid rows and columns
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(2, 0, par0),
			ui.NewCol(2, 0, par1),
			ui.NewCol(2, 0, par2),
			ui.NewCol(2, 0, par3),
			ui.NewCol(2, 0, par4),
			ui.NewCol(2, 0, par5),
		),
	)

	// calculate layout
	ui.Body.Align()

	// render to terminal
	ui.Render(ui.Body)

	return nil
}

// RenderTable renders table
func RenderTable(color string, limit uint, refresh uint) error {
	t := table.New(&table.Options{
		Color:   color,
		Limit:   limit,
		Refresh: refresh,
	})
	return t.Render()
}

// GetColor gets primary color
func getColor(color string) ui.Attribute {
	if color == "" {
		color = "green"
	}

	primaryColor := ui.ColorGreen

	if color == "green" {
		primaryColor = ui.ColorGreen
	} else if color == "cyan" || color == "blue" {
		primaryColor = ui.ColorCyan
	} else if color == "magenta" || color == "pink" || color == "purple" {
		primaryColor = ui.ColorMagenta
	} else if color == "white" {
		primaryColor = ui.ColorWhite
	} else if color == "red" {
		primaryColor = ui.ColorRed
	} else if color == "yellow" || color == "orange" {
		primaryColor = ui.ColorYellow
	}

	return primaryColor
}

func main() {
	var coin = flag.String("coin", "bitcoin", "Cryptocurrency name. ie. bitcoin | ethereum | litecoin | etc...")
	var dateRange = flag.String("date", "7d", "Chart date range. ie. 1h | 1d | 2d | 7d | 30d | 2w | 1m | 3m | 1y")
	var color = flag.String("color", "green", "Primary color. ie. green | cyan | magenta | red | yellow | white")
	var lineChartHeight = flag.Uint("chart-height", 20, "Line chart height: .ie. 15 | 20 | 25 | 30")
	var showTable = flag.Bool("table", false, "Show the top 50 cryptocurrencies in a table.")
	var limit = flag.Uint("limit", 100, "Limit number of cryptocurrencies to return for table. ie. 10 | 25 | 50 | 100")
	var refresh = flag.Uint("refresh", 60, "How often to refetch data in seconds: .ie. 30, 60")
	var showGlobalMarketDash = flag.Bool("global", false, "Show global market data.")

	flag.Parse()

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	if *refresh == 0 {
		var i uint = 60
		refresh = &i
	}

	if *showGlobalMarketDash {
		err = RenderGlobalMarketDash(*color)
	} else if *showTable {
		for {
			err = RenderTable(*color, *limit, *refresh)
			if err != nil {
				panic(err)
			} else {
				return
			}
		}
	} else {
		//err = RenderChartDash(*coin, *dateRange, *color, *lineChartHeight)
		for {
			err = RenderTable(*color, *limit, *refresh)
			if err != nil {
				panic(err)
			} else {
				return
			}
		}
	}

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

	// quit on q
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	// refresh every minute
	ticker := time.NewTicker(time.Duration(int64(*refresh)) * time.Minute)

	// routine
	go func() {
	RESTART:
		for range ticker.C {
			var err error

			if *showGlobalMarketDash {
				err = RenderGlobalMarketDash(*color)
			} else {
				err = RenderChartDash(*coin, *dateRange, *color, *lineChartHeight)
			}

			if err != nil {
				goto RESTART
			}
		}
	}()

	ui.Loop()
}
