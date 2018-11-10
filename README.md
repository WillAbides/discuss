# discuss

Discuss is a command line tool that lists discussions for all teams in
a GitHub org for the past month (up to 30 discussions per team). It
displayes them as a selectable list. Navigate to one you are interested
in an press enter to load it in a browser.

```
Usage:
GITHUB_TOKEN=<token> discuss <org>

org     The org on GitHub to view discussions on
token   A GitHub Personal Access Token
```

## Install

### From source

If you have go 1.11+ installed on your system, run `go get github.com/WillAbides/discuss`

### From binary

Download a binary for your system from the [latest GitHub release](https://github.com/WillAbides/discuss/releases)
