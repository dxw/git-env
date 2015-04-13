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
}

type Config struct {
	Prod  string
	Other []string
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

func (c Config) ProdRemote() string {
	stdout, err := exec.Command("git", "config", "branch."+c.Prod+".remote").Output()
	if err != nil {
		log.Fatalf("Failed to get remote of %s branch.", c.Prod)
	}
	return string(stdout)[:len(stdout)-1]
}

var config Config

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		cmdInit()
		return
	}

	config = LoadConfig()

	if len(os.Args) < 2 {
		help("")
	} else {

		switch os.Args[1] {
		case "start":
			cmdStart(os.Args[2:])
		case "deploy":
			cmdDeploy(os.Args[2:])
		case "help":
			if len(os.Args) > 2 {
				help(os.Args[2])
			} else {
				help("")
			}
		default:
			help("")
		}
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

func cmdStart(args []string) {
	if len(args) < 1 {
		help("start")
	}

	newBranch := args[0]

	gitCommand("checkout", config.Prod)
	gitCommand("pull", "--rebase", config.ProdRemote(), config.Prod)
	gitCommand("checkout", "-b", newBranch)
}

func cmdDeploy(args []string) {
	if len(args) < 1 {
		help("deploy")
	}

	deployEnv := args[0]
	var feature string
	var err error

	if len(args) > 1 {
		feature = args[1]
	} else {
		feature, err = getCurrentBranch()
		if err != nil {
			panic(err)
		}
	}

	if !config.IsEnv(deployEnv) {
		log.Fatalf("Branch %s is not an env branch. Can't merge a feature into it.", deployEnv)
	}

	if config.IsEnv(feature) {
		log.Fatalf("Branch %s is an env branch. Can't merge an env branch into another env branch.", feature)
	}

	gitCommand("checkout", feature)
	gitCommand("pull", "--rebase", config.ProdRemote(), config.Prod)
	gitCommand("checkout", deployEnv)
	gitCommand("pull", "--rebase", config.ProdRemote(), deployEnv)
	gitCommand("merge", feature)
	// Let's not push anything - leave that up to the developer
	// gitCommand("push", config.ProdRemote(), deployEnv)
}

// Everything else

func help(arg string) {
	switch arg {
	//TODO: add in-depth documentation for each command
	// case "init":
	// case "start":
	// case "deploy":
	default:
		fmt.Println("Commands:")
		fmt.Println("  git env help [COMMAND]                     - show help")
		fmt.Println("  git env init                               - configure which ENV branches are being used")
		fmt.Println("  git env start BRANCH_NAME                  - start a new feature branch")
		fmt.Println("  git env deploy ENV_BRANCH [FEATURE_BRANCH] - deploy a feature branch to an ENV branch (FEATURE_BRANCH defaults to current branch)")
	}
}

func gitCommand(args ...string) {
	fmt.Printf("+ git %s\n", strings.Join(args, " "))
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		log.Fatalf("Failed executing command: git %s", strings.Join(args, " "))
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

			return items[1], nil
		}
	}

	return "", errors.New("could not detect current branch")
}
