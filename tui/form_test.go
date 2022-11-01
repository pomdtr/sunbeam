package tui

import "testing"

func TestChecbox(t *testing.T) {
	checkbox := Checkbox{}
	if checked, _ := checkbox.Value().(bool); checked {
		t.Errorf("checkbox should be false by default")
	}
	checkbox.Toggle()
	if checked, _ := checkbox.Value().(bool); !checked {
		t.Errorf("checkbox should be true after toggle")
	}
}
