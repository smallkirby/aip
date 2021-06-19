package main

import (
	"log"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/pkg/term/termios"
	"github.com/smallkirby/aip/cmd"
	"golang.org/x/sys/unix"
)

var original_tty_attr *unix.Termios

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
		{Text: "check <URL>", Description: "Check if the page is public"},
		{Text: "exit", Description: "I miss you..."},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func executer(com string) {
	if strings.HasPrefix(com, "check") {
		s := strings.Split(com, " ")
		if len(s) != 2 {
			println("should specify one URL")
			return
		}
		if result, err := cmd.Check(s[1]); err != nil {
			log.Println(err)
			restoreTty()
			os.Exit(0)
		} else {
			if result {
				println("⚠️　BE CAREFUL, IT MIGHT BE PUBLISHED. ⚠️")
			} else {
				println("Yeah, it's private!")
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
