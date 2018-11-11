package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var version = "development"

func init() {
	mp := os.Getenv("DISCUSS_MAX_PAGES")
	if mp == "" {
		return
	}
	mpi, err := strconv.Atoi(mp)
	if err != nil {
		return
	}
	maxPages = mpi
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	org := os.Args[1]
	if strings.HasPrefix(org, "-") {
		usage()
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		usage()
	}

	ctx := context.Background()
	lastMonth := time.Now().AddDate(0, -1, 0)

	graphqlURL := "https://api.github.com/graphql"
	var loadingChan = make(chan struct{})
	discFunc := func() ([]discussion, error) {
		return getTeamDiscussions(ctx, org, token, graphqlURL, lastMonth, loadingChan)
	}
	err := runUI(loadingChan, discFunc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "got an error:\n%s", err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `discuss version: %s

Usage:
GITHUB_TOKEN=<token> discuss <org>

org     The org on GitHub to view discussions on
token   A GitHub Personal Access Token
`, version)
	os.Exit(2)
}
