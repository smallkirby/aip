package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func ReadConf() (targets []string, e error) {
	homedir, _ := os.UserHomeDir()
	confdir := filepath.Join(homedir, ".aip")
	// check if the dir/file exists
	if _, err := os.Stat(confdir); os.IsNotExist(err) {
		if err := os.Mkdir(confdir, 0775); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to create config directory at %v.\n%v", confdir, err.Error()))
		}
		fmt.Printf("[+] Created config directory at %v.\n", color.BlueString(confdir))
	}
	conffile := filepath.Join(confdir, "aip.conf")
	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		if _, err := os.Create(conffile); err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to create config file at %v.\n%v\n", conffile, err.Error()))
		}
		fmt.Printf("[+] Created config file at %v.\n", color.BlueString(conffile))
	}

	// read config
	fbytes, err := os.ReadFile(conffile)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to read config file at %v.\n%v\n", conffile, err.Error()))
	}
	lines := strings.Split(string(fbytes), "\n")
	for _, line := range lines {
		if len(line) < 1 {
			continue
		} else {
			targets = append(targets, line)
		}
	}
	return
}
