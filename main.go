package main

import (
	"flag"
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

type FlagContext struct {
	help           *bool
	quiet          *bool
	noninteractive *bool
	version        *bool
	mail           *bool
	conffile       *string
}

const VERSION = "0.1.0"

var original_tty_attr *unix.Termios
var target_urls []string
var context FlagContext

func main() {
	parseArgs()
	if *context.help {
		usage()
		os.Exit(0)
	}
	if *context.version {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	if *context.noninteractive {
		doNonInteractive()
	} else {
		if *context.quiet {
			fmt.Fprintln(os.Stderr, "--quiet options is available only with -n option.")
			os.Exit(1)
		} else {
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
	}
}

func doNonInteractive() {
	safe_num := 0
	danger_num := 0
	error_num := 0
	danger_urls := []string{}

	quiet := *context.quiet
	target_urls, err := cmd.ReadConf()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	ch := make(chan cmd.BoolResult, len(target_urls))
	go cmd.CheckAll(target_urls, ch)
	for result := range ch {
		if result.Error != nil {
			fmt.Fprintln(os.Stderr, result.Error.Error()) // not return
			error_num += 1
		} else {
			if result.Result {
				danger := color.New(color.FgRed, color.Bold).SprintFunc()
				if !quiet {
					fmt.Printf("%v: %v\n", danger("PUBLIC "), result.URL)
				}
				danger_urls = append(danger_urls, result.URL)
				danger_num += 1
			} else {
				if !quiet {
					fmt.Printf("%v: %v\n", color.GreenString("private"), result.URL)
				}
				safe_num += 1
			}
		}
	}
	if quiet {
		for _, url := range danger_urls {
			fmt.Println(url)
		}
	} else {
		fmt.Printf("Result: %v public, %v private, %v errors\n", color.RedString(fmt.Sprintf("%v", danger_num)), color.GreenString(fmt.Sprintf("%v", safe_num)), color.BlueString(fmt.Sprintf("%v", error_num)))
	}
	if *context.mail {
		if err := cmd.Mail(target_urls, danger_urls); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
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
				safe_num := 0
				danger_num := 0
				error_num := 0

				ch := make(chan cmd.BoolResult, len(target_urls))
				go cmd.CheckAll(target_urls, ch)
				for result := range ch {
					if result.Error != nil {
						fmt.Fprintln(os.Stderr, result.Error.Error()) // not return
						error_num += 1
					} else {
						if result.Result {
							danger := color.New(color.FgRed, color.Bold).SprintFunc()
							fmt.Printf("%v: %v\n", danger("PUBLIC "), result.URL)
							danger_num += 1
						} else {
							fmt.Printf("%v: %v\n", color.GreenString("private"), result.URL)
							safe_num += 1
						}
					}
				}
				fmt.Printf("Result: %v public, %v private, %v errors\n", color.RedString(fmt.Sprintf("%v", danger_num)), color.GreenString(fmt.Sprintf("%v", safe_num)), color.BlueString(fmt.Sprintf("%v", error_num)))
			} else {
				fmt.Fprintln(os.Stderr, "Read config before check.")
				return
			}
			return
		} else if len(s) != 2 {
			fmt.Fprintln(os.Stderr, "should specify one URL")
			return
		}
		target := s[1]
		if result, err := cmd.Check(target); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
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
				fmt.Fprintln(os.Stderr, err.Error())
				return
			}
		}
	} else if com == "exit" {
		fmt.Fprintln(os.Stderr, "I miss you...")
		restoreTty()
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stderr, "not imp")
	}
}

func parseArgs() {
	context.help = flag.Bool("help", false, "show help.")
	context.quiet = flag.Bool("quiet", false, "silent mode. available only in non-interactive mode.")
	context.version = flag.Bool("version", false, "show version information.")
	context.mail = flag.Bool("mail", false, "send the result to specified e-mail.")
	context.noninteractive = flag.Bool("n", false, "non-interactive mode.")
	context.conffile = flag.String("f", "", "specify config file.")
	flag.Parse()
}

func usage() {
	title := color.New(color.FgBlue, color.Bold).SprintFunc()
	fmt.Println(title("AIP: Am I Public..."))
	fmt.Printf("version: %v\tnirugiri. 2021.\n", VERSION)
	flag.Usage()
}
