package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

var Path = filepath.Join(utils.CacheDir(), "history.json")

type History struct {
	entries map[string]int64
	path    string
}

func Load(historyPath string) (History, error) {
	bts, err := os.ReadFile(historyPath)
	if os.IsNotExist(err) {
		return History{
			entries: map[string]int64{},
			path:    historyPath,
		}, nil
	} else if os.IsNotExist(err) {
		return History{}, err
	}

	var entries map[string]int64
	if err := json.Unmarshal(bts, &entries); err != nil {
		return History{}, err
	}

	return History{
		entries: entries,
		path:    historyPath,
	}, nil
}

func (h History) Sort(items []sunbeam.ListItem) {
	sort.SliceStable(items, func(i, j int) bool {
		keyI := items[i].Id
		keyJ := items[j].Id

		return h.entries[keyI] > h.entries[keyJ]
	})
}

func (h History) Update(key string) {
	h.entries[key] = time.Now().Unix()
}

func (h History) Save() error {
	if h.path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return err
	}

	bts, err := json.MarshalIndent(h.entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(h.path, bts, 0644); err != nil {
		return err
	}

	return nil
}
