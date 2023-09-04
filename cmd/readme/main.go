// Package main provides a helper for generating readme contents.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	usageStart = "<!-- readme-help -->"
	usageStop  = "<!-- readme-help end -->"
	readmeFile = "./README.md"
)

// treat like a const.
var bin = []string{"go", "run", "./cmd/token-forge/main.go"}

// Main.
func main() {
	fileBytes, err := os.ReadFile(readmeFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	readme := string(fileBytes)
	help := getHelp(bin...)

	usage := fmt.Sprintf("%s\n```\n%s```", usageStart, help)
	commands := getCmds(help)
	for _, c := range commands {
		hc := getHelp(append(bin, c)...)
		usage = fmt.Sprintf("%s\n```\n%s```", usage, string(hc))
	}

	preIndex := strings.Index(readme, usageStart)
	postIndex := strings.Index(readme, usageStop) + len(usageStop)
	if preIndex < 0 || postIndex < 0 {
		log.Fatalf("no usage anchor comments detected; must anchor w/ '%s' and '%s'", usageStart, usageStop)
	}

	usage = fmt.Sprintf("%s\n%s", usage, usageStop)
	readmePre := readme[:preIndex]
	readmePost := readme[postIndex:]
	newReadme := fmt.Sprintf("%s%s%s", readmePre, usage, readmePost)

	//#nosec:G306
	if err := os.WriteFile(readmeFile, []byte(newReadme), 0o644); err != nil {
		log.Fatalf("%v", err)
	}

	log.Println("new readme written successfully")
}

func getCmds(out []byte) []string {
	c := []string{}
	commandSection := false
	scanner := bufio.NewScanner(bytes.NewReader(out))
	r := regexp.MustCompile(`^\s{2}\S`) // two spaces only
	for scanner.Scan() {
		line := scanner.Text()
		if commandSection {
			switch {
			case len(strings.TrimSpace(line)) < 1:
				// blank line.
				continue
			case r.MatchString(line):
				// if we're in the command section and find a line that matches the
				// pattern (indicating we have something in the left column), then we
				// have a command.
				tmp := strings.Split(strings.TrimSpace(line), " ") // one space
				c = append(c, tmp[0])
			case !strings.HasPrefix(line, " "):
				// if we've already seen the entry of the command section, and we get
				// a line that's not indented, we've exited the command section.
				commandSection = false
			default:
				// ¯\_(ツ)_/¯
				continue
			}
		} else if strings.HasPrefix(line, "Commands:") {
			commandSection = true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("%v", err)
	}

	return c
}

func getHelp(c ...string) []byte {
	tmp := []string{}
	tmp = append(tmp, c...)
	tmp = append(tmp, "--help")
	cmd := exec.Command(tmp[0], tmp[1:]...) //#nosec:G204

	help, err := cmd.Output()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return help
}
