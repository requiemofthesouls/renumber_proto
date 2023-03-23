package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var usage = func() {
	fmt.Printf("Usage: renumber_proto [path to proto directory]")
	flag.PrintDefaults()
}

func main() {
	// validate arguments
	flag.Usage = usage

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	protoDir := os.Args[1]

	// validate directory path
	info, err := os.Stat(protoDir)
	if os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", protoDir)
	}
	if !info.IsDir() {
		log.Fatalf("Path is not a directory: %s", protoDir)
	}

	// walk the directory
	err = filepath.Walk(protoDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
				log.Println(path)

				var rawProto []byte
				if rawProto, err = os.ReadFile(path); err != nil {
					return err
				}

				var (
					content       = string(rawProto)
					renumberedAll = renumberAllMessages(content)
				)
				if err = os.WriteFile(path, []byte(renumberedAll), 0644); err != nil {
					return err
				}
			}

			return nil
		})

	if err != nil {
		log.Fatalf("Failed to walk directory: %s", err)
	}
}

func renumberAllMessages(text string) string {
	var messages = regexp.MustCompile(`(?sm)^message.*?{$.*?^}`).FindAllString(text, -1)

	for _, msg := range messages {
		renumbered := renumberFields(msg)
		text = strings.ReplaceAll(text, msg, renumbered)
	}

	return text
}

func renumberFields(text string) string {
	var (
		loc      = regexp.MustCompile(`(?sm){$.*^}`).FindStringIndex(text)
		body     = text[loc[0]:loc[1]]
		i        = 1
		replaced = regexp.MustCompile(`(?sm)\d+;$`).
				ReplaceAllStringFunc(body, func(s string) string {
				s = fmt.Sprintf("%d;", i)
				i++
				return s
			})
	)
	return strings.ReplaceAll(text, body, replaced)
}
