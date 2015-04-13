package main

import (
	"fmt"
	"testing"
)

func TestGetCurrentBranch(t *testing.T) {
	tests := map[string]string{
		"* cat\n  dog\n": "cat",
		"  cat\n  dog\n": "",
	}

	for k, v := range tests {
		branch, err := getCurrentBranch_(func() (string, error) {
			return k, nil
		})

		if v == "" {
			if err == nil {
				t.Error("expected error, didn't get one")
			}
		} else {
			if err != nil {
				t.Error(fmt.Errorf("got unexpected error: %s", err))
			}

			if branch != v {
				t.Errorf(`expected branch "%s", got branch "%s"`, v, branch)
			}
		}
	}
}
