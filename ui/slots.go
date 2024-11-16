package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	lipgloss "github.com/charmbracelet/lipgloss/v2"
	"github.com/treethought/ethtop/evm"
)

var slotContainerStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("#FF69B4"))

type SlotsView struct {
	slots       []*evm.BeaconEvent
	limit       int
	latestEpoch uint64
	epochSlots  map[uint64][]*evm.BeaconEvent
	h, w        int

	style lipgloss.Style
}

func NewSlotsView() *SlotsView {
	return &SlotsView{
		epochSlots: make(map[uint64][]*evm.BeaconEvent),
		style:      slotContainerStyle,
	}
}

func (m *SlotsView) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *SlotsView) SetSize(w, h int) {
	x, y := m.style.GetFrameSize()
	m.w = w - x - m.style.GetHorizontalBorderSize()
	m.h = h - y - m.style.GetVerticalBorderSize()
	m.style = m.style.Width(m.w).Height(m.h)
}

func (m *SlotsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *evm.BeaconEvent:
		m.slots = append(m.slots, msg)
		m.epochSlots[msg.Epoch] = append(m.epochSlots[msg.Epoch], msg)
		m.latestEpoch = msg.Epoch
		return m, nil
	}

	return m, nil
}

func (m *SlotsView) View() string {
	// each slot has epoch
	// need to sort slots into rows of ☐
	// with each row being an epoch

	// epochs are 32 slots long, so we need
	// to size rendered slots to fit the screen
	// if width too small then need to render dots

	epochs := []uint64{}
	for epoch := range m.epochSlots {
		epochs = append(epochs, epoch)
	}
	sort.Slice(epochs, func(i, j int) bool {
		return epochs[i] > epochs[j]
	})

	out := ""
	for _, epoch := range epochs {
		out += m.renderEpoch(m.epochSlots[epoch], m.w) + "\n"
	}
	titleStyle := NewStyle().Foreground(lipgloss.Color("#FF69B4"))
	title := titleStyle.Render("beacon chain")

	return m.style.Render(lipgloss.JoinVertical(lipgloss.Top, title, out))

}

func (m *SlotsView) renderEpoch(epoch []*evm.BeaconEvent, maxw int) string {
	maxSlotw := maxw / 32
	slots := []string{}
	for _, slot := range epoch {
		slots = append(slots, m.renderSlot(slot, maxSlotw))
	}

	s := NewStyle().MaxWidth(maxw)

	return s.Render(lipgloss.JoinHorizontal(lipgloss.Left, slots...))
}

func (m *SlotsView) renderSlot(slot *evm.BeaconEvent, maxw int) string {
	base := NewStyle().Margin(0).Padding(0, 1).Width(maxw).MaxWidth(maxw)

	propsedNewEpoch := base.Foreground(lipgloss.Color("#FF69B4")) //.Background(lipgloss.Color("#000000"))
	proposed := base.Foreground(lipgloss.Color("#FF0000"))        //.Background(lipgloss.Color("#000000"))
	confirmed := base.Foreground(lipgloss.Color("#00FF00"))       //.Background(lipgloss.Color("#000000"))
	finalized := base.Foreground(lipgloss.Color("#0000FF"))       //.Background(lipgloss.Color("#000000"))

	var s lipgloss.Style
	switch {
	case slot.EpochTransition && slot.Epoch == m.latestEpoch:
		s = propsedNewEpoch
	case slot.Slot/32 == m.latestEpoch:
		s = proposed
	case slot.Slot/32 == m.latestEpoch-1:
		s = confirmed
	default:
		s = finalized
	}

	sm := s.
		SetString(strings.Repeat("☐", maxw)).
		Padding(0, 1).
		String()

	out := s.Render(fmt.Sprintf("%d", slot.Slot))
	if lipgloss.Width(out) >= maxw {
		out = sm
	}

	return out
}
