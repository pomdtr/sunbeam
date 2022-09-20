package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/adrg/xdg"
	"github.com/go-playground/validator"
)

var CommandDirs []string

func init() {
	if sunbeamCommandDir := os.Getenv("SUNBEAM_COMMAND_DIR"); sunbeamCommandDir != "" {
		CommandDirs = append(CommandDirs, sunbeamCommandDir)
	}
	dataDirs := xdg.DataDirs
	dataDirs = append(dataDirs, xdg.DataHome)
	for _, dataDir := range dataDirs {
		commandDir := path.Join(dataDir, "sunbeam", "commands")
		CommandDirs = append(CommandDirs, commandDir)
	}
}

type any interface{}

type Command struct {
	Script
	Args  []string
	Input CommandInput
}

type CommandInput struct {
	Query  string `json:"query"`
	Params any    `json:"params"`
}

func NewCommand(script Script, args ...string) Command {
	return Command{
		Script: script,
		Args:   args,
	}
}

func (c Command) Run() (res ScriptResponse, err error) {
	err = os.Chmod(c.Script.Path, 0755)
	if err != nil {
		return
	}

	cmd := exec.Command(c.Script.Path, c.Args...)
	cmd.Dir = path.Dir(cmd.Path)

	// Copy process environment
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())

	// Add support dir to environment
	supportDir := path.Join(xdg.StateHome, "sunbeam", c.Script.Metadatas.PackageName)
	cmd.Env = append(cmd.Env, fmt.Sprintf("SUNBEAM_SUPPORT_DIR=%s", supportDir))

	var inbuf, outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	cmd.Stdin = &inbuf

	var bytes []byte
	bytes, err = json.Marshal(c.Input)
	log.Printf("Command input: %s", string(bytes))
	if err != nil {
		err = fmt.Errorf("Error while marshalling input: %w", err)
		return
	}
	inbuf.Write(bytes)

	err = cmd.Run()

	if err != nil {
		err = fmt.Errorf("%s: %s", err, errbuf.String())
		return
	}

	log.Printf("Command output: %s", outbuf.String())
	json.Unmarshal(outbuf.Bytes(), &res)
	err = validate.Struct(res)

	return
}

var validate = validator.New()

type ScriptResponse struct {
	Type    string         `json:"type" validate:"required,oneof=list detail form exit"`
	List    ListResponse   `json:"list" validate:"dive"`
	Detail  DetailResponse `json:"detail" validate:"dive"`
	Form    FormResponse   `json:"form" validate:"dive"`
	Storage any            `json:"storage"`
}

type DetailResponse struct {
	Markdown string `json:"markdown"`
}

type FormResponse struct {
}

type ListResponse struct {
	Title string       `json:"title"`
	Items []ScriptItem `json:"items"`
}

type ScriptItem struct {
	Icon       string         `json:"icon"`
	TitleField string         `json:"title" validate:"required"`
	Subtitle   string         `json:"subtitle"`
	Fill       string         `json:"fill"`
	Actions    []ScriptAction `json:"actions" validate:"required,gte=1,dive"`
}

func (i ScriptItem) FilterValue() string { return i.TitleField }
func (i ScriptItem) Title() string       { return i.TitleField }
func (i ScriptItem) Description() string { return i.Subtitle }

type ScriptAction struct {
	Title   string `json:"title" validate:"required,oneof=copy open open-url"`
	Keybind string `json:"keybind"`
	Type    string `json:"type" validate:"required"`
	Path    string `json:"path"`
	Url     string `json:"url"`
	Content string `json:"content"`
	Params  any    `json:"params"`
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

func extractSunbeamMetadatas(content string) map[string]string {
	r := regexp.MustCompile("@sunbeam.([A-Za-z0-9]+)\\s([\\S ]+)")
	groups := r.FindAllStringSubmatch(content, -1)

	metadataMap := make(map[string]string)
	for _, group := range groups {
		metadataMap[group[1]] = group[2]
	}

	return metadataMap
}

func Parse(script_path string) (Script, error) {
	content, err := ioutil.ReadFile(script_path)
	if err != nil {
		return Script{}, err
	}

	metadatas := extractSunbeamMetadatas(string(content))

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

		script, err := Parse(path.Join(dirPath, file.Name()))
		if err != nil {
			continue
		}

		scripts = append(scripts, script)
	}

	return scripts, nil
}
