package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dev-hack95/mini/client"
	"github.com/dev-hack95/mini/structs"
	"github.com/dev-hack95/mini/utilities"
)

const innerHeight = 1
const footerHeight = 2

var (
	propmtStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type OllamaResponse struct {
	content string
	err     error
	elapsed time.Duration
}

type model struct {
	textInput    textinput.Model
	viewport     viewport.Model
	spinner      spinner.Model
	client       *client.Client
	db           *sql.DB
	memory       []structs.Message
	messages     []string
	isProcessing bool
	ready        bool
	search       bool
	autosave     bool
	err          error
	height       int
	width        int
	mdRenderer   *glamour.TermRenderer
}

func initModel() *model {
	ti := textinput.New()
	ti.CharLimit = 1024 * 100
	ti.Prompt = "Enter Prompt: "
	ti.Reset()
	ti.Focus()
	ti.Width = 30

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	client, err := client.NewClient()

	if err != nil {
		log.Println("Error: error occured while creating new client  \n step-1:  Install ollama: 'curl -fsSL https://ollama.com/install.sh | sh' (for linux refer ollama.com for mac-os and windows)  \n 2) ollama pull granite3.1-moe:latest")
		os.Exit(1)
	}

	db, err := utilities.SessionDB("/.term-ollama/db/sessions.db")

	if err != nil {
		log.Println("Error occured while connecting to the database: " + err.Error())
		os.Exit(1)
	}

	return &model{
		textInput:    ti,
		spinner:      sp,
		client:       client,
		db:           db,
		memory:       []structs.Message{},
		messages:     []string{},
		isProcessing: false,
		ready:        false,
		search:       false,
		autosave:     false,
		err:          nil,
		mdRenderer:   renderer,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		spinner.Tick,
	)
}

func (m model) promptResponse(message []structs.Message) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		response, err := m.client.ChatOllama(message)
		if err != nil {
			return OllamaResponse{content: "", err: err, elapsed: time.Since(start)}
		}

		return OllamaResponse{content: response.Message.Content, err: nil, elapsed: time.Since(start)}
	}
}

func (m model) renderMarkdown(content string) string {
	rendered, err := m.mdRenderer.Render(content)

	if err != nil {
		return content
	}

	return rendered
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		newRenderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(msg.Width),
		)

		m.mdRenderer = newRenderer

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-innerHeight-footerHeight)
			m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height - innerHeight - footerHeight

			if len(m.messages) > 0 {
				m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
			}
		}

		m.textInput.Width = msg.Width - 2

		return m, nil

	case tea.KeyMsg:
		if m.isProcessing {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			default:
				return m, nil
			}
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.search == false {
				prompt := m.textInput.Value()
				if strings.TrimSpace(prompt) == "" {
					return m, nil
				}

				promptMsg := propmtStyle.Render(fmt.Sprintf("> %s", prompt))
				m.messages = append(m.messages, promptMsg)
				m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
				m.viewport.GotoBottom()

				m.textInput.Reset()
				m.textInput.Blur()
				m.isProcessing = true

				if strings.HasPrefix(prompt, "/new") {
					err := utilities.SaveSession(m.db, m.memory)

					if err == nil {
						m.isProcessing = false
						m.textInput.Focus()
						m.messages = []string{}
						m.memory = []structs.Message{}

						m.messages = append(m.messages, infoStyle.Render("New Session"))
						m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
						m.viewport.GotoBottom()
						return m, nil
					}

					if err != nil {
						return m, nil
					}
				}

				if strings.HasPrefix(prompt, "/bye") {
					return m, tea.Quit
				}

				var message structs.Message
				message.Role = "user"
				message.Content = prompt

				m.memory = append(m.memory, message)
				cmds = append(cmds, m.promptResponse(m.memory))
				cmds = append(cmds, spinner.Tick)
				return m, tea.Batch(cmds...)
			}

		}

	case spinner.TickMsg:
		if m.isProcessing {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case OllamaResponse:
		m.isProcessing = false

		var content string
		if msg.err != nil {
			m.err = msg.err
			content = errorStyle.Render(fmt.Sprintf("Error: %v", msg.err))
		} else {
			var message structs.Message
			message.Role = "assistant"
			message.Content = msg.content
			m.memory = append(m.memory, message)

			header := headerStyle.Render("Response: ")
			renderedContent := m.renderMarkdown(msg.content)
			timeInfo := infoStyle.Render(fmt.Sprintf("(Completed in %v)", msg.elapsed))

			content = fmt.Sprintf("%s\n%s\n%s", header, renderedContent, timeInfo)
		}

		m.messages = append(m.messages, content)
		m.viewport.SetContent(strings.Join(m.messages, "\n\n"))
		m.viewport.GotoBottom()

		m.textInput.Focus()

		return m, nil

	}

	if !m.isProcessing {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var s strings.Builder

	s.WriteString(m.viewport.View() + "\n\n")

	if m.isProcessing {
		s.WriteString(m.spinner.View() + " Processing request...\n")
	} else {
		s.WriteString(m.textInput.View() + "\n")
		// m.textInput.Reset()
	}

	s.WriteString(infoStyle.Render("(↑/↓: scroll, ctrl+c: quit)"))

	return s.String()
}

func main() {
	p := tea.NewProgram(
		initModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

}
