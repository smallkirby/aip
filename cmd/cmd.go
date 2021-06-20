package cmd

import (
	"net/url"

	"github.com/smallkirby/aip/pkg/checker"
	"github.com/smallkirby/aip/pkg/conf"
)

func Check(target string) (bool, error) {
	if _, err := url.ParseRequestURI(target); err != nil {
		return false, err
	}
	return checker.CheckPublic(target)
}

func CheckAll(targets []string) (result []bool, reserr error) {
	for _, target := range targets {
		res, err := Check(target)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return
}

func ReadConf() (targets []string, e error) {
	return conf.ReadConf()
}

func AddConf(target string) error {
	return conf.AddConf([]string{target})
}
