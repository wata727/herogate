package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppRun(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Stdout  string
	}{
		{
			Name:    "No subcommand",
			Command: "herogate",
			Stdout:  "Herogate",
		},
	}

	for _, tc := range cases {
		app := NewApp()
		args := strings.Split(tc.Command, " ")
		writer := new(bytes.Buffer)
		app.Writer = writer

		app.Run(args)

		if !strings.Contains(writer.String(), tc.Stdout) {
			t.Fatalf("\nBad: %s\nExpected Contains: %s\n\ntestcase: %s", writer.String(), tc.Stdout, tc.Name)
		}
	}
}
