package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	org := os.Args[1]
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		usage()
	}

	ctx := context.Background()
	lastMonth := time.Now().AddDate(0, -1, 0)

	graphqlURL := "https://api.github.com/graphql"
	var loadingChan = make(chan struct{})
	discFunc := func() ([]discussion, error) {
		return getTeamDiscussions(ctx, org, lastMonth, loadingChan, token, graphqlURL)
	}
	err := runUI(loadingChan, discFunc)
	fmt.Println("foo")
	if err != nil {
		fmt.Fprintf(os.Stderr, "got an error:\n%s", err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `Usage:
GITHUB_TOKEN=<token> discuss <org>

org     The org on GitHub to view discussions on
token   A GitHub Personal Access Token
`)
	os.Exit(2)
}
