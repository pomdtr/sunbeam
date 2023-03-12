package tui

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/utils"
)

type ReloadPageMsg struct{}

type PopPageMsg struct{}

type pushMsg struct {
	container Page
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type SunbeamOptions struct {
	MaxHeight int
	Padding   int
}

type Model struct {
	width, height int
	options       SunbeamOptions
	exitCmd       *exec.Cmd

	pages []*CommandRunner
	form  *Form

	hidden bool
}

func NewModel(root *CommandRunner, options SunbeamOptions) *Model {
	return &Model{pages: []*CommandRunner{
		root,
	}, options: options}
}

func (m *Model) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}

	return m.pages[0].Init()
}

func (m *Model) CurrentUrl() *url.URL {
	if len(m.pages) == 0 {
		return nil
	}

	return m.pages[len(m.pages)-1].baseUrl
}

func (m *Model) handleAction(action schemas.Action) tea.Cmd {
	switch action.Type {
	case schemas.ReloadAction:
		return func() tea.Msg {
			return ReloadPageMsg{}
		}
	case schemas.OpenAction:
		err := browser.OpenURL(action.Target)
		if err != nil {
			return func() tea.Msg {
				return err
			}
		}

		m.hidden = true
		return tea.Quit
	case schemas.CopyAction:
		err := clipboard.WriteAll(action.Text)
		if err != nil {
			return func() tea.Msg {
				return fmt.Errorf("failed to copy text to clipboard: %s", err)
			}
		}

		m.hidden = true
		return tea.Quit
	case schemas.ReadAction:
		page := action.Page

		baseUrl := m.CurrentUrl()
		runner := NewCommandRunner(func(query string) ([]byte, error) {
			switch baseUrl.Scheme {
			case "http", "https":
				target := url.URL{
					Scheme: baseUrl.Scheme,
					Path:   path.Join(baseUrl.Path, page),
				}

				res, err := http.Get(target.String())
				if err != nil {
					return nil, fmt.Errorf("failed to get page: %s", err)
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("failed to get page: %s", res.Status)
				}

				return io.ReadAll(res.Body)
			case "file":
				return os.ReadFile(path.Join(baseUrl.Path, page))
			default:
				return nil, fmt.Errorf("unsupported scheme: %s", m.CurrentUrl().Scheme)
			}
		}, &url.URL{
			Scheme: baseUrl.Scheme,
			Path:   path.Dir(page),
		})
		return m.Push(runner)
	case schemas.RunAction:
		if m.CurrentUrl().Scheme != "file" {
			return func() tea.Msg {
				return fmt.Errorf("unsupported scheme: %s", m.CurrentUrl().Scheme)
			}
		}

		switch action.OnSuccess {
		case schemas.PushOnSuccess:
			return m.Push(NewCommandRunner(func(query string) ([]byte, error) {
				cmd := exec.Command(action.Command, action.Args...)
				cmd.Stdin = strings.NewReader(query)
				cmd.Dir = m.CurrentUrl().Path
				output, err := cmd.Output()
				if err != nil {
					if err, ok := err.(*exec.ExitError); ok {
						return nil, fmt.Errorf("command exit with code %d: %s", err.ExitCode(), err.Stderr)
					}
					return nil, err
				}

				return output, nil
			}, m.CurrentUrl()))
		case schemas.ReloadOnSuccess:
			return func() tea.Msg {
				command := exec.Command(action.Command, action.Args...)
				err := command.Run()
				if err != nil {
					if err, ok := err.(*exec.ExitError); ok {
						return fmt.Errorf("command exit with code %d: %s", err.ExitCode(), err.Stderr)
					}
					return err
				}

				return ReloadPageMsg{}
			}
		case schemas.ExitOnSuccess:
			command := exec.Command(action.Command, action.Args...)
			m.exitCmd = command
			m.hidden = true
			return tea.Quit
		default:
			return func() tea.Msg {
				return fmt.Errorf("unsupported onSuccess")
			}
		}
	default:
		return func() tea.Msg {
			return fmt.Errorf("unknown action type")
		}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			fmt.Sprintln("Escape")
		case tea.KeyCtrlC:
			m.hidden = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case PopPageMsg:
		if m.form != nil {
			m.form = nil
			return m, nil
		}

		if len(m.pages) > 1 {
			m.Pop()
			return m, nil
		}

		m.hidden = true
		return m, tea.Quit
	case schemas.Action:
		if len(msg.Inputs) > 0 {
			formItems := make([]FormItem, len(msg.Inputs))
			for i, input := range msg.Inputs {
				item, err := NewFormItem(input)
				if err != nil {
					return m, func() tea.Msg {
						return fmt.Errorf("failed to create form input: %s", err)
					}
				}

				formItems[i] = item
			}

			form := NewForm(formItems, func(values map[string]string) tea.Cmd {
				for key, value := range values {
					msg.Command = strings.ReplaceAll(msg.Command, fmt.Sprintf("${input:%s}", key), value)
					for i, arg := range msg.Args {
						arg = strings.ReplaceAll(arg, fmt.Sprintf("${input:%s}", key), value)
						msg.Args[i] = arg
					}
				}

				return func() tea.Msg {
					return schemas.Action{
						Command:   msg.Command,
						Args:      msg.Args,
						OnSuccess: msg.OnSuccess,
					}
				}
			})
			m.form = form
			form.SetSize(m.pageWidth(), m.pageHeight())
			return m, form.Init()
		}
		m.form = nil

		return m, m.handleAction(msg)
	}

	// Update the current page
	var cmd tea.Cmd

	if m.form != nil {
		m.form, cmd = m.form.Update(msg)
	} else if len(m.pages) > 0 {
		currentPageIdx := len(m.pages) - 1
		m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	} else {
		return m, nil
	}

	return m, cmd
}

func (m *Model) View() string {
	if m.hidden {
		return ""
	}

	if m.form != nil {
		return m.form.View()
	}

	var pageView string
	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		pageView = currentPage.View()
	}

	return lipgloss.NewStyle().Padding(m.options.Padding).Render(pageView)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.pageWidth(), m.pageHeight())
	}
}

func (m *Model) pageWidth() int {
	return m.width - 2*m.options.Padding
}

func (m *Model) pageHeight() int {
	if m.options.MaxHeight > 0 {
		return utils.Min(m.options.MaxHeight, m.height) - 2*m.options.Padding
	}
	return m.height - 2*m.options.Padding
}

func (m *Model) Push(page *CommandRunner) tea.Cmd {
	page.SetSize(m.pageWidth(), m.pageHeight())
	m.pages = append(m.pages, page)
	return page.Init()
}

func (m *Model) Pop() {
	if len(m.pages) > 0 {
		m.pages = m.pages[:len(m.pages)-1]
	}
}

func (m *Model) Draw() (err error) {
	// Background detection before we start the program
	lipgloss.SetHasDarkBackground(lipgloss.HasDarkBackground())

	os.Setenv("SUNBEAM_RUNNER", "true")

	var p *tea.Program
	if m.options.MaxHeight == 0 {
		p = tea.NewProgram(m, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(m)
	}

	res, err := p.Run()
	if err != nil {
		return err
	}

	model, ok := res.(*Model)
	if !ok {
		return fmt.Errorf("could not convert res back to *Model")
	}

	if model.exitCmd != nil {
		model.exitCmd.Stdin = os.Stdin
		model.exitCmd.Stdout = os.Stdout
		model.exitCmd.Stderr = os.Stderr

		return model.exitCmd.Run()
	}

	return nil
}
