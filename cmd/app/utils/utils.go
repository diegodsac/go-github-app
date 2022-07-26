package utils

import (
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/diegodsac/go-github-app/cmd/app/config"
	"github.com/google/go-github/v41/github"
)

func InitGitHubClient() {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, 222025, 27686641, "config/github-app.pem")

	if err != nil {
		log.Fatal(err)
	}

	config.Config.GitHubClient = github.NewClient(&http.Client{Transport: itr})
}
