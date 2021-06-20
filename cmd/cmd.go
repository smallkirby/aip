package cmd

import (
	"net/url"

	"github.com/smallkirby/aip/pkg/checker"
	"github.com/smallkirby/aip/pkg/conf"
)

type BoolResult struct {
	Result bool
	URL    string
	Error  error
}

func Check(target string) (bool, error) {
	if _, err := url.ParseRequestURI(target); err != nil {
		return false, err
	}
	return checker.CheckPublic(target)
}

func CheckAll(targets []string, ch chan BoolResult) {
	for _, target := range targets {
		res, err := Check(target)
		if err != nil {
			ch <- BoolResult{false, "", err}
		}
		ch <- BoolResult{res, target, nil}
	}
	close(ch)
}

func ReadConf() (targets []string, e error) {
	return conf.ReadConf()
}

func AddConf(target string) error {
	return conf.AddConf([]string{target})
}
