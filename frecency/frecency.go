package frecency

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

type Sorter struct {
	ExactQueryMatchWeight int
	FrecencyData
	RecentSelectionsMatchWeight int
}

type FrecencyData struct {
	Queries    map[string][]Entry `json:"queries"`
	Selections map[string]Entry   `json:"selections"`
}

func (data *FrecencyData) Inc(resultId string, query string) {
	if selection, ok := data.Selections[resultId]; ok {
		selection.TimesSelected++
		selection.SelectedAt = append(selection.SelectedAt, time.Now().Unix())
		if len(selection.SelectedAt) > 10 {
			nbEntries := len(selection.SelectedAt)
			selection.SelectedAt = selection.SelectedAt[nbEntries-10:]
		}

		data.Selections[resultId] = selection
	} else {
		data.Selections[resultId] = Entry{
			Id:            resultId,
			TimesSelected: 1,
			SelectedAt:    []int64{time.Now().Unix()},
		}
	}

	if query == "" {
		return
	}

	if entries, ok := data.Queries[query]; ok {
		found := false
		for i, e := range entries {
			if e.Id == resultId {
				found = true
				entries[i].TimesSelected++
				entries[i].SelectedAt = append(entries[i].SelectedAt, time.Now().Unix())

				if len(entries[i].SelectedAt) > 10 {
					nbEntries := len(entries[i].SelectedAt)
					entries[i].SelectedAt = entries[i].SelectedAt[nbEntries-10:]
				}
			}
		}

		if !found {
			data.Queries[query] = append(data.Queries[query], Entry{
				Id:            resultId,
				TimesSelected: 1,
				SelectedAt:    []int64{time.Now().Unix()},
			})
		}
	} else {
		data.Queries[query] = []Entry{
			{
				Id:            resultId,
				TimesSelected: 1,
				SelectedAt:    []int64{time.Now().Unix()},
			},
		}
	}

}

type Result interface {
	ID() string
}

type Entry struct {
	Id            string
	TimesSelected int     `json:"times_selected"`
	SelectedAt    []int64 `json:"selected_at"`
}

func (entry Entry) Score(currentTime time.Time) int {
	score := 0

	for _, selectedAt := range entry.SelectedAt {
		selectionTime := time.Unix(int64(selectedAt), 0)
		timeSinceSelection := currentTime.Sub(selectionTime)
		if timeSinceSelection < 3*time.Hour {
			score += 100
		} else if timeSinceSelection < 24*time.Hour {
			score += 80
		} else if timeSinceSelection < 3*24*time.Hour {
			score += 60
		} else if timeSinceSelection < 7*24*time.Hour {
			score += 30
		} else if timeSinceSelection < 14*24*time.Hour {
			score += 10
		}
	}

	if len(entry.SelectedAt) == 0 {
		return 0
	}
	return entry.TimesSelected * (score / len(entry.SelectedAt))
}

func NewSorter() *Sorter {
	return &Sorter{
		FrecencyData: FrecencyData{
			Queries:    make(map[string][]Entry),
			Selections: make(map[string]Entry),
		},
		ExactQueryMatchWeight:       100,
		RecentSelectionsMatchWeight: 33,
	}
}

func (f Sorter) Score(resultId string, query string) int {
	now := time.Now()
	frecencyForQuery, ok := f.Queries[query]
	if ok {
		for _, entry := range frecencyForQuery {
			if entry.Id != resultId {
				continue
			}

			if frecencyScore := f.ExactQueryMatchWeight * entry.Score(now); frecencyScore > 0 {
				return frecencyScore
			}
		}
	}

	selection, ok := f.Selections[resultId]
	if ok {
		if frecencyScore := f.RecentSelectionsMatchWeight * selection.Score(now); frecencyScore > 0 {
			return frecencyScore
		}
	}

	return 0
}

func (f Sorter) Sort(results []Result, query string) {
	sort.SliceStable(results, func(i, j int) bool {
		return f.Score(results[i].ID(), query) > f.Score(results[j].ID(), query)
	})
}

func Load(filepath string) (*Sorter, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	sorter := NewSorter()
	if err := json.Unmarshal(bytes, &sorter.FrecencyData); err != nil {
		return nil, err
	}

	return sorter, nil
}

func (f Sorter) Save(filepath string) (err error) {
	bytes, err := json.Marshal(f.FrecencyData)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath, bytes, 0644); err != nil {
		return err
	}

	return nil
}
