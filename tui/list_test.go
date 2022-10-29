package tui

import "testing"

func TestItemView(t *testing.T) {
	type testCase struct {
		item  ListItem
		width int
		want  string
	}

	item := ListItem{
		Title:       "Title",
		Subtitle:    "Subtitle",
		Accessories: []string{"31*"},
	}
	cases := map[string]testCase{
		"no accessories": {
			item:  item,
			width: 0,
			want:  "",
		},
		"accessories truncated": {
			item:  item,
			width: 2,
			want:  "Ti",
		},
		"title truncated": {
			item:  item,
			width: 7,
			want:  "Tit 31*",
		},
		"subtitle truncated": {
			item:  item,
			width: 13,
			want:  "Title Sub 31*",
		},
		"list expanded": {
			item:  item,
			width: 20,
			want:  "Title Subtitle   31*",
		},
	}

	for key, c := range cases {
		t.Run(key, func(t *testing.T) {
			if c.width != len(c.want) {
				t.Errorf("test case width (%d) does not match expected length (%d)", c.width, len(c.want))
			}
			got := c.item.View(c.width)
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
