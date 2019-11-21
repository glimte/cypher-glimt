package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/urfave/cli"
)

// var profile string
var app = cli.NewApp()
var profileName, username, password, address string
var file, command string

func main() {
	cypherExec := checkPreRequisits("cypher-shell")
	catExec := checkPreRequisits("cat")

	// Identify home folder and make sure we have a place to store our config
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
	}
	configFolder := createProfileFolder(filepath.Join(homePath, "/", ".cg"))

	// load app information, global options, commands and config
	info()
	appConfig(configFolder)

	f := createDefaultConfig(filepath.Join(configFolder, "/", "config"))

	app.Action = func(c *cli.Context) error {
		if c.String("profile") > "" {
			profileName = c.String("profile")
			_, l := scanConfigForProfile(f, profileName)
			loadProfileFromFile(f, l)

		} else {
			profileName = "default"
			_, l := scanConfigForProfile(f, profileName)
			loadProfileFromFile(f, l)
		}

		if c.String("file") > "" {
			file = c.String("file")
		} else {
			file = ""
		}

		if c.String("command") > "" {
			command = c.String("command")
		} else {
			command = ""
		}
		//Require a either file or command flag to not terminate.
		if command == "" && file == "" {
			log.Fatalf("\nCommand or File input is required....\nRun with flag --help to review global options")
		}
		// Fetch username and password as input, this allows to not only refer a profile.
		if c.String("username") > "" {
			username = c.String("username")
		}
		if c.String("password") > "" {
			password = c.String("password")
		}
		if c.String("address") > "" {
			address = c.String("address")
		}
		return nil
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	csArgs := []string{
		"-u" + username,
		"-p" + password,
		"-a" + address}

	if file != "" {
		csArgs = append(csArgs, "-f"+file)
		if err := execute(&b,
			exec.Command(catExec, file),
			exec.Command(cypherExec, csArgs...),
		); err != nil {
			log.Fatalln(err)
		}
		io.Copy(os.Stdout, &b)
		os.Exit(0)
	}

	if command != "" {
		var b bytes.Buffer

		if err := execute(&b,
			exec.Command("echo", command),
			exec.Command(cypherExec, csArgs...),
		); err != nil {
			log.Fatalln(err)
		}
		io.Copy(os.Stdout, &b)
		os.Exit(0)
	}
}

func execute(outputBuffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var errorBuffer bytes.Buffer
	pipeStack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		stack[i].Stdout = stdoutPipe
		stack[i].Stderr = &errorBuffer
		stack[i+1].Stdin = stdinPipe
		pipeStack[i] = stdoutPipe
	}
	stack[i].Stdout = outputBuffer
	stack[i].Stderr = &errorBuffer
	if err := call(stack, pipeStack); err != nil {
		log.Fatalln(string(errorBuffer.Bytes()), err)
	}
	return err
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	// Adding fmt.Println to see what commands are running
	// fmt.Println(stack)
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

func checkPreRequisits(e string) string {
	p, err := exec.LookPath(e)
	if err != nil {
		log.Fatal(err)
		log.Printf("Please install missing prereq %v: ", e)
	}

	return p
}

func info() {

	cli.AppHelpTemplate = fmt.Sprintf(`%s
WEBSITE: https://glimte.com
`, cli.AppHelpTemplate)

	app.Name = "Cypher-glimt"
	app.Usage = `is a thin wrapper for cypher-shell, providing extra tools to make easier to work a cross,
   and with multiple environments using Neo4j and cypher-shell. Cypher-glimt does also make is possible
   to introduce pipelines and to allow for management of Neo4j with Continuous deployment.`
	app.UsageText = "cypher-glimt [command] / [global options]"
	app.Author = "Glimte"
	app.Version = "0.0.1"
}

func appConfig(cf string) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "profile",
			Value: "default",
			Usage: "Use a specific profile from your config or credential file",
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "",
			Usage: "File or files to load and parse with cypher-shell",
		},
		cli.StringFlag{
			Name:  "command, c",
			Value: "",
			Usage: "Command to pass along to cypher-shell",
		},
		cli.StringFlag{
			Name:  "username, u",
			Value: "",
			Usage: "Username to connect as",
		},
		cli.StringFlag{
			Name:  "password, p",
			Value: "",
			Usage: "Password to connect with.",
		},
		cli.StringFlag{
			Name:  "address, a",
			Value: "",
			Usage: "Address, port and protocol to connect to (bolt://localhost:7687)",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "setup cypher-glimt and configure profile values",

			Action: func(c *cli.Context) {
				fmt.Printf("Profilename: ")
				fmt.Scan(&profileName)
				fmt.Printf("Username: ")
				fmt.Scan(&username)
				fmt.Printf("Password: ")
				fmt.Scan(&password)
				fmt.Printf("Address: ")
				fmt.Scan(&address)

				profile := map[string]map[string]string{
					profileName: map[string]string{
						"username": username,
						"password": password,
						"address":  address,
					},
				}
				writeProfile(filepath.Join(cf, "/", "config"), profile)

			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
}
