package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GetClient(config *oauth2.Config) (*http.Client, error) {
	tok, err := tokenFromFile()
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tok); err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\ninput > ", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.New(fmt.Sprintf("\nUnable to read authorization code: %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("\nUnable to retrieve token from web: %v", err))
	}
	return tok, nil
}

func saveToken(token *oauth2.Token) error {
	confhome, _ := os.UserHomeDir()
	path := filepath.Join(confhome, ".aip/gtoken.json")
	fmt.Printf("\n\nSaving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)
	return nil
}

func tokenFromFile() (*oauth2.Token, error) {
	homedir, _ := os.UserHomeDir()
	confdir := filepath.Join(homedir, ".aip")
	// check if the dir/file exists
	if _, err := os.Stat(confdir); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Failed to open config directory at %v.\n%v", confdir, err.Error()))
	}
	conffile := filepath.Join(confdir, "gtoken.json")
	if stat, err := os.Stat(conffile); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("Failed to open config file at %v.\n%v\n", conffile, err.Error()))
	} else {
		if stat.Mode() != 0600 {
			return nil, errors.New(fmt.Sprintf("Invalid permission for config file(%v): %v", conffile, stat.Mode()))
		}
	}

	f, err := os.Open(conffile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func SendMail(toaddr string, subject string, body string) error {
	ctx := context.Background()

	homedir, _ := os.UserHomeDir()
	confdir := filepath.Join(homedir, ".aip")
	// check if the dir/file exists
	if _, err := os.Stat(confdir); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Failed to open config directory at %v.\n%v", confdir, err.Error()))
	}
	conffile := filepath.Join(confdir, "gcred.json")
	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Failed to open config file at %v.\n%v\n", conffile, err.Error()))
	}

	b, err := ioutil.ReadFile(conffile)
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return err
	}
	client, err := GetClient(config)
	if err != nil {
		return err
	}
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	var message gmail.Message
	msg := bytes.NewBuffer([]byte(""))
	msg.WriteString("From: 'me'\r\n")
	msg.WriteString(fmt.Sprintf("To: %v\r\n", toaddr))
	msg.WriteString(fmt.Sprintf("Subject: AIP: %v\r\n", subject))
	msg.WriteString("\r\n")
	msg.WriteString(body)

	message.Raw = base64.StdEncoding.EncodeToString(msg.Bytes())
	message.Raw = strings.Replace(message.Raw, "/", "_", -1)
	message.Raw = strings.Replace(message.Raw, "+", "-", -1)
	message.Raw = strings.Replace(message.Raw, "=", "", -1)

	_, err = srv.Users.Messages.Send("me", &message).Do()
	if err != nil {
		return err
	}
	return nil
}
