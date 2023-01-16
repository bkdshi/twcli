package twcli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

type Tweet struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

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
		ClientID:     "NHZXR280YjBNWFhhUEhBczVYX2o6MTpjaQ",
		ClientSecret: "3wMtYEN34sQsHyqWg8-Je6UgQ50KPC6rl_-4MhnCtS4Z9N0bS4",
		Scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
		RedirectURL: "https://www.bkds-hi.com/callback",
	}
	return conf
}

type App struct {
	client *http.Client
}

func (app *App) authorization() {
	ctx := context.Background()
	conf := getConfig()
	token := getToken(ctx, conf)
	app.client = conf.Client(ctx, token)
}

func (app *App) search(id string) {
	url := fmt.Sprintf("https://api.twitter.com/2/tweets?ids=%v", id)
	res, err := app.client.Get(url)

	if err != nil {
		fmt.Println(res)
		log.Fatal(err)
	}
	defer res.Body.Close()
	byteArray, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(byteArray)) // htmlをstringで取得
}

func (app *App) tweet(text string) {
	json := fmt.Sprintf(`{"text": "%v"}`, text)
	res, err := app.client.Post("https://api.twitter.com/2/tweets", "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	byteArray, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(byteArray)) // htmlをstringで取得
}

func (app *App) list(username string) {
	if username == "me" {
		user := app.getMe()
		username = user.Id
	}
	query := fmt.Sprintf("from:%v", username)
	url := fmt.Sprintf("https://api.twitter.com/2/tweets/search/recent?query=%v", query)
	res, err := app.client.Get(url)

	if err != nil {
		fmt.Println(res)
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	type Response struct {
		Data []Tweet `json:"data"`
	}

	var response Response
	json.Unmarshal(body, &response)

	response_json, err := json.MarshalIndent(response.Data, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(response_json))
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func (app *App) getMe() User {
	res, err := app.client.Get("https://api.twitter.com/2/users/me")

	if err != nil {
		fmt.Println(res)
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	type Response struct {
		Data User `json:"data"`
	}

	var response Response
	json.Unmarshal(body, &response)

	return response.Data
}

func (app *App) showMe() {
	User := app.getMe()

	user_json, err := json.MarshalIndent(User, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(user_json))
}

func main() {
	var app App

	show_me := flag.Bool("u", false, "show your info")
	search_id := flag.String("s", "", "search id")
	show_tweet := flag.Bool("l", false, "list tweet from user name")
	flag.Parse()

	app.authorization()

	if *show_me {
		app.showMe()
	} else if len(*search_id) > 0 {
		app.search(*search_id)
	} else if *show_tweet {
		if len(flag.Args()) == 0 {
			app.list("me")
		} else {
			app.list(flag.Args()[0])
		}
	} else {
		app.tweet(strings.Join(flag.Args(), " "))
	}

}
