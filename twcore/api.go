package twcore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type App struct {
	client *http.Client
}

type Tweet struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

func (app *App) Authorization() {
	ctx := context.Background()
	conf := getConfig()
	token := getToken(ctx, conf)
	app.client = conf.Client(ctx, token)
}

func (app *App) Search(id string) {
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

func (app *App) Tweet(text string) string {
	json := fmt.Sprintf(`{"text": "%v"}`, text)
	res, err := app.client.Post("https://api.twitter.com/2/tweets", "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	byteArray, _ := ioutil.ReadAll(res.Body)
	return string(byteArray)
}

func (app *App) GetList(username string) []Tweet {
	if username == "me" {
		user := app.GetMe()
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

	return response.Data
}

func (app *App) ShowList(username string) {
	tweets := app.GetList(username)
	for _, v := range tweets {
		fmt.Printf("id: %v\t\t%v\n", v.Id, strings.ReplaceAll(v.Text, "\n", "\t"))
	}
}

func (app *App) GetMe() User {
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

func (app *App) ShowMe() {
	User := app.GetMe()

	user_json, err := json.MarshalIndent(User, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(user_json))
}
