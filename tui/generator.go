package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/google/shlex"
	"github.com/pomdtr/sunbeam/types"

	"gopkg.in/yaml.v3"
)

type PageGenerator func(input string) ([]byte, error)

type CmdGenerator struct {
	Command string
	Args    []string
	Dir     string
}

func NewCommandGenerator(command string, input string, dir string) PageGenerator {
	for _, value := range os.Environ() {
		tokens := strings.SplitN(value, "=", 2)
		if len(tokens) != 2 {
			continue
		}

		command = strings.ReplaceAll(command, fmt.Sprintf("${env:%s}", tokens[0]), shellescape.Quote(tokens[1]))
		input = strings.ReplaceAll(input, fmt.Sprintf("${env:%s}", tokens[0]), tokens[1])
	}

	return func(query string) ([]byte, error) {
		command := strings.ReplaceAll(command, "${query}", shellescape.Quote(input))

		args, err := shlex.Split(command)
		if err != nil {
			return nil, err
		}

		if len(args) == 0 {
			return nil, fmt.Errorf("no command provided")
		}

		var extraArgs []string
		if len(args) > 1 {
			extraArgs = args[1:]
		}

		cmd := exec.Command(args[0], extraArgs...)
		cmd.Stdin = strings.NewReader(input)
		cmd.Dir = dir
		cmd.Stdin = strings.NewReader(input)
		output, err := cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return nil, fmt.Errorf("script exited with code %d: %s", exitError.ExitCode(), string(exitError.Stderr))
			}

			return nil, err
		}

		return output, nil
	}
}

func NewFileGenerator(name string) PageGenerator {
	return func(input string) ([]byte, error) {
		if path.Ext(name) == ".json" {
			return os.ReadFile(name)
		}

		if path.Ext(name) == ".yaml" || path.Ext(name) == ".yml" {
			bytes, err := os.ReadFile(name)
			if err != nil {
				return nil, err
			}

			var page types.Page
			if err := yaml.Unmarshal(bytes, &page); err != nil {
				return nil, err
			}

			return json.Marshal(page)
		}

		return nil, fmt.Errorf("unsupported file type")
	}
}

func NewHttpGenerator(target string, method string, headers map[string]string, body string) PageGenerator {
	for _, env := range os.Environ() {
		tokens := strings.SplitN(env, "=", 2)
		if len(tokens) != 2 {
			continue
		}
		target = strings.Replace(target, fmt.Sprintf("${env:%s}", tokens[0]), url.QueryEscape(tokens[1]), -1)
		body = strings.Replace(body, fmt.Sprintf("${env:%s}", tokens[0]), tokens[1], -1)
		for key, value := range headers {
			headers[key] = strings.Replace(value, fmt.Sprintf("${env:%s}", tokens[0]), tokens[1], -1)
		}
	}

	return func(query string) ([]byte, error) {
		target = strings.Replace(target, "${query}", url.QueryEscape(query), -1)
		body = strings.Replace(body, "${query}", query, -1)
		for key, value := range headers {
			headers[key] = strings.Replace(value, "${query}", value, -1)
		}

		req, err := http.NewRequest(method, target, strings.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("could not create request: %w", err)
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not make request: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}

		return io.ReadAll(res.Body)
	}
}
