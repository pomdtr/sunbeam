package main

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jinzhu/copier"
)

type ListPage struct {
	Script Script
	list.Model
}

func NewListPage(script Script, l list.Model) ListPage {
	l.Title = script.Title()

	return ListPage{
		Script: script,
		Model:  l,
	}
}

func (page ListPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			selectedItem, ok := page.Model.SelectedItem().(RaycastItem)
			if !ok {
				return page, tea.Quit
			}
			primaryAction := selectedItem.Actions[0]
			return page, runAction(page.Script, primaryAction)
		}
	case tea.WindowSizeMsg:
		page.SetSize(msg.Width, msg.Height)
		return page, nil
	case ScriptResponse:
		log.Printf("Pushing %d items", len(msg.List.Items))
		items := make([]list.Item, len(msg.List.Items))
		for i, item := range msg.List.Items {
			items[i] = item
		}
		cmd = page.SetItems(items)
	}

	page.Model, cmd = page.Model.Update(msg)

	return page, cmd
}

func (page ListPage) Init() tea.Cmd {
	return func() tea.Msg {
		res, err := runCommand(page.Script.Path)
		if err != nil {
			return err
		}
		return res
	}
}

type RootPage struct {
	commandDir string
	list.Model
}

func NewRootPage(commandDir string) RootPage {
	d := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Root"

	return RootPage{Model: l, commandDir: commandDir}
}

func (page RootPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			selectedItem, ok := page.Model.SelectedItem().(Script)
			if !ok {
				return page, tea.Quit
			}
			return page, func() tea.Msg {
				return PushMsg{Script: selectedItem}
			}
		}
	case tea.WindowSizeMsg:
		page.SetSize(msg.Width, msg.Height)
		return page, nil
	case []Script:
		items := make([]list.Item, len(msg))
		for i, script := range msg {
			items[i] = script
		}

		cmd = page.SetItems(items)
		return page, cmd
	}

	page.Model, cmd = page.Model.Update(msg)

	return page, cmd
}

func (page RootPage) Init() tea.Cmd {
	return func() tea.Msg {
		scripts, err := ScanDir(page.commandDir)
		if err != nil {
			return err
		}
		return scripts
	}
}

type PageFactory struct {
	list   list.Model
	form   tea.Model
	detail tea.Model
}

func NewPageFactory() PageFactory {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	return PageFactory{
		list: l,
	}
}

func (f *PageFactory) SetSize(w, h int) {
	f.list.SetSize(w, h)
}

func (f PageFactory) BuildPage(s Script) tea.Model {
	var l list.Model
	copier.CopyWithOption(&l, &f.list, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return NewListPage(s, f.list)
}
