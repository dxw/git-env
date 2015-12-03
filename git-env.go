package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Option struct {
	Name     string
	Question string
	Default  string
}

type Config struct {
	Prod       string
	Other      []string
	ProdDeploy string
}

func LoadConfig_(getOption func(string) (string, error)) (*Config, error) {
	config := Config{}

	cfg := map[string]string{}

	for _, opt := range options {
		s, err := getOption(opt.Name)
		if err != nil {
			return nil, err
		}
		cfg[opt.Name] = s
	}
	config.Prod = cfg["prod"]
	config.Other = strings.Split(cfg["other"], " ")
	config.ProdDeploy = cfg["prod-deploy"]

	return &config, nil
}

func LoadConfig() (*Config, error) {
	return LoadConfig_(getOption)
}

func getOption(opt string) (string, error) {
	stdout, err := exec.Command("git", "config", "env-branch."+opt).Output()
	if err != nil {
		return "", errors.New("This repo isn't git env enabled. Run 'git env init' first.")
	}
	return string(stdout)[:len(stdout)-1], nil
}

func (c Config) IsEnv(branch string) bool {
	if branch == c.Prod {
		return true
	}
	for _, b := range c.Other {
		if branch == b {
			return true
		}
	}
	return false
}
func (c Config) IsProd(branch string) bool {
	return branch == c.Prod
}

func (c Config) ProdRemote() string {
	stdout, err := exec.Command("git", "config", "branch."+c.Prod+".remote").Output()
	if err != nil {
		log.Fatalf("Failed to get remote of %s branch.", c.Prod)
	}
	return string(stdout)[:len(stdout)-1]
}

var (
	VERSION = "1.2.0-dev"
	config  *Config
	options = []Option{
		{
			Name:     "prod",
			Question: "What is your production branch?",
			Default:  "master",
		},
		{
			Name:     "other",
			Question: "What other environment branches do you have?",
			Default:  "stage dev",
		},
		{
			Name:     "prod-deploy",
			Question: "What command should be run to deploy to the production branch?",
			Default:  "git checkout {{.env}} && git merge --no-ff {{.feature}}",
		},
	}
)

func main() {
	if len(os.Args) < 2 {
		help("")
		return
	}

	// Do help and init commands before loading config

	switch os.Args[1] {
	case "init":
		cmdInit()
		return
	case "help":
		if len(os.Args) > 2 {
			help(os.Args[2])
		} else {
			help("")
		}
		return
	case "version":
		fmt.Printf("git-env version: %s\n", VERSION)
		return
	}

	var err error
	config, err = LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// Do start and deploy commands only after loading config

	switch os.Args[1] {
	case "start":
		cmdStart(os.Args[2:])
	case "deploy":
		cmdDeploy(os.Args[2:])
	default:
		help("")
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

	// Check branch and origin/branch point at the same commit
	if gitRevParse(deployEnv) != gitRevParse(config.ProdRemote()+"/"+deployEnv) {
		log.Fatalf("Branch %s and branch %s/%s do not point at the same commit.", deployEnv, config.ProdRemote(), deployEnv)
	}

	// Rebase feature and env against upstream
	gitCommand("checkout", feature)
	gitCommand("pull", "--rebase", config.ProdRemote(), config.Prod)
	gitCommand("checkout", deployEnv)
	gitCommand("pull", "--rebase", config.ProdRemote(), deployEnv)

	if config.IsProd(deployEnv) {
		// In a production merge use --no-ff so the branch names are preserved
		gitCommand("checkout", feature)

		s := bytes.NewBufferString("")
		err := template.Must(template.New("").Parse(config.ProdDeploy)).Execute(s, map[string]string{"env": deployEnv, "feature": feature})
		if err != nil {
			panic(err)
		}

		runCommand("sh", "-c", s.String())
	} else {
		// In a non-production merge rebase against the remote env branch and merge
		gitCommand("merge", feature)
	}
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
		fmt.Println("  git env help                               - show this help")
		fmt.Println("  git env version                            - show git-env's version")
		fmt.Println("  git env init                               - configure which ENV branches are being used")
		fmt.Println("  git env start BRANCH_NAME                  - start a new feature branch")
		fmt.Println("  git env deploy ENV_BRANCH [FEATURE_BRANCH] - deploy a feature branch to an ENV branch (FEATURE_BRANCH defaults to current branch)")
	}
}

func gitCommand(args ...string) {
	runCommand("git", args...)
}

func runCommand(cmd string, args ...string) {
	fmt.Printf("+ %s %s\n", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	err := c.Run()

	if err != nil {
		log.Fatalf("Failed executing command: %s %s", cmd, strings.Join(args, " "))
	}
}

func gitBranch() (string, error) {
	stdout, err := exec.Command("git", "branch").Output()
	return string(stdout), err
}

func gitRevParse(branch string) string {
	stdout, err := exec.Command("git", "rev-parse", branch).Output()
	if err != nil {
		log.Fatalf("Failed executing command: git rev-parse %s", branch)
	}
	return string(stdout)
}

func getCurrentBranch_(gitBranch func() (string, error)) (string, error) {
	stdout, err := gitBranch()
	if err != nil {
		return "", err
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "* ") {
			items := strings.Split(line, " ")

			return items[1], nil
		}
	}

	return "", errors.New("could not detect current branch")
}

func getCurrentBranch() (string, error) {
	return getCurrentBranch_(gitBranch)
}
