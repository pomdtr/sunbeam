package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"

	"github.com/adrg/xdg"
	"github.com/atotto/clipboard"
	"github.com/go-playground/validator"
	"github.com/skratchdot/open-golang/open"
	"tailscale.com/jsondb"
)

var CommandDir string

func init() {
	if commandDir := os.Getenv("SUNBEAM_COMMAND_DIR"); commandDir != "" {
		CommandDir = commandDir
	} else {
		CommandDir = xdg.DataHome
	}
}

type Command struct {
	Script
	Env   map[string]string
	Args  []string
	Input CommandInput
}

type CommandInput struct {
	Query   string `json:"query"`
	Form    any    `json:"form"`
	Storage any    `json:"storage"`
}

func (c Command) Run() (res ScriptResponse) {
	log.Printf("Running command %s with args %s", c.Script.Url.Path, c.Args)
	// Check if the number of arguments is correct
	if len(c.Args) < len(c.Metadatas.Arguments) {
		formItems := make([]FormItem, 0)
		for i := len(c.Args); i < len(c.Metadatas.Arguments); i++ {
			formItems = append(formItems, FormItem{
				Type: "text",
				Id:   c.Metadatas.Arguments[i].Placeholder,
				Name: c.Metadatas.Arguments[i].Placeholder,
			})
		}
		return ScriptResponse{
			Type: "form",
			Form: &FormResponse{
				Method: "args",
				Items:  formItems,
			},
		}
	}

	cmd := exec.Command(c.Script.Url.Path, c.Args...)
	cmd.Dir = path.Dir(cmd.Path)

	// Copy process environment
	cmd.Env = make([]string, len(os.Environ()))
	copy(cmd.Env, os.Environ())

	// Add custom environment variables to the process
	for key, value := range c.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add support dir to environment
	supportDir := path.Join(xdg.DataHome, "sunbeam", c.Script.Metadatas.PackageName, "support")
	cmd.Env = append(cmd.Env, fmt.Sprintf("SUNBEAM_SUPPORT_DIR=%s", supportDir))

	storagePath := path.Join(xdg.DataHome, "sunbeam", c.Script.Metadatas.PackageName, "storage.json")
	storage, err := jsondb.Open[any](storagePath)
	if err != nil {
		log.Printf("Unable to init storage: %s", err)
	}
	c.Input.Storage = &storage.Data

	var inbuf, outbuf, errbuf bytes.Buffer
	cmd.Stderr = &errbuf
	cmd.Stdout = &outbuf
	cmd.Stdin = &inbuf

	var bytes []byte
	bytes, err = json.Marshal(c.Input)
	if err != nil {
		err = fmt.Errorf("Error while marshalling input: %w", err)
		return
	}
	inbuf.Write(bytes)

	err = cmd.Run()

	if err != nil {
		return
	}

	if c.Metadatas.Mode != "interactive" {
		return ScriptResponse{
			Type: "detail",
		}
	}

	err = json.Unmarshal(outbuf.Bytes(), &res)
	if err != nil {
		return
	}
	err = Validator.Struct(res)
	if err != nil {
		return
	}

	if res.Storage != nil {
		storage.Data = &res.Storage
		err = storage.Save()
		if err != nil {
			return
		}
	}

	return
}

var Validator = validator.New()

type ScriptResponse struct {
	Type    string          `json:"type" validate:"required,oneof=list detail form action"`
	List    *ListResponse   `json:"list,omitempty"`
	Detail  *DetailResponse `json:"detail,omitempty"`
	Form    *FormResponse   `json:"form,omitempty"`
	Action  *ScriptAction   `json:"action,omitempty"`
	Storage any             `json:"storage,omitempty"`
}

type DetailResponse struct {
	Format  string         `json:"format"`
	Text    string         `json:"text"`
	Actions []ScriptAction `json:"actions"`
}

type FormResponse struct {
	Method string     `json:"dest" validate:"oneof=args stdin"`
	Items  []FormItem `json:"items"`
}

type FormItem struct {
	Type    string `json:"type" validate:"required,oneof=text password"`
	Id      string `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Default string `json:"value"`
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
	Title   string   `json:"title" validate:"required,oneof=copy open open-url"`
	Keybind string   `json:"keybind"`
	Type    string   `json:"type" validate:"required"`
	Path    string   `json:"path,omitempty"`
	Url     string   `json:"url,omitempty"`
	Content string   `json:"content,omitempty"`
	Args    []string `json:"args,omitempty"`
}

type Script struct {
	Url       url.URL
	Metadatas ScriptMetadatas
}

func (s Script) FilterValue() string { return s.Metadatas.Title }
func (s Script) Title() string       { return s.Metadatas.Title }
func (s Script) Description() string { return s.Metadatas.PackageName }

type ScriptMetadatas struct {
	SchemaVersion        int                 `json:"schemaVersion" validate:"required,eq=1"`
	Title                string              `json:"title" validate:"required"`
	Mode                 string              `json:"mode" validate:"required,oneof=interactive silent fullOutput"`
	PackageName          string              `json:"packageName" validate:"required"`
	Description          string              `json:"description,omitempty"`
	Arguments            []ScriptArgument    `json:"argument1,omitempty" validate:"omitempty,dive"`
	Environment          []ScriptEnvironment `json:"environment,omitempty" validate:"omitempty,dive"`
	Icon                 string              `json:"icon,omitempty"`
	CurrentDirectoryPath string              `json:"currentDirectoryPath,omitempty"`
	NeedsConfirmation    bool                `json:"needsConfirmation,omitempty"`
	Author               string              `json:"author,omitempty"`
	AuthorUrl            string              `json:"authorUrl,omitempty"`
}

type ScriptArgument struct {
	Type           string `json:"type" validate:"required,oneof=text"`
	Placeholder    string `json:"placeholder" validate:"required"`
	Optional       bool   `json:"optional"`
	PercentEncoded bool   `json:"percentEncoded"`
	Secure         bool   `json:"secure"`
}

func (s ScriptArgument) Title() string {
	return s.Placeholder
}

type ScriptEnvironment struct {
	Name string `json:"name" validate:"required"`
}

func (s ScriptEnvironment) Title() string {
	return s.Name
}

func extractSunbeamMetadatas(content string) ScriptMetadatas {
	r := regexp.MustCompile("@sunbeam.([A-Za-z0-9]+)\\s([\\S ]+)")
	groups := r.FindAllStringSubmatch(content, -1)

	metadataMap := make(map[string]string)
	for _, group := range groups {
		metadataMap[group[1]] = group[2]
	}

	metadatas := ScriptMetadatas{}
	json.Unmarshal([]byte(metadataMap["schemaVersion"]), &metadatas.SchemaVersion)
	json.Unmarshal([]byte(metadataMap["needsConfirmation"]), &metadatas.NeedsConfirmation)

	for _, key := range []string{"argument1", "argument2", "argument3"} {
		var argument ScriptArgument
		err := json.Unmarshal([]byte(metadataMap[key]), &argument)
		if err != nil {
			break
		}
		metadatas.Arguments = append(metadatas.Arguments, argument)
	}

	for _, key := range []string{"environment1", "environment2", "environment3"} {
		var environment ScriptEnvironment
		err := json.Unmarshal([]byte(metadataMap[key]), &environment)
		if err != nil {
			break
		}
		metadatas.Environment = append(metadatas.Environment, environment)
	}

	metadatas.Title = metadataMap["title"]
	metadatas.Mode = metadataMap["mode"]
	metadatas.PackageName = metadataMap["packageName"]
	metadatas.Icon = metadataMap["icon"]
	metadatas.CurrentDirectoryPath = metadataMap["currentDirectoryPath"]
	metadatas.Author = metadataMap["author"]
	metadatas.AuthorUrl = metadataMap["autorUrl"]
	metadatas.Description = metadataMap["description"]

	return metadatas
}

func Parse(script_path string) (Script, error) {
	content, err := ioutil.ReadFile(script_path)
	if err != nil {
		return Script{}, err
	}

	metadatas := extractSunbeamMetadatas(string(content))

	scriptURL := url.URL{
		Scheme: "file",
		Path:   script_path,
	}

	scripCommand := Script{Url: scriptURL, Metadatas: metadatas}

	err = Validator.Struct(scripCommand)
	if err != nil {
		log.Printf("Error while parsing script %s: %s", script_path, err)
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
			dirScripts, _ := ScanDir(path.Join(dirPath, file.Name()))
			scripts = append(scripts, dirScripts...)
		}

		// if script is not executable
		if file.Mode()&0111 == 0 {
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

func RunAction(action ScriptAction) (err error) {
	switch action.Type {
	case "open":
		err = open.Run(action.Path)
		return err
	case "open-url":
		err := open.Run(action.Url)
		return err
	case "copy":
		err = clipboard.WriteAll(action.Content)
		return err
	}
	return nil
}
