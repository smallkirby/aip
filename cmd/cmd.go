package cmd

import (
	"net/url"

	"github.com/smallkirby/aip/pkg/checker"
)

func Check(target string) (bool, error) {
	if _, err := url.ParseRequestURI(target); err != nil {
		return false, err
	}
	return checker.CheckPublic(target)
}
