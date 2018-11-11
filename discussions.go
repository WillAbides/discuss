package main

import (
	"bytes"
	"context"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type teams struct {
	PageInfo struct {
		EndCursor   graphql.String
		HasNextPage bool
	}
	Nodes []struct {
		team `graphql:"..."`
	}
}

type team struct {
	Name        string
	ID          string
	Discussions struct {
		Nodes []discussion
	} `graphql:"discussions(first: 31, orderBy: {field: CREATED_AT, direction: DESC})"`
}

type discussion struct {
	Title     string
	CreatedAt time.Time
	URL       string
	Team      struct {
		Name string
	}
	Author struct {
		Login string
	}
}

type rt struct {
	realTransport http.RoundTripper
	lchan         chan struct{}
}

var maxPages int

func (t *rt) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lchan <- struct{}{}
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	req.Header.Add("Accept", "application/vnd.github.echo-preview+json")
	return t.realTransport.RoundTrip(req)
}

var query struct {
	Organization struct {
		Teams teams `graphql:"teams(first: 100, orderBy: {field: NAME, direction: ASC}, after: $cursor)"`
	} `graphql:"orgs: organization(login: $login)"`
}

func getTeamDiscussions(ctx context.Context, org string, targetTime time.Time, loading chan struct{}, token, graphqlURL string) ([]discussion, error) {
	var discussions []discussion
	variables := map[string]interface{}{
		"login":  graphql.String(org),
		"cursor": (*graphql.String)(nil),
	}

	allTeams := map[string]team{}

	client := newGraphqlClient(token, loading, graphqlURL)

	pageCount := 0
	for {
		err := client.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}
		for _, tt := range query.Organization.Teams.Nodes {
			allTeams[tt.team.ID] = tt.team
		}
		if !query.Organization.Teams.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = query.Organization.Teams.PageInfo.EndCursor
		pageCount++
		if maxPages > 0 && pageCount >= maxPages {
			break
		}
	}
	close(loading)

	for _, team := range allTeams {
		for _, discussion := range team.Discussions.Nodes {
			if discussion.CreatedAt.After(targetTime) {
				discussions = append(discussions, discussion)
			}
		}
	}
	sort.Slice(discussions, func(i, j int) bool {
		return discussions[i].CreatedAt.After(discussions[j].CreatedAt)
	})
	return discussions, nil
}

func newGraphqlClient(token string, loading chan struct{}, graphqlURL string) *graphql.Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tripper := oauth2.NewClient(context.Background(), src).Transport
	r := rt{
		realTransport: tripper,
		lchan:         loading,
	}
	httpClient := &http.Client{
		Transport: &r,
	}
	client := graphql.NewClient(graphqlURL, httpClient)
	return client
}
