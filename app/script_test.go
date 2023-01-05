package app

import "testing"

func TestRenderCommand(t *testing.T) {
	s := Script{
		Command: "ls -l ${{ directory }}",
	}

	expected := "ls -l ~"
	command, err := s.Cmd(map[string]any{
		"directory": "~",
	})

	if err != nil {
		t.Fatalf("An error occured: %s", err)
	}

	if command != expected {
		t.Fatalf("%s != %s", command, expected)
	}
}
