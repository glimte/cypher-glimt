package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func scanConfigForProfile(f string, p string) (bool, int) {
	createConfigFile(f)

	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("failed opening config file: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	var line int
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "[profile "+p) {
			return true, line
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}

	return false, 0
}

func loadProfileFromFile(f string, l int) {
	file, err := os.Open(f)
	if err != nil {
		log.Fatalf("failed loading profile from file: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	file.Close()

	// range the result, extract information about the profile
	result := txtlines[l+1 : l+4]
	for _, eachline := range result {
		newValue := strings.Split(eachline, "=")

		switch {
		case newValue[0] == "username":
			username = newValue[1]
		case newValue[0] == "password":
			password = newValue[1]
		case newValue[0] == "address":
			address = newValue[1]
		}
	}
}

func createProfileFolder(p string) string {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.MkdirAll(p, os.ModePerm)
		fmt.Println("config folder does not exist, creating....")
	}
	return p
}

func createDefaultConfig(f string) string {
	p, _ := scanConfigForProfile(f, "default")

	if p != true {

		profile := map[string]map[string]string{ // Created nested map
			"default": map[string]string{
				"username": "neo4j",
				"password": "neo4j",
				"address":  "bolt://localhost:7687",
			},
		}
		createConfigFile(f)
		writeProfile(f, profile)
	}
	return f
}

func createConfigFile(f string) {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		fmt.Println("config file does not exist, creating...")

		var tmp string
		ioutil.WriteFile(f, []byte(tmp), 0666)
	}
}

func writeProfile(f string, c map[string]map[string]string) {

	createConfigFile(f)

	p, err := os.OpenFile(f, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer p.Close()
	for key, value := range c {
		fmt.Fprintf(p, "[profile %v]\n", key)

		for key, value := range value {
			fmt.Fprintf(p, "%v=%v\n", key, value)
		}
		fmt.Fprintf(p, "\n")
	}
}
