package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vaughan0/go-ini"
)

func main() {
	if len(os.Args) < 2 {
		help()
	}

	switch os.Args[1] {
	case "init":
		_init()
	case "branch":
		branch(os.Args[2:])
	default:
		help()
	}
}

func help() {
	fmt.Println("TODO: implement help")
	os.Exit(1)
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
	if len(args) < 1 {
		help()
	}

	config, err := readConfig()
	if err != nil {
		panic(err)
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

func runCommand(cmd ...string) {
	err := exec.Command(cmd[0], cmd[1:]...).Run()
	if err != nil {
		log.Fatalf("Failed executing command: %#v\n", cmd)
	}
}

func readConfig() (map[string]string, error) {
	stdout, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		panic(err)
	}
	pth := string(stdout)[:len(stdout)-1]

	file, err := ini.LoadFile(filepath.Join(pth, ".git/config"))
	if err != nil {
		panic(err)
	}

	config := map[string]string(file.Section("env-branch"))
	return config, nil
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
