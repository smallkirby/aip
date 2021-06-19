package checker

import (
	"io/ioutil"
	"net/http"
)

func fetchPage(url string) (int, string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(url)
	if err != nil {
		return -1, "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, "", err
	}
	bodystr := string(body)
	return res.StatusCode, bodystr, nil
}

// XXX not imp, just a mock
func CheckPublic(url string) (bool, error) {
	code, _, err := fetchPage(url)
	println(code) // XXX
	if err != nil {
		return true, err
	}
	if code/100 == 4 {
		return false, nil
	}

	return true, nil
}
