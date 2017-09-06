package main

import (
  "fmt"
  "time"
  "os"
  "strconv"

  ui "github.com/gizak/termui"
  coinApi "github.com/miguelmota/go-coinmarketcap"
  "github.com/dustin/go-humanize"
)

func FloatToString(input_num float64) string {
    // to convert a float number to a string
    return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func Render(coin string, dateRange string) {
  err := ui.Init()
  if err != nil {
    panic(err)
  }
  defer ui.Close()

  if coin == "" {
    coin = "bitcoin"
  }

  if dateRange == "" {
    dateRange = "30d"
  }

  var oneYear int64 = (59 * 60 * 24 * 365)
  var threeMonths int64 = (59 * 60 * 24 * 90)
  var oneMonth int64 = (60 * 60 * 24 * 30)
  var oneWeek int64 = (60 * 60 * 24 * 7)
  var oneDay int64 = (60 * 60 * 24)
  now := time.Now()
  secs := now.Unix()
  start := secs - oneMonth
  end := secs

  if dateRange == "1d" {
    start = secs - oneDay
  } else if dateRange == "7d" {
    start = secs - oneWeek
  } else if dateRange == "30d" {
    start = secs - oneMonth
  } else if dateRange == "90d" {
    start = secs - threeMonths
  } else if dateRange == "1y" {
    start = secs - oneYear
  }

  coinInfo, _ := coinApi.GetCoinData(coin)
  graphData, _ := coinApi.GetCoinGraphData(coin, start, end)

  sinps := (func() []float64 {
    n := len(graphData.PriceUsd)
    ps := make([]float64, n)
    for i := range graphData.PriceUsd {
      ps[i] = graphData.PriceUsd[i][1]
    }
    return ps
  })()

  rows1 := [][]string{
    []string{"Name", "Symbol", "Price", "Market Cap", "24h Volume", "Available Supply", "Total Supply"},
    []string{coinInfo.Name, coinInfo.Symbol, humanize.Commaf(coinInfo.PriceUsd), humanize.Commaf(coinInfo.MarketCapUsd), humanize.Commaf(coinInfo.Usd24hVolume), humanize.Commaf(coinInfo.AvailableSupply), humanize.Commaf(coinInfo.TotalSupply)},
  }

  table1 := ui.NewTable()
  table1.Rows = rows1
  table1.FgColor = ui.ColorGreen
  table1.BgColor = ui.ColorDefault
  table1.BorderFg = ui.ColorGreen
  table1.Y = 0
  table1.X = 0
  table1.Width = 100
  table1.Height = 5

  chartTitle := "Price History"
  lc2 := ui.NewLineChart()
  lc2.BorderLabel = fmt.Sprintf("%s: %s", chartTitle, dateRange)
  lc2.Mode = "dot"
  lc2.Data = sinps[4:]
  lc2.Width = 100
  lc2.Height = 16
  lc2.X = 0
  lc2.Y = 7
  lc2.AxesColor = ui.ColorGreen
  lc2.LineColor = ui.ColorGreen | ui.AttrBold
  lc2.BorderFg = ui.ColorGreen

  par0 := ui.NewPar(fmt.Sprintf("%f %%", coinInfo.PercentChange1h))
  par0.Height = 3
  par0.Width = 20
  par0.Y = 1
  par0.TextFgColor = ui.ColorGreen
  par0.BorderLabel = "1h ▲"
  par0.BorderFg = ui.ColorGreen

  par1 := ui.NewPar(fmt.Sprintf("%f%%", coinInfo.PercentChange24h))
  par1.Height = 3
  par1.Width = 20
  par1.Y = 1
  par1.TextFgColor = ui.ColorGreen
  par1.BorderLabel = "24h ▲"
  par1.BorderFg = ui.ColorGreen

  par2 := ui.NewPar(fmt.Sprintf("%f%%", coinInfo.PercentChange7d))
  par2.Height = 3
  par2.Width = 20
  par2.Y = 1
  par2.TextFgColor = ui.ColorGreen
  par2.BorderLabel = "7d ▲"
  par2.BorderFg = ui.ColorGreen

  ui.Body.AddRows(
    ui.NewRow(
      ui.NewCol(12, 0, table1),
    ),
    ui.NewRow(
      ui.NewCol(4, 0, par0),
      ui.NewCol(4, 0, par1),
      ui.NewCol(4, 0, par2),
    ),
    ui.NewRow(
      ui.NewCol(12, 0, lc2),
    ),
  )

  // calculate layout
  ui.Body.Align()

  ui.Render(ui.Body)

  ui.Handle("/sys/kbd/q", func(ui.Event) {
    ui.StopLoop()
  })
  ui.Loop()
}

func main() {
  coin := ""
  dateRange := ""

  argsWithoutProg := os.Args[1:]

  if len(argsWithoutProg) > 0 {
    coin = argsWithoutProg[0]
  }

  if len(argsWithoutProg) > 1 {
    dateRange = argsWithoutProg[1]
  }

  Render(coin, dateRange)
}
