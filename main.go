package main

import (
	"flag"
	"strings"

	"github.com/bkdshi/twcli/twcore"
)

func main() {
	var app twcore.App

	show_me := flag.Bool("u", false, "show your info")
	search_id := flag.String("s", "", "search id")
	show_tweet := flag.Bool("l", false, "list tweet from user name")
	flag.Parse()

	app.Authorization()

	if *show_me {
		app.ShowMe()
	} else if len(*search_id) > 0 {
		app.Search(*search_id)
	} else if *show_tweet {
		if len(flag.Args()) == 0 {
			app.List("me")
		} else {
			app.List(flag.Args()[0])
		}
	} else {
		app.Tweet(strings.Join(flag.Args(), " "))
	}

}
