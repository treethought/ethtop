package ui

import (
	"context"

	log "github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea/v2"
	lipgloss "github.com/charmbracelet/lipgloss/v2"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/treethought/ethtop/config"
	"github.com/treethought/ethtop/evm"
)

func listenSub(heads <-chan *types.Header) tea.Cmd {
	return func() tea.Msg {
		h := <-heads
		log.Println("got head: ", h.Number)
		return h
	}
}
func startSub(heads chan<- *types.Header, c *evm.Client) tea.Cmd {
	return func() tea.Msg {
		go c.SubscribeHeads(context.Background(), heads)
		return nil
	}
}
func startEventSub(events chan<- *evm.BeaconEvent, c *evm.Client) tea.Cmd {
	return func() tea.Msg {
		c.SubscribeSlots(context.Background(), events)
		return nil
	}
}

func listenEventSub(events <-chan *evm.BeaconEvent) tea.Cmd {
	return func() tea.Msg {
		e := <-events
		return e
	}
}

type AppContext struct {
	Network string
}

type App struct {
	ctx          *AppContext
	cfg          *config.Config
	focusedModel tea.Model
	focused      string
	navname      string
	client       *evm.Client

	headChan  chan *types.Header
	eventChan chan *evm.BeaconEvent
	log       *log.Entry

	slotsView *SlotsView
	x, y      int
}

func NewApp(cfg *config.Config, ctx *AppContext) *App {
	if ctx == nil {
		ctx = &AppContext{}
	}
	client, err := evm.NewClient(cfg.RPC.HTTP, cfg.RPC.WS)
	if err != nil {
		log.WithError(err).Fatal("error creating client")
	}
	a := &App{
		ctx:       ctx,
		cfg:       cfg,
		log:       log.WithField("source", "app"),
		client:    client,
		slotsView: NewSlotsView(),
	}
	a.SetNavName("ethtop")

	return a
}

func (a *App) SetNavName(name string) {
	a.navname = name
}

func (a *App) focusMain() {
}

func (a *App) Init() (tea.Model, tea.Cmd) {
	a.log.Info("initializing app")
	a.headChan = make(chan *types.Header)
	a.eventChan = make(chan *evm.BeaconEvent, 100)
	sub := tea.Batch(
		startSub(a.headChan, a.client),
		listenSub(a.headChan),
		startEventSub(a.eventChan, a.client),
		listenEventSub(a.eventChan),
	)

	return a, tea.Sequence(tea.RequestWindowSize(), sub)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	a.log.WithField("msg", msg).Debug("update")
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.x, a.y = msg.Width, msg.Height
		x, y := msg.Width/2, msg.Height
		a.slotsView.SetSize(x, y)
		return a, nil

	case *evm.BeaconEvent:
		_, cmd := a.slotsView.Update(msg)
		return a, tea.Batch(cmd, listenEventSub(a.eventChan))

	case *types.Header:
		log.WithField("header", msg).Info("got header")
		return a, listenSub(a.headChan)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	return a, tea.Batch(cmds...)

}

func (a *App) View() string {

	titleStyle := NewStyle().Foreground(lipgloss.Color("#FF69B4"))

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.PlaceHorizontal(a.x, lipgloss.Center, titleStyle.Render(a.navname)),
		a.slotsView.View(),
	)

}
