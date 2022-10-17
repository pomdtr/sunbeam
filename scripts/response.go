package scripts

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/go-playground/validator"
)

var Validator = validator.New()

type ScriptResponse struct {
	Type   string          `json:"type" validate:"required,oneof=list detail form action"`
	List   *ListResponse   `json:"list,omitempty"`
	Detail *DetailResponse `json:"detail,omitempty"`
	Form   *FormResponse   `json:"form,omitempty"`
	Action *ScriptAction   `json:"action,omitempty"`
}

type DetailResponse struct {
	Title   string         `json:"title"`
	Format  string         `json:"format" validate:"required,oneof=text markdown"`
	Text    string         `json:"text"`
	Actions []ScriptAction `json:"actions"`
}

type FormResponse struct {
	Title  string     `json:"title"`
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
	Title                string        `json:"title"`
	SearchBarPlaceholder string        `json:"searchBarPlaceholder"`
	OnQueryChange        *ScriptAction `json:"onQueryChange,omitempty"`
	Items                []ScriptItem  `json:"items"`
}

type ScriptItem struct {
	Icon     string         `json:"icon"`
	RawTitle string         `json:"title" validate:"required"`
	Subtitle string         `json:"subtitle"`
	Fill     string         `json:"fill"`
	Actions  []ScriptAction `json:"actions" validate:"required,gte=1,dive"`
}

func (i ScriptItem) FilterValue() string { return i.RawTitle }
func (i ScriptItem) Title() string       { return i.RawTitle }
func (i ScriptItem) Description() string { return i.Subtitle }

type ScriptAction struct {
	Type     string   `json:"type" validate:"required,oneof=copy open url push"`
	RawTitle string   `json:"title"`
	Keybind  string   `json:"keybind"`
	Path     string   `json:"path,omitempty"`
	Push     bool     `json:"push,omitempty"`
	Command  []string `json:"command,omitempty"`
	Url      string   `json:"url,omitempty"`
	Content  string   `json:"content,omitempty"`
	Args     []string `json:"args,omitempty"`
}

func (a ScriptAction) Title() string {
	if a.RawTitle != "" {
		return a.RawTitle
	}
	switch a.Type {
	case "open":
		return "Open"
	case "open-url":
		return "Open in Browser"
	case "copy":
		return "Copy to Clibpoard"
	case "push":
		return "Switch Page"
	default:
		return ""
	}
}

func (a ScriptAction) Description() string {
	return a.Keybind
}

func (a ScriptAction) FilterValue() string {
	return a.Title()
}

type Script struct {
	Url       url.URL
	Metadatas ScriptMetadatas
}

func (s Script) FilterValue() string { return s.Metadatas.Title }
func (s Script) Title() string       { return s.Metadatas.Title }
func (s Script) Description() string { return s.Metadatas.PackageName }

func (s Script) RequiredArguments() []ScriptArgument {
	var res []ScriptArgument
	for _, arg := range s.Metadatas.Arguments {
		if !arg.Optional {
			res = append(res, arg)
		}
	}
	return res
}

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
	r := regexp.MustCompile(`@sunbeam.([A-Za-z0-9]+)\s([\S ]+)`)
	groups := r.FindAllStringSubmatch(content, -1)

	metadataMap := make(map[string]string)
	for _, group := range groups {
		metadataMap[group[1]] = group[2]
	}

	metadatas := ScriptMetadatas{}
	_ = json.Unmarshal([]byte(metadataMap["schemaVersion"]), &metadatas.SchemaVersion)
	_ = json.Unmarshal([]byte(metadataMap["needsConfirmation"]), &metadatas.NeedsConfirmation)

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
	content, err := os.ReadFile(script_path)
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
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return []Script{}, err
	}

	scripts := []Script{}
	for _, file := range files {
		if file.IsDir() {
			dirScripts, _ := ScanDir(path.Join(dirPath, file.Name()))
			scripts = append(scripts, dirScripts...)
		}

		if fileinfo, err := file.Info(); err != nil || fileinfo.Mode()&0111 == 0 {
			log.Printf("%s is not executable", file.Name())
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
