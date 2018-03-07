package table

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	slice "github.com/bradfitz/slice"
	humanize "github.com/dustin/go-humanize"
	cmc "github.com/miguelmota/go-coinmarketcap"
	gc "github.com/rgburke/goncurses"
	pad "github.com/willf/pad/utf8"
)

var wg sync.WaitGroup

// Service service struct
type Service struct {
	stdsrc        *gc.Window
	screenRows    int
	screenCols    int
	mainwin       *gc.Window
	menuwin       *gc.Window
	menuwinWidth  int
	menuwinHeight int
	menusubwin    *gc.Window
	helpbarwin    *gc.Window
	helpwin       *gc.Window
	helpVisible   bool
	logwin        *gc.Window
	menu          *gc.Menu
	menuItems     []*gc.MenuItem
	menuData      []string
	menuHeader    string
	menuWidth     int
	menuHeight    int
	coins         []*cmc.Coin
	sortBy        string
	sortDesc      bool
	limit         uint
	refresh       uint
	primaryColor  string
	lastLog       string
	currentItem   int
}

// Options options struct
type Options struct {
	Color   string
	Limit   uint
	Refresh uint
}

var once sync.Once

// New returns new service
func New(opts *Options) *Service {
	var instance *Service
	//	once.Do(func() {
	instance = &Service{}
	instance.primaryColor = opts.Color
	instance.limit = opts.Limit
	instance.refresh = opts.Refresh
	//	})

	return instance
}

// Render starts GUI
func (s *Service) Render() error {
	var err error
	s.stdsrc, err = gc.Init()
	defer gc.End()
	if err != nil {
		return err
	}
	gc.UseDefaultColors()
	gc.StartColor()
	s.setColorPairs()
	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)
	s.stdsrc.Keypad(true)
	cols, rows := GetScreenSize()
	s.screenRows = rows
	s.screenCols = cols
	s.helpVisible = false

	s.renderMainWindow()
	err = s.fetchData()
	if err != nil {
		return nil
	}

	go func() {
		ticker := time.NewTicker(time.Duration(int64(s.refresh)) * time.Minute)
		for {
			select {
			case <-ticker.C:
				//s.menuwin.Clear()
				//s.menuwin.Refresh()
				s.fetchData()
				s.setMenuData()
				err := s.renderMenu()
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	s.sortBy = "rank"
	s.sortDesc = false
	s.setMenuData()
	err = s.renderMenu()
	if err != nil {
		panic(err)
	}
	defer s.menu.UnPost()

	wg.Add(1)
	resizeChannel := make(chan os.Signal)
	signal.Notify(resizeChannel, syscall.SIGWINCH)
	go s.onWindowResize(resizeChannel)
	s.renderLogWindow()
	s.renderHelpBar()
	s.renderHelpWindow()

	//stdsrc.GetChar() // required so it doesn't exit
	//wg.Wait()

	for {
		gc.Update()
		ch := s.menuwin.GetChar()
		chstr := fmt.Sprint(ch)
		//s.log(fmt.Sprint(ch))
		switch {
		case ch == gc.KEY_DOWN, chstr == "106": // "j"
			if s.currentItem < len(s.menuItems)-1 {
				s.currentItem = s.currentItem + 1
				s.menu.Current(s.menuItems[s.currentItem])
			}
		case ch == gc.KEY_UP, chstr == "107": // "k"
			if s.currentItem > 0 {
				s.currentItem = s.currentItem - 1
				s.menu.Current(s.menuItems[s.currentItem])
			}
		case ch == gc.KEY_RETURN, ch == gc.KEY_ENTER, chstr == "32":
			s.menu.Driver(gc.REQ_TOGGLE)
			for _, item := range s.menu.Items() {
				if item.Value() {
					s.handleClick(item.Index())
					break
				}
			}
			s.menu.Driver(gc.REQ_TOGGLE)
		case chstr == "114": // "r"
			s.handleSort("rank", false)
		case chstr == "110": // "n"
			s.handleSort("name", true)
		case chstr == "115": // "s"
			s.handleSort("symbol", false)
		case chstr == "112": // "p
			s.handleSort("price", true)
		case chstr == "109": // "m
			s.handleSort("marketcap", true)
		case chstr == "118": // "v
			s.handleSort("24hvolume", true)
		case chstr == "49": // "1"
			s.handleSort("1hchange", true)
		case chstr == "50": // "2"
			s.handleSort("24hchange", true)
		case chstr == "55": // "7"
			s.handleSort("7dchange", true)
		case chstr == "116": // "t"
			s.handleSort("totalsupply", true)
		case chstr == "97": // "a"
			s.handleSort("availablesupply", true)
		case chstr == "108": // "l"
			s.handleSort("lastupdated", true)
		case chstr == "21": // ctrl-u
			s.currentItem = s.currentItem - s.menuHeight
			if s.currentItem < 0 {
				s.currentItem = 0
			}
			if s.currentItem >= len(s.menuItems) {
				s.currentItem = len(s.menuItems) - 1
			}
			//s.log(fmt.Sprintf("%v %v", s.currentItem, s.screenRows))
			s.menu.Current(s.menuItems[s.currentItem])
		case fmt.Sprint(ch) == "4": // ctrl-d
			s.currentItem = s.currentItem + s.menuHeight
			if s.currentItem < 0 {
				s.currentItem = 0
			}
			if s.currentItem >= len(s.menuItems) {
				s.currentItem = len(s.menuItems) - 1
			}
			//s.log(fmt.Sprintf("%v %v", s.currentItem, s.screenRows))
			s.menu.Current(s.menuItems[s.currentItem])
		case chstr == "104", chstr == "63": // "h", "?"
			s.toggleHelp()
		case chstr == "3", chstr == "113", chstr == "27": // ctrl-c, "q", esc
			if s.helpVisible && chstr == "27" {
				s.toggleHelp()
			} else {
				// quit
				return nil
			}
		default:
			s.menu.Driver(gc.DriverActions[ch])
		}
	}
}

func (s *Service) fetchData() error {
	coins, err := cmc.GetAllCoinData(int(s.limit))
	if err != nil {
		return err
	}

	s.coins = []*cmc.Coin{}
	for i := range coins {
		coin := coins[i]
		s.coins = append(s.coins, &coin)
	}

	return nil
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
	err := s.renderMenu()
	if err != nil {
		panic(err)
	}
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
		pad.Right("[r]ank", 13, " "),
		pad.Right("[n]ame", 13, " "),
		pad.Right("[s]ymbol", 8, " "),
		pad.Left("[p]rice", 10, " "),
		pad.Left("[m]arket cap", 17, " "),
		pad.Left("24H [v]olume", 15, " "),
		pad.Left("[1]H%", 9, " "),
		pad.Left("[2]4H%", 9, " "),
		pad.Left("[7]D%", 9, " "),
		pad.Left("[t]otal supply", 20, " "),
		pad.Left("[a]vailable supply", 19, " "),
		pad.Left("[l]ast updated", 17, " "),
	}
	header := ""
	for _, h := range headers {
		header = fmt.Sprintf("%s%s", header, h)
	}

	s.menuHeader = header
}

// SetColorPairs sets color pairs
func (s *Service) setColorPairs() {
	switch s.primaryColor {
	case "green":
		gc.InitPair(1, gc.C_GREEN, gc.C_BLACK)
	case "cyan", "blue":
		gc.InitPair(1, gc.C_CYAN, gc.C_BLACK)
	case "magenta", "pink", "purple":
		gc.InitPair(1, gc.C_MAGENTA, gc.C_BLACK)
	case "white":
		gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)
	case "red":
		gc.InitPair(1, gc.C_RED, gc.C_BLACK)
	case "yellow", "orange":
		gc.InitPair(1, gc.C_YELLOW, gc.C_BLACK)
	default:
		gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)
	}

	gc.InitPair(2, gc.C_BLACK, gc.C_BLACK)
	gc.InitPair(3, gc.C_BLACK, gc.C_GREEN)
	gc.InitPair(4, gc.C_BLACK, gc.C_CYAN)
	gc.InitPair(5, gc.C_WHITE, gc.C_BLUE)
	gc.InitPair(6, gc.C_BLACK, -1)
}

// RenderMainWindow renders main window
func (s *Service) renderMainWindow() error {
	if s.mainwin == nil {
		var err error
		s.mainwin, err = gc.NewWindow(s.screenRows, s.screenCols, 0, 0)
		if err != nil {
			return err
		}
	}
	s.mainwin.Clear()
	s.mainwin.ColorOn(2)
	s.mainwin.MoveWindow(0, 0)
	s.mainwin.Resize(s.screenRows, s.screenCols)
	s.mainwin.Box(0, 0)
	s.mainwin.Refresh()
	return nil
}

// ResizeWindows resizes windows
func (s *Service) resizeWindows() {
	gc.ResizeTerm(s.screenRows, s.screenCols)
	//s.log(fmt.Sprintf("%v %v", s.screenCols, s.screenRows))
	s.renderMainWindow()
	s.renderMenu()
	s.renderHelpBar()
	s.renderLogWindow()
	s.renderHelpWindow()
}

func (s *Service) renderHelpBar() error {
	var err error
	if s.helpbarwin == nil {
		s.helpbarwin, err = gc.NewWindow(1, s.screenCols, s.screenRows-1, 0)
		if err != nil {
			return err
		}
	}

	s.helpbarwin.Clear()
	s.helpbarwin.Resize(1, s.screenCols)
	s.helpbarwin.MoveWindow(s.screenRows-1, 0)
	s.helpbarwin.ColorOn(2)
	s.helpbarwin.Box(0, 0)
	s.helpbarwin.ColorOff(2)
	s.helpbarwin.ColorOn(1)
	s.helpbarwin.MovePrint(0, 0, "[q]uit [h]elp")
	s.helpbarwin.ColorOff(1)
	s.helpbarwin.Refresh()
	return nil
}

func (s *Service) renderLogWindow() error {
	var err error
	if s.logwin == nil {
		s.logwin, err = gc.NewWindow(1, 20, s.screenRows-1, s.screenCols-20)
		if err != nil {
			return err
		}
	}

	s.logwin.Clear()
	s.logwin.Resize(1, 20)
	s.logwin.MoveWindow(s.screenRows-1, s.screenCols-20)
	s.logwin.ColorOn(2)
	s.logwin.Box(0, 0)
	s.logwin.ColorOff(2)
	s.logwin.ColorOn(1)
	s.logwin.MovePrint(0, 0, s.lastLog)
	s.logwin.ColorOff(1)
	s.logwin.Refresh()
	return nil
}

// Log logs debug messages
func (s *Service) log(msg string) {
	s.lastLog = msg
	s.logwin.Clear()
	s.logwin.ColorOn(2)
	s.logwin.Box(0, 0)
	s.logwin.ColorOff(2)
	s.logwin.ColorOn(1)
	s.logwin.MovePrint(0, 0, msg)
	s.logwin.ColorOff(1)
	s.logwin.Refresh()
}

func (s *Service) toggleHelp() {
	s.helpVisible = !s.helpVisible
	s.renderHelpWindow()
}

func (s *Service) renderHelpWindow() error {
	if !s.helpVisible {
		if s.helpwin != nil {
			s.helpwin.ClearOk(true)
			s.helpwin.Clear()
			s.helpwin.SetBackground(gc.ColorPair(6))
			s.helpwin.ColorOn(6)
			s.helpwin.Resize(0, 0)
			s.helpwin.MoveWindow(200, 200)
			s.helpwin.Refresh()
			s.renderMenu()
		}
		return nil
	}

	var err error
	if s.helpwin == nil {
		s.helpwin, err = gc.NewWindow(21, 40, (s.screenRows/2)-11, (s.screenCols/2)-20)
		if err != nil {
			return err
		}
	}

	s.helpwin.Clear()
	s.helpwin.SetBackground(gc.ColorPair(1))
	s.helpwin.ColorOn(1)
	s.helpwin.Resize(21, 40)
	s.helpwin.MoveWindow((s.screenRows/2)-11, (s.screenCols/2)-20)
	s.helpwin.Box(0, 0)
	s.helpwin.MovePrint(0, 1, "Help")
	s.helpwin.MovePrint(1, 1, "<up> or <k> to navigate up")
	s.helpwin.MovePrint(2, 1, "<down> or <j> to navigate down")
	s.helpwin.MovePrint(3, 1, "<ctrl-u> to to page up")
	s.helpwin.MovePrint(4, 1, "<ctrl-d> to to page down")
	s.helpwin.MovePrint(5, 1, "<enter> or <space> to open coin link")
	s.helpwin.MovePrint(6, 1, "<1> to sort by 1 hour change")
	s.helpwin.MovePrint(7, 1, "<2> to sort by 24 hour volume")
	s.helpwin.MovePrint(8, 1, "<7> to sort by 7 day change")
	s.helpwin.MovePrint(9, 1, "<a> to sort by available supply")
	s.helpwin.MovePrint(10, 1, "<h> or <?> to toggle help")
	s.helpwin.MovePrint(11, 1, "<l> to sort by last updated")
	s.helpwin.MovePrint(12, 1, "<m> to sort by market cap")
	s.helpwin.MovePrint(13, 1, "<n> to sort by name")
	s.helpwin.MovePrint(14, 1, "<r> to sort by rank")
	s.helpwin.MovePrint(15, 1, "<s> to sort by symbol")
	s.helpwin.MovePrint(16, 1, "<t> to sort by total supply")
	s.helpwin.MovePrint(17, 1, "<p> to sort by price")
	s.helpwin.MovePrint(18, 1, "<v> to sort by 24 hour volume")
	s.helpwin.MovePrint(19, 1, "<q> or <esc> to quit application.")
	s.helpwin.Refresh()
	return nil
}

// OnWindowResize sends event to channel when resize event occurs
func (s *Service) onWindowResize(channel chan os.Signal) {
	//stdScr, _ := gc.Init()
	//stdScr.ScrollOk(true)
	//gc.NewLines(true)
	for {
		<-channel
		//gc.StdScr().Clear()
		//rows, cols := gc.StdScr().MaxYX()
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
func (s *Service) renderMenu() error {
	s.menuwinWidth = s.screenCols
	s.menuwinHeight = s.screenRows - 1
	s.menuWidth = s.screenCols
	s.menuHeight = s.screenRows - 2

	//if len(s.menuItems) == 0 {
	items := make([]*gc.MenuItem, len(s.menuData))
	var err error
	for i, val := range s.menuData {
		items[i], err = gc.NewItem(val, "")
		if err != nil {
			return err
		}
		//defer items[i].Free()
	}

	s.menuItems = items
	//}

	if s.menu == nil {
		var err error
		s.menu, err = gc.NewMenu(s.menuItems)
		if err != nil {
			return err
		}
	} else {
		s.menu.UnPost()
		s.menu.SetItems(s.menuItems)
		s.menu.Current(s.menuItems[s.currentItem])
	}

	if s.menuwin == nil {
		var err error
		s.menuwin, err = gc.NewWindow(s.menuwinHeight, s.menuwinWidth, 0, 0)
		s.menuwin.ScrollOk(true)
		if err != nil {
			return err
		}

		s.menuwin.Keypad(true)
		s.menu.SetWindow(s.menuwin)
		s.menusubwin = s.menuwin.Derived(s.menuHeight, s.menuWidth, 1, 0)
		s.menu.SubWindow(s.menusubwin)
		s.menu.Option(gc.O_ONEVALUE, false)
		s.menu.Format(s.menuHeight, 0)
		s.menu.Mark("")
	} else {
		s.menusubwin.Resize(s.menuHeight, s.menuWidth)
		s.menuwin.Resize(s.menuHeight, s.menuWidth)
	}

	//s.menuwin.Clear()
	s.menuwin.ColorOn(2)
	s.menuwin.Box(0, 0)
	s.menuwin.ColorOff(2)
	s.menuwin.ColorOn(1)
	s.menuwin.MovePrint(0, 0, s.menuHeader)
	s.menuwin.ColorOff(1)
	s.menuwin.ColorOn(2)
	s.menuwin.MoveAddChar(2, 0, gc.ACS_LTEE)
	//s.menuwin.HLine(2, 1, gc.ACS_HLINE, s.screenCols-6)
	s.menuwin.ColorOff(2)
	s.menu.Post()
	s.menuwin.Refresh()
	return nil
}
