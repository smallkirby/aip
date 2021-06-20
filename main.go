package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/pkg/term/termios"
	"github.com/smallkirby/aip/cmd"
	"github.com/smallkirby/aip/pkg/conf"
	"golang.org/x/sys/unix"
)

var original_tty_attr *unix.Termios
var target_urls []string

func main() {
	// go-prompt changes current tty attributes, and doesn't restore it...
	if tty_attr, err := termios.Tcgetattr(0); err != nil {
		log.Fatal(err)
	} else {
		original_tty_attr = tty_attr
	}
	defer restoreTty()

	// run it
	p := prompt.New(
		executer,
		completer,
		prompt.OptionPrefix("(command)> "),
	)
	p.Run()
}

func restoreTty() {
	termios.Tcsetattr(0, termios.TCSANOW, original_tty_attr)
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "check", Description: "Check all targets in config file."},
		{Text: "check <URL>", Description: "Check if the page is public"},
		{Text: "conf read", Description: "Read configuration file."},
		{Text: "conf add <target>", Description: "Add target URL in configuration."},
		{Text: "exit", Description: "I miss you..."},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executer(com string) {
	if strings.HasPrefix(com, "check") {
		s := strings.Split(com, " ")
		if len(s) == 1 {
			if len(target_urls) >= 1 {
				res, err := cmd.CheckAll(target_urls)
				if err != nil {
					println(err.Error())
					return
				} else {
					for ix, r := range res {
						if r {
							danger := color.New(color.FgRed, color.Bold).SprintFunc()
							fmt.Printf("%v: %v\n", danger("PUBLIC "), target_urls[ix])
						} else {
							fmt.Printf("%v: %v\n", color.GreenString("private"), target_urls[ix])
						}
					}
				}
			} else {
				println("Read config before check.")
				return
			}
			return
		} else if len(s) != 2 {
			println("should specify one URL")
			return
		}
		target := s[1]
		if result, err := cmd.Check(target); err != nil {
			log.Println(err.Error())
			restoreTty()
			os.Exit(0)
		} else {
			if result {
				danger := color.New(color.FgGreen, color.Bold).SprintFunc()
				fmt.Printf("%v: %v\n", danger("PUBLIC "), target)
			} else {
				fmt.Printf("%v: %v\n", color.GreenString("private"), target)
			}
		}
	} else if strings.HasPrefix(com, "conf") {
		s := strings.Split(com, " ")
		if len(s) == 1 {
			fmt.Printf("Invalid command: %v\n", com)
			return
		}
		if s[1] == "read" {
			if targets, err := conf.ReadConf(); err != nil {
				log.Fatalln(err.Error())
			} else {
				target_urls = targets
				fmt.Printf("Set %v URL as targets.\n", len(target_urls))
			}
		} else if s[1] == "add" {
			if len(s) != 3 {
				fmt.Printf("Invalid command: %v\n", com)
				return
			}
			if err := cmd.AddConf(s[2]); err != nil {
				println(err.Error())
				return
			}
		}
	} else if com == "exit" {
		println("I miss you...")
		restoreTty()
		os.Exit(0)
	} else {
		println("not imp")
	}
}
