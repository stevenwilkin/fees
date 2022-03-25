package main

import (
	"fmt"
	"os"
	"time"
	"unicode"

	"github.com/stevenwilkin/fees/binance"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/joho/godotenv/autoload"
)

var (
	margin = lipgloss.NewStyle().Margin(1, 2, 0, 2)
	bold   = lipgloss.NewStyle().Bold(true)
)

type priceMsg float64
type bnbMsg float64

type model struct {
	price     float64
	bnb       float64
	textInput textinput.Model
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			if !unicode.IsDigit(msg.Runes[0]) {
				return m, nil
			}
		}
	case priceMsg:
		m.price = float64(msg)
	case bnbMsg:
		m.bnb = float64(msg)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	tradingVolume := 0.0
	if m.bnb > 0 && m.price > 0 {
		tradingVolume = (m.bnb * m.price) / (0.075 / 100)
	}

	bnbUsdt := fmt.Sprintf("%s: %.2f", bold.Render("BNBUSDT"), m.price)
	bnb := fmt.Sprintf("%s:     %.3f", bold.Render("BNB"), m.bnb)
	value := fmt.Sprintf("%s:   %.2f", bold.Render("Value"), m.bnb*m.price)
	volume := fmt.Sprintf("%s:  %.0f", bold.Render("Volume"), tradingVolume)

	input := m.textInput.View()

	return margin.Render(fmt.Sprintf(
		"%s\n\n%s\n%s\n%s\n\n%s", bnbUsdt, bnb, value, volume, input))
}

func main() {
	input := textinput.New()
	input.Focus()

	m := model{
		textInput: input}
	p := tea.NewProgram(m, tea.WithAltScreen())

	b := &binance.Binance{
		ApiKey:    os.Getenv("BINANCE_API_KEY"),
		ApiSecret: os.Getenv("BINANCE_API_SECRET")}

	go func() {
		for price := range b.Price() {
			p.Send(priceMsg(price))
		}
	}()

	go func() {
		t := time.NewTicker(1 * time.Second)

		for {
			bnb, err := b.GetBalance()
			if err != nil {
				panic(err)
			}

			p.Send(bnbMsg(bnb))
			<-t.C
		}
	}()

	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
