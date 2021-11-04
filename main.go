package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kzaag/gnuflag"
)

func printUsage() {
	fmt.Printf("%s [OPTIONS]\nOPTIONS:\n\t-s\tGenerate sport categories\n\t-l\tGenerate languages\n", os.Args[0])
}

func main() {

	gotFlag := false

	// user may provide 2 boolean flags: -sl to specify generation of
	// langs or sports or both
	gnuflag.Getopt(os.Args[1:], func(opt, optarg string) bool {
		gotFlag = true
		switch opt {
		case "s":
			if err := genSports(); err != nil {
				log.Fatal(err)
			}
		case "l":
			if err := genLangs(); err != nil {
				log.Fatal(err)
			}
		case "f":
			t, err := NewGoogleTranslateToken()
			if err != nil {
				log.Fatal(err)
			}
			t.AccessToken = strings.TrimRight(t.AccessToken, ".")
			var tres TranslateResponse
			if err := GoogleTranslate(t, &TranslateRequest{
				Q:      []string{"dzień dobry", "do widzenia"},
				Source: "pl",
				Target: "en",
				Format: "text",
			}, &tres); err != nil {
				log.Fatal(err)
			}
			s, _ := json.MarshalIndent(tres, "", "  ")
			fmt.Println(string(s))
		default:
			printUsage()
		}
		return true
	}, "s", "l", "f")

	if !gotFlag {
		printUsage()
	}
}
