package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/pomdtr/sunbeam/schemas"
	"github.com/pomdtr/sunbeam/types"

	"gopkg.in/yaml.v3"
)

type PageGenerator func() ([]byte, error)

func ExpandAction(action types.Action, old, new string) types.Action {
	expandCommad := func(command *types.Command) *types.Command {
		if command == nil {
			return nil
		}

		command.Name = strings.ReplaceAll(command.Name, old, shellescape.Quote(new))
		for i, arg := range command.Args {
			command.Args[i] = strings.ReplaceAll(arg, old, shellescape.Quote(new))
		}

		command.Dir = strings.ReplaceAll(command.Dir, old, new)
		command.Input = strings.ReplaceAll(command.Input, old, new)

		return command
	}

	action.Command = expandCommad(action.Command)
	action.Url = strings.ReplaceAll(action.Url, old, url.QueryEscape(new))
	action.Text = strings.ReplaceAll(action.Text, old, new)
	action.Path = strings.ReplaceAll(action.Path, old, new)
	if action.Page != nil {
		expandCommad(action.Page.Command)
	}

	return action
}

func NewCommandGenerator(command *types.Command) PageGenerator {
	return func() ([]byte, error) {
		output, err := command.Output()
		if err != nil {
			return nil, err
		}

		var v any
		if err := json.Unmarshal(output, &v); err != nil {
			return nil, err
		}

		if err := schemas.Validate(v); err != nil {
			return nil, err
		}

		var page types.Page
		if err := json.Unmarshal(output, &page); err != nil {
			return nil, err
		}

		page, err = expandPage(page, command.Dir)
		if err != nil {
			return nil, err
		}

		return json.Marshal(page)
	}
}

func NewFileGenerator(name string) PageGenerator {
	return func() ([]byte, error) {
		var page types.Page
		if path.Ext(name) == ".json" {
			bytes, err := os.ReadFile(name)
			if err != nil {
				return nil, err
			}

			var v any
			if err := json.Unmarshal(bytes, &v); err != nil {
				return nil, err
			}

			if err := schemas.Validate(v); err != nil {
				return nil, err
			}

			if err := json.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}
		} else if path.Ext(name) == ".yaml" || path.Ext(name) == ".yml" {
			bytes, err := os.ReadFile(name)
			if err != nil {
				return nil, err
			}

			var v any
			if err := yaml.Unmarshal(bytes, &v); err != nil {
				return nil, err
			}

			if err := schemas.Validate(v); err != nil {
				return nil, err
			}

			if err := yaml.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unsupported file type")
		}

		page, err := expandPage(page, filepath.Dir(name))
		if err != nil {
			return nil, err
		}

		return json.Marshal(page)
	}
}

func expandPage(page types.Page, dir string) (types.Page, error) {
	var err error
	expandAction := func(action types.Action) (types.Action, error) {
		switch action.Type {
		case types.CopyAction:
			if action.Title == "" {
				action.Title = "Copy"
			}
		case types.RunAction:
			if action.Title == "" {
				action.Title = "Run"
			}
			if action.Command.Dir == "" {
				action.Command.Dir = dir
			}
		case types.PushPageAction:
			if action.Title == "" {
				action.Title = "Push"
			}

			if action.Page.Command.Dir == "" {
				action.Page.Command.Dir = dir
			}

			if !filepath.IsAbs(action.Page.Path) {
				action.Path = filepath.Join(dir, action.Path)
			}
		case types.OpenPathAction:
			if action.Title == "" {
				action.Title = "Open File"
			}

			if action.Path != "" && !filepath.IsAbs(action.Path) {
				action.Path = filepath.Join(dir, action.Path)
			}
		case types.OpenUrlAction:
			if action.Title == "" {
				action.Title = "Open in Browser"
			}
		}

		return action, nil
	}

	for i, action := range page.Actions {
		action, err = expandAction(action)
		if err != nil {
			return page, err
		}

		page.Actions[i] = action
	}

	switch page.Type {
	case types.DetailPage:
		if page.Preview.Command.Dir == "" {
			page.Preview.Command.Dir = dir
		}

	case types.ListPage:
		for i, item := range page.Items {
			for j, action := range item.Actions {
				action, err := expandAction(action)
				if err != nil {
					return page, err
				}
				page.Items[i].Actions[j] = action
			}
		}
	}

	return page, nil
}
