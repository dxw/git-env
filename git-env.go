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

func LoadConfig() Config {
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

	return config
}

func (c Config) IsEnv(branch string) bool {
	if branch == config.Prod {
		return true
	}
	for _, b := range config.Other {
		if branch == b {
			return true
		}
	}
	return false
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
		config = LoadConfig()
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

		gitCommand("checkout", config.Prod)
		gitCommand("pull", "--rebase", "origin", config.Prod)
		gitCommand("checkout", "-b", newBranch)

	case "deploy":

		if len(args) < 2 {
			help()
		}

		deployEnv := args[1]
		var feature string
		var err error

		if len(args) > 2 {
			feature = args[2]
		} else {
			feature, err = getCurrentBranch()
			if err != nil {
				panic(err)
			}
		}

		if !config.IsEnv(deployEnv) {
			log.Fatalf("Branch %s is not an env branch. Can't merge a feature into it.\n", deployEnv)
		}

		if config.IsEnv(feature) {
			log.Fatalf("Branch %s is an env branch. Can't merge an env branch into another env branch.\n", deployEnv)
		}

		gitCommand("checkout", feature)
		gitCommand("pull", "--rebase", "origin", config.Prod)
		gitCommand("checkout", deployEnv)
		gitCommand("pull", "--rebase", "origin", deployEnv)
		gitCommand("merge", feature)
		gitCommand("push", "origin", deployEnv)

	default:
		help()
	}
}

// Everything else

func help() {
	fmt.Println("TODO: implement help")
	os.Exit(1)
}

func gitCommand(args ...string) {
	fmt.Printf("+ git %s\n", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		log.Fatalf("Failed executing command: git %s\n", strings.Join(args, " "))
	}
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
