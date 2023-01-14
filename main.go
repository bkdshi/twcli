package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/oauth2"
)

func getToken(ctx context.Context, conf *oauth2.Config) *oauth2.Token {
	challenge := "gwSi9PRAQ3uEKQPKyJip9LCTLTXW5eRADsFb8FztJCsEKN7K9"
	codeChallenge := oauth2.SetAuthURLParam("code_challenge", challenge)
	codeChallengeMethod := oauth2.SetAuthURLParam("code_challenge_method", "plain")
	url := conf.AuthCodeURL("state", codeChallenge, codeChallengeMethod)
	// fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	err := exec.Command("xdg-open", url).Start()

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

func authorize() {}

func main() {
	search_id := flag.String("s", "", "search id")
	flag.Parse()

	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "NHZXR280YjBNWFhhUEhBczVYX2o6MTpjaQ",
		ClientSecret: "3wMtYEN34sQsHyqWg8-Je6UgQ50KPC6rl_-4MhnCtS4Z9N0bS4",
		Scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
		RedirectURL: "https://www.bkds-hi.com/callback",
	}

	raw, err := ioutil.ReadFile("./token.json")
	if err != nil {
		fmt.Println(err.Error())
		token := getToken(ctx, conf)
		bytes, _ := json.MarshalIndent(token, "", "	")
		fmt.Println(string(bytes))
		_ = ioutil.WriteFile("token.json", bytes, 0644)
		os.Exit(1)
	}

	var token *oauth2.Token

	json.Unmarshal(raw, &token)
	fmt.Println(*token)

	tmp := conf.TokenSource(ctx, token)
	new_token, err := tmp.Token()
	if err != nil {
		fmt.Println(err)
	}

	if token == new_token {
		fmt.Println("TOKEN IS SAME")
	} else {
		fmt.Println("TOKEN IS RENEWED")
		bytes, _ := json.MarshalIndent(new_token, "", "	")
		fmt.Println(string(bytes))
		_ = ioutil.WriteFile("token.json", bytes, 0644)
	}

	fmt.Println(*new_token)

	client := conf.Client(ctx, token)

	if len(*search_id) > 0 {
		fmt.Println(*search_id)
		url := fmt.Sprintf("https://api.twitter.com/2/tweets?ids=%v", *search_id)
		res, err := client.Get(url)

		if err != nil {
			fmt.Println(res)
			log.Fatal(err)
		}
		defer res.Body.Close()
		byteArray, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(byteArray)) // htmlをstringで取得
	} else {
		text := strings.Join(flag.Args(), " ")
		json := fmt.Sprintf(`{"text": "%v"}`, text)
		fmt.Println(json)
		res, err := client.Post("https://api.twitter.com/2/tweets", "application/json", bytes.NewBuffer([]byte(json)))
		defer res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		byteArray, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(byteArray)) // htmlをstringで取得
	}

}
