package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ScreenSeession struct {
	Id          int    `json:"id",number`
	SessionName string `json:"sessionName",string`
	Created     string `json:"created",string`
	Mode        string `json:"mode",string`
}

type ScreenSessions struct {
	ScreenSessionsSlice []ScreenSeession `json:"sessions",ScreenSeession`
}

func (ss *ScreenSessions) GetScreenSessions(scanner *bufio.Scanner) {
	doAReturn := false
	fullScan := ""
	for scanner.Scan() {
		fullScan += scanner.Text() + "\n"
		if fullScan == "There is a screen on:\n" || fullScan == "There are screens on:\n" {
			doAReturn = true
			fullScan = ""
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	if doAReturn {
		lines := strings.Split(fullScan, "\n")
		if len(lines) > 0 {
			lines = lines[:len(lines)-2]
		}

		for i := 0; i < len(lines); i++ {
			if lines[i] != "Remove dead screens with 'screen -wipe'." {
				session := ScreenSeession{}
				s := strings.Fields(lines[i])
				for j := 0; j < len(s); j++ {
					if j%5 == 0 {
						idAndName := strings.Split(s[j], ".")
						session.Id, _ = strconv.Atoi(idAndName[0])
						session.SessionName = idAndName[1]

					} else if j%5 == 1 {
						session.Created = s[j]
					} else if j%5 == 2 {
						session.Created += " " + s[j]
					} else if j%5 == 3 {
						session.Created += " " + s[j]
					} else if j%5 == 4 {
						session.Mode = s[j]
					}
				}
				ss.ScreenSessionsSlice = append(ss.ScreenSessionsSlice, session)
			}
		}
	}
	fmt.Println(fullScan)
}

func main() {
	cmd := exec.Command("screen", "-ls")
	out, err := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start err=%v", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(out)
	defer cmd.Wait()
	ss := ScreenSessions{}
	ss.GetScreenSessions(scanner)
	j, _ := json.Marshal(ss.ScreenSessionsSlice)
	j, _ = json.MarshalIndent(ss.ScreenSessionsSlice, "", "  ")
	fmt.Println(string(j))

	// Create a directory to store the screen sessions
	path := ".screensessions"
	wd, _ := os.Getwd()
	parent := filepath.Dir(wd)
	if path != parent {
		usersHomeDirectory, _ := os.UserHomeDir()
		os.Chdir(usersHomeDirectory)
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		os.Chdir(path)
	}
	_ = ioutil.WriteFile("screensessions.json", j, 0644)
}
