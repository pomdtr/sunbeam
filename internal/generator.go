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
	"github.com/pomdtr/sunbeam/utils"

	"gopkg.in/yaml.v3"
)

type PageGenerator func() ([]byte, error)

func ExpandAction(action types.Action, inputs map[string]string) types.Action {
	for key, value := range inputs {
		action.Command = strings.ReplaceAll(action.Command, fmt.Sprintf("${input:%s}", key), shellescape.Quote(value))
		action.Url = strings.ReplaceAll(action.Url, fmt.Sprintf("${input:%s}", key), url.QueryEscape(value))
		action.Text = strings.ReplaceAll(action.Text, fmt.Sprintf("${input:%s}", key), value)
		action.Path = strings.ReplaceAll(action.Path, fmt.Sprintf("${input:%s}", key), value)
	}

	for _, env := range os.Environ() {
		tokens := strings.SplitN(env, "=", 2)
		if len(tokens) != 2 {
			continue
		}
		action.Command = strings.ReplaceAll(action.Command, fmt.Sprintf("${env:%s}", tokens[0]), shellescape.Quote(tokens[1]))
		action.Url = strings.ReplaceAll(action.Url, fmt.Sprintf("${env:%s}", tokens[0]), url.QueryEscape(tokens[1]))
		action.Text = strings.ReplaceAll(action.Text, fmt.Sprintf("${env:%s}", tokens[0]), tokens[1])
		action.Path = strings.ReplaceAll(action.Path, fmt.Sprintf("${env:%s}", tokens[0]), tokens[1])
	}

	if strings.HasPrefix(action.Path, "~/") {
		homeDir, _ := os.UserHomeDir()
		action.Path = filepath.Join(homeDir, action.Path[2:])
	}

	return action
}

func NewCommandGenerator(command string, input string, dir string) PageGenerator {
	return func() ([]byte, error) {
		output, err := utils.RunCommand(command, strings.NewReader(input), dir)
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

		page, err = expandPage(page, dir)
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
			if action.Dir == "" {
				action.Dir = dir
			}
		case types.PushPageAction:
			if action.Title == "" {
				action.Title = "Push"
			}

			if action.Page.Dir == "" {
				action.Page.Dir = dir
			}

			if !filepath.IsAbs(action.Page.Path) {
				action.Path = filepath.Join(dir, action.Path)
			}
		case types.OpenFileAction:
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
		if page.Preview.Dir == "" {
			page.Preview.Dir = dir
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
