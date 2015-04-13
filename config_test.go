package main

import "testing"

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig_(func(opt string) (string, error) {
		switch opt {
		case "prod":
			return "master", nil
		case "other":
			return "stage dev", nil
		}
		return "", nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if config.Prod != "master" {
		t.Error("config.Prod not set correctly")
	}

	if len(config.Other) != 2 || config.Other[0] != "stage" || config.Other[1] != "dev" {
		t.Errorf("config.Others not set correctly: %#v", config.Other)
	}
}

func TestConfigIsEnv(t *testing.T) {
	config := Config{
		Prod:  "master",
		Other: []string{"stage", "dev"},
	}

	tests := map[string]bool{
		"master":      true,
		"stage":       true,
		"dev":         true,
		"master/":     false,
		"cat":         false,
		"feature/123": false,
	}

	for k, v := range tests {
		if v {
			if !config.IsEnv(k) {
				t.Errorf(`branch "%s" not reported as env branch when it should be`, k)
			}
		} else {
			if config.IsEnv(k) {
				t.Errorf(`branch "%s" reported as env branch when it shouldn't be`, k)
			}
		}
	}
}

func TestConfigProdRemote(t *testing.T) {
	t.Skip("this function uses exec.Command")
}
