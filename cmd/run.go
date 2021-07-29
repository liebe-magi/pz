/*
Copyright © 2021 MagicalLiebe <magical.liebe@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/atotto/clipboard"
	"github.com/mgutz/ansi"
	"github.com/sclevine/agouti"
	"github.com/spf13/cobra"
)

const (
	stdoutColor = "green"
	stderrColor = "red"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Access to paiza.jp",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := getConfigDir()
		if err != nil {
			log.Fatalln(err)
		}
		var temp string
		if len(args) == 1 {
			temp = args[0]
		} else {
			config, err := getConfig()
			if err != nil {
				log.Fatalln(err)
			}
			temp = config.Setting.DefalutTemp
		}
		tempDir := filepath.Join(configDir, temp)
		if f, err := os.Stat(tempDir); os.IsNotExist(err) || !f.IsDir() {
			log.Fatalf("[ERR] Template: %v is not found\n", args[0])
		}
		if err := run(temp); err != nil {
			log.Fatalln(err)
		}
	},
}

func run(temp string) error {
	url := "https://paiza.jp/sign_in"
	driver := agouti.ChromeDriver()
	if err := driver.Start(); err != nil {
		return err
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		log.Printf("%v", err)
		return err
	}

	if err = page.Navigate(url); err != nil {
		log.Printf("%v", err)
		return err
	}

	if err = login(page); err != nil {
		log.Printf("%v", err)
		return err
	}

	if err = interactive(page, temp); err != nil {
		return err
	}
	return nil
}

func login(p *agouti.Page) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	email := p.FindByID("email")
	pass := p.FindByID("password")

	email.Fill(config.User.Email)
	pass.Fill(config.User.Pass)

	if err := p.FirstByClass("a-button-primary-large").Submit(); err != nil {
		return err
	}

	return nil
}

func interactive(p *agouti.Page, temp string) error {
	stdin := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for stdin.Scan() {
		text := stdin.Text()
		textSplit := strings.Split(text, " ")
		if text == "get" || text == "g" {
			if err := get(p, temp); err != nil {
				fmt.Println(err)
			}
		} else if text == "test" || text == "t" {
			quesID, _, _, err := getQuestion(p)
			if err != nil {
				fmt.Println(err)
			} else {
				if err := test(temp, quesID); err != nil {
					fmt.Println(err)
				}
			}
		} else if textSplit[0] == "debug" || textSplit[0] == "d" {
			quesID, _, _, err := getQuestion(p)
			if err != nil {
				fmt.Println(err)
			} else {
				if len(textSplit) != 2 {
					fmt.Println("  [ERR] Select a number that you want to debug.")
				} else {
					programID, err := strconv.Atoi(textSplit[1])
					if err != nil {
						fmt.Println(err)
					}
					if err := debug(temp, quesID, programID); err != nil {
						fmt.Println(err)
					}
				}
			}
		} else if text == "submit" || text == "s" {
			quesID, _, _, err := getQuestion(p)
			if err != nil {
				fmt.Println(err)
			} else {
				if err := submit(p, temp, quesID); err != nil {
					fmt.Println(err)
				}
			}
		} else if text == "exit" || text == "e" {
			break
		} else if text == "" {
		} else {
			fmt.Printf("  [ERR] Unknown Command: %v\n", text)
		}
		fmt.Print("> ")
	}
	return nil
}

func get(p *agouti.Page, temp string) error {
	quesID, input, output, err := getQuestion(p)
	if err != nil {
		return err
	}
	fmt.Printf("  Get sample: input(%v) output(%v)\n", len(input), len(output))
	if err = downloadSample(quesID, input, output); err != nil {
		return err
	}
	if temp != "" {
		tempConfig, tempDir, err := getTemplateConfig(temp)
		if err != nil {
			return err
		}

		srcPath := filepath.Join(tempDir, tempConfig.File)
		src, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer src.Close()

		quesPath, err := getQuesDir(quesID)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(quesPath, tempConfig.File)
		if !Exists(dstPath) {
			fmt.Printf("  Copy template: %v\n", tempConfig.File)
			dst, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				return err
			}
		}
	}
	return nil
}

func getTemplateConfig(temp string) (templateConfig, string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return templateConfig{}, "", err
	}
	tempDir := filepath.Join(configDir, temp)
	tempConfigPath := filepath.Join(tempDir, "template.toml")

	var tempConfig templateConfig
	if _, err = toml.DecodeFile(tempConfigPath, &tempConfig); err != nil {
		return templateConfig{}, "", err
	}
	return tempConfig, tempDir, nil
}

func getQuestion(p *agouti.Page) (string, []string, []string, error) {
	dom, err := p.HTML()
	if err != nil {
		return "", []string{}, []string{}, err
	}
	contents := strings.NewReader(dom)
	contentsDom, err := goquery.NewDocumentFromReader(contents)
	quesID, err := p.FindByClass("section3").First("h2").Text()
	if err != nil {
		quesID, err = p.FindByID("tab-problem").First("h2").Text()
		if err != nil {
			return "", []string{}, []string{}, fmt.Errorf("  [ERR] This page is not contain a question.")
		}
	}
	quesID = strings.Split(quesID, ":")[0]
	input := []string{}
	output := []string{}
	contentsDom.Find("pre[class=sample-content__input]").Each(func(i int, s *goquery.Selection) {
		if i%2 == 0 {
			input = append(input, strings.TrimSpace(s.Text()))
		} else {
			output = append(output, strings.TrimSpace(s.Text()))
		}
	})
	url, err := p.URL()
	if err != nil {
		return "", []string{}, []string{}, err
	}
	fmt.Printf("  ID: %v (%v)\n", quesID, url)
	return quesID, input, output, nil
}

func downloadSample(quesID string, input, output []string) error {
	quesDir, err := getQuesDir(quesID)
	p := filepath.Join(quesDir, "tests")

	// mkdir: ./[rank]/[question]/tests
	if f, err := os.Stat(p); os.IsNotExist(err) || !f.IsDir() {
		err = os.MkdirAll(p, 0755)
	}

	// write: input
	for i, s := range input {
		in := filepath.Join(p, fmt.Sprintf("input_%v.txt", i))
		if err = writeFile(in, s); err != nil {
			return err
		}
	}

	// write: output
	for i, s := range output {
		out := filepath.Join(p, fmt.Sprintf("output_%v.txt", i))
		if err = writeFile(out, s); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(filename string, content string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	b := []byte(content)
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func runCommand(cmd *exec.Cmd, verbose bool) (stdout, stderr string, exitCode int, err error) {
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	var bufout, buferr bytes.Buffer
	outReader2 := io.TeeReader(outReader, &bufout)
	errReader2 := io.TeeReader(errReader, &buferr)

	if err = cmd.Start(); err != nil {
		return
	}

	go printOutputWithHeader("    >> ", stdoutColor, outReader2, verbose)
	go printOutputWithHeader("    >> ", stderrColor, errReader2, verbose)

	err = cmd.Wait()

	stdout = bufout.String()
	stderr = buferr.String()

	if err != nil {
		if err2, ok := err.(*exec.ExitError); ok {
			if s, ok := err2.Sys().(syscall.WaitStatus); ok {
				err = nil
				exitCode = s.ExitStatus()
			}
		}
	}
	return
}

func printOutputWithHeader(header, color string, r io.Reader, verbose bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if verbose {
			fmt.Printf("%s%s\n", header, ansi.Color(scanner.Text(), color))
		}
	}
}

func test(temp, quesID string) error {
	if temp == "" {
		return fmt.Errorf("  [ERR] Select a template.")
	}
	tempConfig, _, err := getTemplateConfig(temp)
	if err != nil {
		return err
	}
	quesDir, err := getQuesDir(quesID)
	if err != nil {
		return err
	}
	programPath := filepath.Join(quesDir, tempConfig.File)
	cmdStrings := strings.Replace(tempConfig.Run, tempConfig.File, programPath, 1)
	if err != nil {
		return err
	}
	testDir := filepath.Join(quesDir, "tests")
	files, err := ioutil.ReadDir(testDir)
	sampleSize := 0
	for _, file := range files {
		if strings.Contains(file.Name(), "input_") && strings.Contains(file.Name(), ".txt") {
			s := strings.Replace(strings.Replace(file.Name(), "input_", "", 1), ".txt", "", 1)
			n, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			n++
			if n > sampleSize {
				sampleSize = n
			}
		}
	}
	count := 0
	for i := 0; i < sampleSize; i++ {
		fmt.Printf("  Test: case %v\n", i+1)
		inputPath := filepath.Join(testDir, fmt.Sprintf("input_%v.txt", i))
		input, err := ioutil.ReadFile(inputPath)
		if err != nil {
			return err
		}
		cmd, err := convCmd(cmdStrings)
		cmd.Stdin = bytes.NewBufferString(string(input))
		start := time.Now()
		stdout, _, _, err := runCommand(cmd, false)
		if err != nil {
			return err
		}
		end := time.Now()
		outputPath := filepath.Join(testDir, fmt.Sprintf("output_%v.txt", i))
		output, err := ioutil.ReadFile(outputPath)
		if err != nil {
			return err
		}
		outputList := strings.Split(string(output), "\n")
		stdoutList := strings.Split(strings.TrimSpace(stdout), "\n")
		check := true
		if len(outputList) != len(stdoutList) {
			check = false
		} else {
			for i := 0; i < len(outputList); i++ {
				if outputList[i] != stdoutList[i] {
					check = false
				}
			}
		}
		var outText string
		if check {
			outText = fmt.Sprintf("    [Success] Passed the test. (time:%vs)", (end.Sub(start)).Seconds())
			outText = ansi.Color(outText, "green")
			count++
		} else {
			outText = "    [Falure] Did not pass the test."
			outText = ansi.Color(outText, "red")
		}
		fmt.Println(outText)
	}
	var t string
	if count == sampleSize {
		t = ansi.Color("All clear!", "green")
	} else {
		t = ansi.Color(fmt.Sprintf("%v failed...", sampleSize-count), "red")
	}
	fmt.Printf("  [Result] Pass %v/%v (%v)\n", count, sampleSize, t)
	return nil
}

func getQuesDir(quesID string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	rank := string([]rune(quesID)[0])
	quesDir := filepath.Join(pwd, rank, quesID)
	return quesDir, nil
}

func convCmd(cmdText string) (*exec.Cmd, error) {
	cmdStrings := strings.Split(cmdText, " ")
	var cmd *exec.Cmd
	if len(cmdStrings) == 0 {
		return cmd, fmt.Errorf("  [ERR] Set a execution command to template.toml")
	} else if len(cmdStrings) == 1 {
		cmd = exec.Command(cmdStrings[0])
	} else {
		cmd = exec.Command(cmdStrings[0], cmdStrings[1:]...)
	}
	return cmd, nil
}

func submit(p *agouti.Page, temp, quesID string) error {
	if temp == "" {
		return fmt.Errorf("  [ERR] Select a template.")
	}
	tempConfig, _, err := getTemplateConfig(temp)
	if err != nil {
		return err
	}
	langSelect := p.FindByID("language_id")
	if err := langSelect.Select(tempConfig.Lang); err != nil {
		return err
	}
	quesDir, err := getQuesDir(quesID)
	if err != nil {
		return err
	}
	programPath := filepath.Join(quesDir, tempConfig.File)
	program, err := ioutil.ReadFile(programPath)
	if err != nil {
		return err
	}
	programText := strings.TrimSpace(string(program))
	if err := clipboard.WriteAll(programText); err != nil {
		return err
	}
	fmt.Println("  Copied to the clipboard.")
	return nil
}

func debug(temp, quesID string, programID int) error {
	if temp == "" {
		return fmt.Errorf("  [ERR] Select a template.")
	}
	tempConfig, _, err := getTemplateConfig(temp)
	if err != nil {
		return err
	}
	quesDir, err := getQuesDir(quesID)
	if err != nil {
		return err
	}
	programPath := filepath.Join(quesDir, tempConfig.File)
	cmdStrings := strings.Replace(tempConfig.Run, tempConfig.File, programPath, 1)
	if err != nil {
		return err
	}
	testDir := filepath.Join(quesDir, "tests")
	files, err := ioutil.ReadDir(testDir)
	sampleSize := 0
	for _, file := range files {
		if strings.Contains(file.Name(), "input_") && strings.Contains(file.Name(), ".txt") {
			s := strings.Replace(strings.Replace(file.Name(), "input_", "", 1), ".txt", "", 1)
			n, err := strconv.Atoi(s)
			if err != nil {
				return err
			}
			n++
			if n > sampleSize {
				sampleSize = n
			}
		}
	}
	if 0 <= programID-1 && programID <= sampleSize {
		fmt.Printf("  Debug: case %v\n", programID)
		inputPath := filepath.Join(testDir, fmt.Sprintf("input_%v.txt", programID-1))
		input, err := ioutil.ReadFile(inputPath)
		if err != nil {
			return err
		}
		cmd, err := convCmd(cmdStrings)
		cmd.Stdin = bytes.NewBufferString(string(input))
		start := time.Now()
		_, _, _, err = runCommand(cmd, true)
		if err != nil {
			return err
		}
		end := time.Now()
		fmt.Printf("  [Finish] time:%vs\n", (end.Sub(start)).Seconds())
	} else {
		return fmt.Errorf("  [ERR] Could not find a sample: %v", programID)
	}
	return nil
}
func init() {
	rootCmd.AddCommand(runCmd)
}
