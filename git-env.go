package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		_init()
	case "branch":
		branch(os.Args[3:])
	default:
		help()
	}
}

func help() {
	fmt.Println("TODO: implement help")
}

type Option struct {
	Name     string
	Question string
	Default  string
}

func _init() {
	options := []Option{
		{
			Name:     "prod",
			Question: "What is your production environment branch?",
			Default:  "master",
		},
		{
			Name:     "other",
			Question: "What other environment branches do you have?",
			Default:  "stage dev",
		},
		{
			Name:     "types",
			Question: "What branch types do you have?",
			Default:  "feature hotfix",
		},
	}

	values := map[string]string{}
	reader := bufio.NewReader(os.Stdin)

	for _, opt := range options {
		fmt.Printf("%s [%s] ", opt.Question, opt.Default)
		value, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if value == "\n" {
			values[opt.Name] = opt.Default
		} else {
			values[opt.Name] = value[:len(value)-1]
		}
	}

	for k, v := range values {
		err := exec.Command("git", "config", "--local", "--replace-all", "env-branch."+k, v).Run()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("You're ready to go.")
}

func branch(args []string) {
}
