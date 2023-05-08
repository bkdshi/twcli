package twcore

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/oauth2"
)

func makeChallenge() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 50)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	var challenge string
	for _, v := range b {
		challenge += string(letters[int(v)%len(letters)])
	}

	return challenge
}

func getFirstToken(ctx context.Context, conf *oauth2.Config) *oauth2.Token {
	challenge := makeChallenge()
	codeChallenge := oauth2.SetAuthURLParam("code_challenge", challenge)
	codeChallengeMethod := oauth2.SetAuthURLParam("code_challenge_method", "plain")
	url := conf.AuthCodeURL("state", codeChallenge, codeChallengeMethod)
	// fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	err := exec.Command("xdg-open", url).Start()

	fmt.Println("Please in put code which located on your web browser.")
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	codeVerifier := oauth2.SetAuthURLParam("code_verifier", challenge)
	token, err := conf.Exchange(ctx, code, codeVerifier)
	fmt.Println("Exchange")
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func getToken(ctx context.Context, conf *oauth2.Config) *oauth2.Token {
	dir := os.Getenv("HOME")
	dir = filepath.Join(dir, ".config", "twcli")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil
	}

	file := filepath.Join(dir, "token.json")
	raw, err := ioutil.ReadFile(file)

	// if there is not "token.json", create a new token and make file.
	if err != nil {
		fmt.Println(err.Error())
		token := getFirstToken(ctx, conf)
		bytes, _ := json.MarshalIndent(token, "", "	")
		// fmt.Println(string(bytes))
		_ = ioutil.WriteFile(file, bytes, 0644)
		return token
	}
	// load token from "token.json".
	var token *oauth2.Token

	json.Unmarshal(raw, &token)
	// fmt.Println(*token)

	// renew token process.
	freshTokenSorce := conf.TokenSource(ctx, token)
	new_token, err := freshTokenSorce.Token()
	if err != nil {
		fmt.Println(err)
	}

	if token == new_token {
		return token
		// fmt.Println("TOKEN IS SAME")
	} else {
		// fmt.Println("TOKEN IS RENEWED")
		// fmt.Println(*new_token)
		bytes, _ := json.MarshalIndent(new_token, "", "	")
		// fmt.Println(string(bytes))
		_ = ioutil.WriteFile(file, bytes, 0644)
		return new_token
	}

	// return token
}

func getConfig() *oauth2.Config {
	conf := &oauth2.Config{
		ClientID:     "TWJXNDk0dlBsa25ILS1vcXZSMm06MTpjaQ",
		ClientSecret: "rzNOyRZZ2luULq7WRxg9WyjlCMcJ4QJBZQf8D8g9mKOSaGYi2_",
		Scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
		RedirectURL: "https://www.bkds-hi.com/callback",
	}
	return conf
}
