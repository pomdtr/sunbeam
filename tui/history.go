package tui

import (
	"encoding/json"
	"os"
	"path"
	"time"
)

type History struct {
	filepath string
	history  map[string]int64
}

func (h *History) Add(key string) {
	h.history[key] = time.Now().Unix()
}

func (h *History) Get(key string) int64 {
	return h.history[key]
}

func (h *History) Save() error {
	if _, err := os.Stat(h.filepath); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(h.filepath), 0755); err != nil {
			return err
		}
	}

	bytes, err := json.Marshal(h.history)
	if err != nil {
		return err
	}

	if err := os.WriteFile(h.filepath, bytes, 0644); err != nil {
		return err
	}

	return nil
}

func LoadHistory(filepath string) (*History, error) {
	var history map[string]int64
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return &History{
			filepath: filepath,
			history:  make(map[string]int64),
		}, nil
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &history); err != nil {
		return nil, err
	}
	return &History{
		filepath: filepath,
		history:  history,
	}, nil
}
