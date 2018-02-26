package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/bradfitz/slice"
	"github.com/fatih/color"
	cmc "github.com/miguelmota/go-coinmarketcap"
	gc "github.com/rgburke/goncurses"
	pad "github.com/willf/pad/utf8"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var wg sync.WaitGroup

// Service service struct
type Service struct {
	screenRows int
	screenCols int
	mainwin    *gc.Window
	menuwindow *gc.Window
	logwin     *gc.Window
	menu       *gc.Menu
	menuItems  []*gc.MenuItem
	menuData   []string
	menuHeader string
	coins      []*cmc.Coin
	sortBy     string
	sortDesc   bool
}

// Options options struct
type Options struct {
}

// New returns new service
func New(opts *Options) *Service {
	return &Service{}
}

// Start starts GUI
func (s *Service) Start() {
	stdsrc, err := gc.Init()
	defer gc.End()
	if err != nil {
		log.Fatal(err)
	}
	gc.StartColor()
	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)
	stdsrc.Keypad(true)
	s.setColorPairs()
	cols, rows := GetScreenSize()
	s.screenRows = rows
	s.screenCols = cols

	s.renderMainWindow()
	_ = gc.NewPanel(s.mainwin)
	gc.UpdatePanels()
	gc.Update()

	coins, err := cmc.GetAllCoinData(100)
	if err != nil {
		log.Fatal(err)
	}

	for i := range coins {
		coin := coins[i]
		s.coins = append(s.coins, &coin)
	}

	s.sortBy = "rank"
	s.sortDesc = false
	s.setMenuData()
	s.renderMenu()
	defer s.menu.UnPost()

	wg.Add(1)
	resizeChannel := make(chan os.Signal)
	signal.Notify(resizeChannel, syscall.SIGWINCH)
	go s.onWindowResize(resizeChannel)
	s.renderLogWindow()
	s.log("Use <up/down> arrows to navigate. <q> to exit. <F> keys to sort. <Enter> to visit coin on CoinMarketCap.")

	//stdsrc.GetChar() // required so it doesn't exit
	//wg.Wait()
	for {
		gc.Update()
		ch := s.menuwindow.GetChar()
		switch ch {
		case gc.KEY_RETURN, gc.KEY_ENTER:
			s.menu.Driver(gc.REQ_TOGGLE)
			for _, item := range s.menu.Items() {
				if item.Value() {
					s.handleClick(item.Index())
					break
				}
			}
			s.menu.Driver(gc.REQ_TOGGLE)
		case gc.KEY_F1:
			s.handleSort("rank", false)
		case gc.KEY_F2:
			s.handleSort("name", true)
		case gc.KEY_F3:
			s.handleSort("symbol", false)
		case gc.KEY_F4:
			s.handleSort("price", true)
		case gc.KEY_F5:
			s.handleSort("marketcap", true)
		case gc.KEY_F6:
			s.handleSort("24hvolume", true)
		case gc.KEY_F7:
			s.handleSort("1hchange", true)
		case gc.KEY_F8:
			s.handleSort("24hchange", true)
		case gc.KEY_F9:
			s.handleSort("7dchange", true)
		case gc.KEY_F10:
			s.handleSort("totalsupply", true)
		case gc.KEY_F11:
			s.handleSort("availablesupply", true)
		case gc.KEY_F12:
			s.handleSort("lastupdated", true)
		case 'q':
			return
		default:
			s.menu.Driver(gc.DriverActions[ch])
		}
	}
}

func (s *Service) handleClick(idx int) {
	slug := strings.ToLower(strings.Replace(s.coins[idx].Name, " ", "-", -1))
	exec.Command("open", fmt.Sprintf("https://coinmarketcap.com/currencies/%s", slug)).Output()
}

func (s *Service) handleSort(name string, desc bool) {
	if s.sortBy == name {
		s.sortDesc = !s.sortDesc
	} else {
		s.sortBy = name
		s.sortDesc = desc
	}
	s.setMenuData()
	s.renderMenu()
}

func (s *Service) setMenuData() {
	slice.Sort(s.coins[:], func(i, j int) bool {
		if s.sortDesc == true {
			i, j = j, i
		}
		switch s.sortBy {
		case "rank":
			return s.coins[i].Rank < s.coins[j].Rank
		case "name":
			return s.coins[i].Name < s.coins[j].Name
		case "symbol":
			return s.coins[i].Symbol < s.coins[j].Symbol
		case "price":
			return s.coins[i].PriceUsd < s.coins[j].PriceUsd
		case "marketcap":
			return s.coins[i].MarketCapUsd < s.coins[j].MarketCapUsd
		case "24hvolume":
			return s.coins[i].Usd24hVolume < s.coins[j].Usd24hVolume
		case "1hchange":
			return s.coins[i].PercentChange1h < s.coins[j].PercentChange1h
		case "24hchange":
			return s.coins[i].PercentChange24h < s.coins[j].PercentChange24h
		case "7dchange":
			return s.coins[i].PercentChange7d < s.coins[j].PercentChange7d
		case "totalsupply":
			return s.coins[i].TotalSupply < s.coins[j].TotalSupply
		case "availablesupply":
			return s.coins[i].AvailableSupply < s.coins[j].AvailableSupply
		case "lastupdated":
			return s.coins[i].LastUpdated < s.coins[j].LastUpdated
		default:
			return s.coins[i].Rank < s.coins[j].Rank
		}
	})

	var menuData []string
	for _, coin := range s.coins {
		unix, _ := strconv.ParseInt(coin.LastUpdated, 10, 64)
		lastUpdated := time.Unix(unix, 0).Format("15:04:05 Jan 02")
		fields := []string{
			pad.Right(fmt.Sprint(coin.Rank), 4, " "),
			pad.Right(coin.Name, 22, " "),
			pad.Right(coin.Symbol, 6, " "),
			pad.Left(humanize.Commaf(coin.PriceUsd), 12, " "),
			pad.Left(humanize.Commaf(coin.MarketCapUsd), 17, " "),
			pad.Left(humanize.Commaf(coin.Usd24hVolume), 15, " "),
			pad.Left(fmt.Sprintf("%.2f%%", coin.PercentChange1h), 9, " "),
			pad.Left(fmt.Sprintf("%.2f%%", coin.PercentChange24h), 9, " "),
			pad.Left(fmt.Sprintf("%.2f%%", coin.PercentChange7d), 9, " "),
			pad.Left(humanize.Commaf(coin.TotalSupply), 20, " "),
			pad.Left(humanize.Commaf(coin.AvailableSupply), 18, " "),
			pad.Left(fmt.Sprintf("%s", lastUpdated), 18, " "),
			// add %percent of cap
		}
		var str string
		for _, f := range fields {
			str = fmt.Sprintf("%s%s", str, f)
		}
		menuData = append(menuData, str)
	}

	s.menuData = menuData

	headers := []string{
		pad.Right("Rank", 6, " "),
		pad.Right("Name", 20, " "),
		pad.Right("Symbol", 10, " "),
		pad.Right("Price", 10, " "),
		pad.Right("Market Cap", 17, " "),
		pad.Right("24 Volume", 16, " "),
		pad.Right("1H%", 9, " "),
		pad.Right("24H%", 9, " "),
		pad.Right("7D%", 8, " "),
		pad.Right("Total Supply", 20, " "),
		pad.Right("Available Supply", 19, " "),
		pad.Right("Last Updated", 16, " "),
	}
	header := "   "
	for _, h := range headers {
		header = fmt.Sprintf("%s%s", header, h)
	}

	s.menuHeader = header
}

// SetColorPairs sets color pairs
func (s *Service) setColorPairs() {
	gc.InitPair(1, gc.C_RED, gc.C_BLACK)
	gc.InitPair(2, gc.C_CYAN, gc.C_BLACK)
	gc.InitPair(3, gc.C_WHITE, gc.C_BLACK)
	gc.InitPair(4, gc.C_YELLOW, gc.C_BLACK)
	gc.InitPair(5, gc.C_BLACK, gc.C_BLACK)
	gc.InitPair(6, gc.C_BLACK, gc.C_GREEN)
	gc.InitPair(7, gc.C_BLACK, gc.C_CYAN)
}

// RenderMainWindow renders main window
func (s *Service) renderMainWindow() {
	if s.mainwin == nil {
		var err error
		s.mainwin, err = gc.NewWindow(s.screenRows, s.screenCols, 0, 0)
		if err != nil {
			log.Fatal(err)
		}
	}
	s.mainwin.Clear()
	s.mainwin.ColorOn(5)
	s.mainwin.Resize(s.screenRows, s.screenCols)
	s.mainwin.Box(0, 0)
	s.mainwin.Refresh()
}

// ResizeWindows resizes windows
func (s *Service) resizeWindows() {
	s.renderMainWindow()
	s.renderMenu()
	s.renderLogWindow()
	s.log(fmt.Sprintf("%v %v %v", time.Now().Unix(), s.screenCols, s.screenRows))
}

func (s *Service) renderLogWindow() {
	var err error
	if s.logwin == nil {
		s.logwin, err = gc.NewWindow(3, s.screenCols-4, s.screenRows-4, 2)
		if err != nil {
			log.Fatal(err)
		}
	}
	s.logwin.Clear()
	s.logwin.Resize(3, s.screenCols-4)
	s.logwin.MoveWindow(2, 30)
	s.logwin.ColorOn(3)
	s.logwin.Box(0, 0)
	s.logwin.Refresh()
}

// Log logs debug messages
func (s *Service) log(msg string) {
	s.logwin.Clear()
	s.logwin.ColorOn(3)
	s.logwin.Box(0, 0)
	s.logwin.MovePrint(1, 1, msg)
	s.logwin.Refresh()
}

// OnWindowResize sends event to channel when resize event occurs
func (s *Service) onWindowResize(channel chan os.Signal) {
	stdScr, _ := gc.Init()
	stdScr.ScrollOk(true)
	gc.NewLines(true)
	for {
		<-channel
		//gc.StdScr().Clear()
		//y, x := gc.StdScr().MaxYX()
		cols, rows := GetScreenSize()
		s.screenRows = rows
		s.screenCols = cols
		s.resizeWindows()
		//gc.End()
		//gc.Update()
		//gc.StdScr().Refresh()
	}
}

// RenderMenu renders menu
func (s *Service) renderMenu() {
	//if len(s.menuItems) == 0 {
	items := make([]*gc.MenuItem, len(s.menuData))
	var err error
	for i, val := range s.menuData {
		items[i], err = gc.NewItem(val, "")
		if err != nil {
			log.Fatal(err)
		}
		//defer items[i].Free()
	}

	s.menuItems = items
	//}

	//if s.menu == nil {
	//		var err error
	s.menu, err = gc.NewMenu(s.menuItems)
	if err != nil {
		log.Fatal(err)
	}
	//} else {
	//		s.menu.SetItems(s.menuItems)
	//	}

	//if s.menuwindow == nil {
	//var err error
	s.menuwindow, err = gc.NewWindow(s.screenRows-6, s.screenCols-4, 2, 2)
	s.menuwindow.ScrollOk(true)
	if err != nil {
		log.Fatal(err)
	}

	s.menuwindow.Keypad(true)
	s.menu.SetWindow(s.menuwindow)
	dwin := s.menuwindow.Derived(s.screenRows-10, s.screenCols-10, 3, 1)
	s.menu.SubWindow(dwin)
	s.menu.Option(gc.O_ONEVALUE, false)
	s.menu.Format(s.screenRows-10, 1)
	s.menu.Mark(" * ")
	//} else {
	//	s.menuwindow.Resize(s.screenRows-6, s.screenCols-4)
	//}

	s.menuwindow.Clear()
	s.menuwindow.Box(0, 0)
	s.menuwindow.ColorOn(6)
	s.menuwindow.MovePrint(1, 1, s.menuHeader)
	s.menuwindow.ColorOff(6)
	s.menuwindow.MoveAddChar(2, 0, gc.ACS_LTEE)
	s.menuwindow.HLine(2, 1, gc.ACS_HLINE, s.screenCols-6)
	s.menu.Post()
	s.menuwindow.Refresh()
}

func main() {
	service := New(&Options{})
	service.Start()
}
