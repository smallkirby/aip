package cmd

import (
	"fmt"
	"net/url"

	"github.com/smallkirby/aip/pkg/checker"
	"github.com/smallkirby/aip/pkg/conf"
	"github.com/smallkirby/aip/pkg/mail"
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

func Mail(alltargets []string, dangers []string) error {
	toaddr, err := conf.GetMail()
	if err != nil {
		return err
	}
	var subject, body string
	if len(dangers) != 0 {
		subject = fmt.Sprintf("%d pages might be in PUBLIC!", len(dangers))
		body = "Some of your pages might be published. Check immediately.\n\n"
		body += "Below pages seems in public:\n"
		for _, target := range dangers {
			body += "\t" + target + "\n"
		}
		body += "\n\nBelow pages are in private:\n"
		for _, target := range alltargets {
			var dup = false
			for _, danger := range dangers {
				if danger == target {
					dup = true
					break
				}
			}
			if dup {
				continue
			}
			body += "\t" + target + "\n"
		}
	} else {
		subject = "all your pages are in safe."
		body = "Confirmed that below pages are in private:\n"
		for _, target := range alltargets {
			body += "\t" + target + "\n"
		}
	}
	body += "\n\nfrom AIP: Am I Public...?\n"

	if err := mail.SendMail(toaddr, subject, body); err != nil {
		return err
	}
	return nil
}
