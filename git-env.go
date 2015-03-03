package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Option struct {
	Name     string
	Question string
	Default  string
}

var options = []Option{
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

type Config struct {
	Prod  string
	Other []string
	Types []string
}

var config Config

func main() {
	if len(os.Args) < 2 {
		help()
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "branch":
		readConfig()
		cmdBranch(os.Args[2:])
	default:
		help()
	}
}

// Commands

func cmdInit() {
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

func cmdBranch(args []string) {
	if len(args) < 1 {
		help()
	}

	switch args[0] {
	case "start":
		if len(args) < 2 {
			help()
		}
		newBranch := args[1]

		runCommand("git", "checkout", config["prod"])
		runCommand("git", "pull", "--rebase", "origin", config["prod"])
		runCommand("git", "checkout", "-b", newBranch)

	case "deploy":

		if len(args) < 2 {
			help()
		}

		deployEnv := args[1]
		var feature string

		if len(args) > 2 {
			feature = args[2]
		} else {
			feature, err = getCurrentBranch()
			if err != nil {
				panic(err)
			}
		}

		runCommand("git", "checkout", feature)
		runCommand("git", "pull", "--rebase", "origin", config["prod"])
		runCommand("git", "checkout", deployEnv)
		runCommand("git", "pull", "--rebase", "origin", deployEnv)
		runCommand("git", "merge", feature)
		runCommand("git", "push", "origin", deployEnv)

	default:
		help()
	}
}

// Everything else

func help() {
	fmt.Println("TODO: implement help")
	os.Exit(1)
}

func runCommand(cmd ...string) {
	err := exec.Command(cmd[0], cmd[1:]...).Run()
	if err != nil {
		log.Fatalf("Failed executing command: %#v\n", cmd)
	}
}

func readConfig() {
	config = Config{}

	cfg := map[string]string{}

	for _, opt := range options {
		stdout, err := exec.Command("git", "config", "env-branch."+opt.Name).Output()
		if err != nil {
			log.Fatalf("This repo isn't git env enabled. Run 'git env init' first.")
		}
		cfg[opt.Name] = string(stdout)[:len(stdout)-1]
	}
	config.Prod = cfg["prod"]
	config.Other = strings.Split(cfg["other"], " ")
	config.Types = strings.Split(cfg["types"], " ")
}

func getCurrentBranch() (string, error) {
	stdout, err := exec.Command("git", "branch").Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			items := strings.Split(line, " ")

			//TODO
			// return error if the branch is an environment

			return items[1], nil
		}
	}

	return "", errors.New("could not detect current branch")
}
