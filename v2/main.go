package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
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
	coins      []*cmc.Coin
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

	var menuData []string
	for i := range coins {
		coin := coins[i]
		s.coins = append(s.coins, &coin)
	}

	slice.Sort(s.coins[:], func(i, j int) bool {
		return s.coins[i].Rank < s.coins[j].Rank
	})

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

	s.renderMenu()
	defer s.menu.UnPost()

	wg.Add(1)
	resizeChannel := make(chan os.Signal)
	signal.Notify(resizeChannel, syscall.SIGWINCH)
	go s.onWindowResize(resizeChannel)
	s.renderLogWindow()
	s.log("Use up/down arrows to navigate. 'q' to exit")

	//stdsrc.GetChar() // required so it doesn't exit
	//wg.Wait()
	for {
		gc.Update()
		switch ch := s.menuwindow.GetChar(); ch {
		case gc.KEY_RETURN, gc.KEY_ENTER:
			s.menu.Driver(gc.REQ_TOGGLE)
			for _, item := range s.menu.Items() {
				if item.Value() {
					coin := s.coins[item.Index()]
					exec.Command("open", fmt.Sprintf("https://coinmarketcap.com/currencies/%s", coin.Name)).Output()
					break
				}
			}
			s.menu.Driver(gc.REQ_TOGGLE)
		case 'q':
			return
		default:
			s.menu.Driver(gc.DriverActions[ch])
		}
	}
}

// SetColorPairs sets color pairs
func (s *Service) setColorPairs() {
	gc.InitPair(1, gc.C_RED, gc.C_BLACK)
	gc.InitPair(2, gc.C_CYAN, gc.C_BLACK)
	gc.InitPair(3, gc.C_WHITE, gc.C_BLACK)
	gc.InitPair(4, gc.C_YELLOW, gc.C_BLACK)
	gc.InitPair(5, gc.C_BLACK, gc.C_BLACK)
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
	if len(s.menuItems) == 0 {
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
	}

	if s.menu == nil {
		var err error
		s.menu, err = gc.NewMenu(s.menuItems)
		if err != nil {
			log.Fatal(err)
		}
	}

	if s.menuwindow == nil {
		var err error
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
	} else {
		s.menuwindow.Resize(s.screenRows-6, s.screenCols-4)
	}

	title := "CoinMarketCap"
	s.menuwindow.Clear()
	s.menuwindow.Box(0, 0)
	s.menuwindow.ColorOn(3)
	s.menuwindow.MovePrint(1, 1, title)
	s.menuwindow.ColorOff(3)
	s.menuwindow.MoveAddChar(2, 0, gc.ACS_LTEE)
	s.menuwindow.HLine(2, 1, gc.ACS_HLINE, s.screenCols-6)
	s.menu.Post()
	s.menuwindow.Refresh()
}

func main() {
	service := New(&Options{})
	service.Start()
}
