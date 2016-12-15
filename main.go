package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func FatalIf(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var (
		input  string
		output string
	)

	flag.StringVar(&input, "input", "", "input file, typically an auto-generated SQL script containing SQLCMD variables")
	flag.StringVar(&output, "output", "", "output file, SQLCMD translated to plain old SQL")
	flag.Parse()

	if len(input) == 0 || len(output) == 0 {
		log.Fatalln("Please provide both -input and -output")
	}

	TranslateFile(input, output)
}

const SetVarPrefix = ":setvar "

func TranslateFile(in, out string) {
	fin, err := os.Open(in)
	FatalIf(err)

	fout, err := os.Create(out)
	FatalIf(err)

	defer fin.Close()
	defer fout.Close()

	variables := make(map[string]string)

	s := bufio.NewScanner(fin)
	for s.Scan() {
		content := s.Text()
		line := content

		found := ScanVariablesInLine(line, variables)
		if found {
			line = ""
		}

		WriteLine(fout, line, variables)
	}
}

func ScanVariablesInLine(line string, variables map[string]string) bool {
	if !strings.HasPrefix(line, ":") {
		return false
	}

	if !strings.HasPrefix(line, SetVarPrefix) {
		return true
	}

	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, SetVarPrefix)
	parts := strings.SplitN(line, " ", 2)

	log.Printf("Found variable: %q\n", parts[0])
	variables["$("+parts[0]+")"] = strings.Trim(parts[1], `"`)

	return true
}

func WriteLine(fout io.Writer, line string, variables map[string]string) {
	if len(line) == 0 {
		fmt.Fprintf(fout, "\r\n")
		return
	}

	line = ReplaceVariablesInLine(line, variables)
	fmt.Fprintf(fout, "%s\r\n", line)
}

func ReplaceVariablesInLine(line string, variables map[string]string) string {
	for k, v := range variables {
		if strings.Contains(line, k) {
			line = strings.Replace(line, k, v, -1)
		}
	}

	return line
}
