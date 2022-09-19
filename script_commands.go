package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-playground/validator"
	"github.com/skratchdot/open-golang/open"
)

type any interface{}

var validate = validator.New()

type ScriptResponse struct {
	Type    string         `json:"type" validate:"required,oneof=list detail form exit"`
	List    ListResponse   `json:"list" validate:"dive"`
	Detail  DetailResponse `json:"detail" validate:"dive"`
	Form    FormResponse   `json:"form" validate:"dive"`
	Storage any            `json:"storage"`
}

type DetailResponse struct {
}

type FormResponse struct {
}

type ListResponse struct {
	Title string        `json:"title"`
	Items []RaycastItem `json:"items"`
}

type RaycastItem struct {
	Icon       string          `json:"icon"`
	TitleField string          `json:"title" validate:"required"`
	Subtitle   string          `json:"subtitle"`
	Fill       string          `json:"fill"`
	Actions    []RaycastAction `json:"actions" validate:"required,gte=1,dive"`
}

func (i RaycastItem) FilterValue() string { return i.TitleField }
func (i RaycastItem) Title() string       { return i.TitleField }
func (i RaycastItem) Description() string { return i.Subtitle }

type RaycastAction struct {
	Title   string   `json:"title" validate:"required,oneof=copy open open-url"`
	Type    string   `json:"type" validate:"required"`
	Path    string   `json:"path"`
	Url     string   `json:"url"`
	Content string   `json:"content"`
	Args    []string `json:"args"`
}

func runAction(script Script, action RaycastAction) tea.Cmd {
	switch action.Type {
	case "open":
		err := open.Run(action.Path)
		if err != nil {
			return func() tea.Msg {
				return err
			}
		}
		return tea.Quit
	case "open-url":
		open.Start(action.Url)
		return tea.Quit
	case "copy":
		clipboard.WriteAll(action.Content)
		return tea.Quit
	case "callback":
		return func() tea.Msg {
			return PushMsg{Script: script, Args: action.Args}
		}
	default:
		log.Fatalf("Unknown action type: %s", action.Type)
		return nil
	}
}

type Script struct {
	Path      string
	Metadatas ScriptMetadatas
}

func (s Script) FilterValue() string { return s.Metadatas.Title }
func (s Script) Title() string       { return s.Metadatas.Title }
func (s Script) Description() string { return s.Metadatas.PackageName }

type ScriptMetadatas struct {
	SchemaVersion        int    `validate:"required,eq=1"`
	Title                string `validate:"required"`
	Mode                 string `validate:"required,oneof=command"`
	PackageName          string
	Description          string
	Icon                 string
	CurrentDirectoryPath string
	NeedsConfirmation    bool
	Author               string
	AutorUrl             string
}

func extractRaycastMetadatas(content string) map[string]string {
	r := regexp.MustCompile("@raycast.([A-Za-z0-9]+)\\s([\\S ]+)")
	groups := r.FindAllStringSubmatch(content, -1)

	metadataMap := make(map[string]string)
	for _, group := range groups {
		metadataMap[group[1]] = group[2]
	}

	return metadataMap
}

func ParseScript(script_path string) (Script, error) {
	content, err := ioutil.ReadFile(script_path)
	if err != nil {
		return Script{}, err
	}

	metadatas := extractRaycastMetadatas(string(content))

	var schemaVersion int
	err = json.Unmarshal([]byte(metadatas["schemaVersion"]), &schemaVersion)
	if err != nil {
		return Script{}, err
	}

	var needsConfirmation bool
	json.Unmarshal([]byte(metadatas["schemaVersion"]), &needsConfirmation)

	scripCommand := Script{Path: script_path, Metadatas: ScriptMetadatas{
		SchemaVersion:        schemaVersion,
		Title:                metadatas["title"],
		Mode:                 metadatas["mode"],
		PackageName:          metadatas["packageName"],
		Icon:                 metadatas["icon"],
		CurrentDirectoryPath: metadatas["currentDirectoryPath"],
		NeedsConfirmation:    needsConfirmation,
		Author:               metadatas["author"],
		AutorUrl:             metadatas["autorUrl"],
		Description:          metadatas["description"],
	}}

	err = validate.Struct(scripCommand)
	if err != nil {
		println(err)
		return Script{}, err
	}

	return scripCommand, nil
}

func ScanDir(dirPath string) ([]Script, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return []Script{}, err
	}

	scripts := []Script{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		script, err := ParseScript(path.Join(dirPath, file.Name()))
		if err != nil {
			continue
		}

		scripts = append(scripts, script)
	}

	return scripts, nil
}

func runCommand(scriptPath string, args ...string) (res ScriptResponse, err error) {
	err = os.Chmod(scriptPath, 0755)
	if err != nil {
		return
	}

	cmd := exec.Command(scriptPath, args...)
	cmd.Dir = path.Dir(scriptPath)

	// Copy process environment
	// cmd.Env = make([]string, len(os.Environ()))
	// copy(cmd.Env, os.Environ())

	// Add support dir to environment
	// supportDir := path.Join(xdg.StateHome, "raycast", script.PackageName)
	// cmd.Env = append(cmd.Env, fmt.Sprintf("RAYCAST_SUPPORT_DIR=%s", supportDir))

	// add extension storage to environment
	// storage, _ := json.Marshal(a.db.Data.ExtensionStorage[script.PackageName])
	// cmd.Env = append(cmd.Env, fmt.Sprintf("RAYCAST_STORAGE=%s", string(storage)))

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err = cmd.Run()
	if err != nil {
		return res, err
	}
	json.Unmarshal(outbuf.Bytes(), &res)

	err = validate.Struct(res)
	return
}
