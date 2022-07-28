package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/diegodsac/go-github-app/cmd/app/config"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v45/github"
	githubapps "github.com/palantir/go-githubapp/githubapp"
)

type EP struct {
	E github.IssueCommentEvent
	P EventPayload
	c *gin.Context
}

type PRCommentHandler struct {
	githubapps.ClientCreator

	preamble string
}

type Event string

const (
	Install     Event = "installation"
	Ping        Event = "ping"
	Push        Event = "push"
	PullRequest Event = "pull_request"
	Delete      Event = "delete"
	Create      Event = "create"
)

var Events = []Event{
	Install,
	Ping,
	Push,
	PullRequest,
	Delete,
	Create,
}

var Consumers = map[string]func(EP) error{
	string(Install):     consumeInstallEvent,
	string(Ping):        consumePingEvent,
	string(Push):        consumePushEvent,
	string(PullRequest): consumePullRequestEvent,
	string(Delete):      consumeDeleteEvent,
	string(Create):      consumeCreateEvent,
}

func VerifySignature(payload []byte, signature string) bool {
	key := hmac.New(sha256.New, []byte(config.Config.GitHubWebhookSecret))
	key.Write([]byte(string(payload)))
	computedSignature := "sha256=" + hex.EncodeToString(key.Sum(nil))
	log.Printf("computed signature: %s", computedSignature)

	return computedSignature == signature
}

func ConsumeEvent(c *gin.Context) {
	var ev EP
	ev.c = c
	payload, _ := ioutil.ReadAll(c.Request.Body)

	if !VerifySignature(payload, c.GetHeader("X-Hub-Signature-256")) {
		c.AbortWithStatus(http.StatusUnauthorized)
		log.Println("signatures don't match")
	}

	event := c.GetHeader("X-GitHub-Event")

	json.Unmarshal(payload, &ev.E)

	for _, e := range Events {
		if string(e) == event {
			log.Printf("consuming event: %s", e)
			// var p EventPayload
			json.Unmarshal(payload, &ev.P)
			if err := Consumers[string(e)](ev); err != nil {
				log.Printf("couldn't consume event %s, error: %+v", string(e), err)
				// We're responding to GitHub API, we really just want to say "OK" or "not OK"
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"reason": err})
			}
			log.Printf("consumed event: %s", e)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
	log.Printf("Unsupported event: %s", event)
	c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{"reason": "Unsupported event: " + event})
}

func consumeInstallEvent(payload EP) error {
	// Process event ...
	// Insert data into database ...
	return nil
}

func consumePingEvent(payload EP) error {
	// Process event ...
	// Insert data into database ...
	return nil
}

func consumePushEvent(payload EP) error {

	// Process event ...
	// Insert data into database ...

	log.Printf("Received push from %s, by user %s, on branch %s",
		payload.P.Repository.FullName,
		payload.P.Pusher.Name,
		payload.P.Ref)

	// Enumerating commits
	var commits []string
	for _, commit := range payload.P.Commits {
		commits = append(commits, commit.ID)
	}
	log.Printf("Pushed commits: %v", commits)

	return nil
}

func consumePullRequestEvent(payload EP) error {
	// Process event ...
	// Insert data into database ...
	return nil
}

func consumeDeleteEvent(payload EP) error {
	log.Printf("Received Delete from %s, by user %s, on branch %s",
		payload.P.Repository.FullName,
		payload.P.Pusher.Name,
		payload.P.Ref)
	return nil
}

func consumeCreateEvent(payload EP) error {
	// var ctx context.Context
	// var h githubapps.ClientCreator
	// var branch *string
	// var ttt *github.Reference
	var baseRef *github.Reference
	var ctx = context.Background()
	// var err error
	var baseBranch, commitBranch, headBranch string
	commitBranch = "PREV"
	headBranch = "main"
	commitMessage := "commit"
	log.Printf("Received Branch from %s, by user %s, on branch %s",
		payload.P.Repository.FullName,
		payload.P.RefType,
		payload.P.Ref)

	sourceOwner := payload.E.GetRepo().GetOwner().GetLogin()
	sourceRepo := payload.E.GetRepo().GetName()
	baseBranch = payload.P.Ref
	baseRef, _, err := config.Config.GitHubClient.Git.GetRef(ctx, sourceOwner, sourceRepo, "refs/heads/"+baseBranch)
	if err != nil {
		return err
	}

	log.Printf("\n########## BASE REF:##########\n %s", baseRef)
	ref, _, err := config.Config.GitHubClient.Git.GetRef(ctx, sourceOwner, sourceRepo, "refs/heads/"+commitBranch)

	// log.Printf("########## REF TESTE:##########\n %s", ref)
	// log.Printf("########## REF ERR:##########\n %s", err)
	if ref == nil {
		newRef := &github.Reference{
			Ref:    github.String("refs/heads/" + commitBranch),
			Object: &github.GitObject{SHA: baseRef.Object.SHA}}

		ref, _, err = config.Config.GitHubClient.Git.CreateRef(ctx, sourceOwner, sourceRepo, newRef)
		if err != nil {
			return err
		}
	}
	log.Printf("\n########## CREATE REF ##########\n %s", ref)
	mergeReq := &github.RepositoryMergeRequest{
		Base:          &baseBranch,
		Head:          &headBranch,
		CommitMessage: &commitMessage,
	}

	config.Config.GitHubClient.Repositories.Merge(ctx, sourceOwner, sourceRepo, mergeReq)
	log.Printf("\n########## MERGE ##########\n")
	// installationID := githubapps.GetInstallationIDFromEvent(&payload.E)
	// ctx, _ := githubapps.PrepareRepoContext(payload.c, installationID, payload.E.GetRepo())
	// log.Printf("Received Branch from %s", config.Config.GitHubClient)
	// if _, _, err := config.Config.GitHubClient.Git.CreateRef(ctx, payload.E.GetRepo().GetOwner().GetLogin(), payload.E.GetRepo().GetName(), &github.Reference{Ref: github.String("refs/heads/Dev"),Object: &github.GitObject{SHA: }}); err != nil {
	// 	log.Printf("Failed to comment on pull request")
	// }
	// repo := payload.E.GetRepo().String()
	// owner := payload.E.GetRepo().GetOwner().GetLogin()
	// log.Printf("C: %s", payload.c)
	// owner := payload.c.Param("owner")
	// repo := payload.c.Param("repo")
	// // *branch = "refs/heads/dev"
	// // ttt.Ref = branch
	// _, _, err := config.Config.GitHubClient.Git.CreateRef(payload.c, owner, repo, &github.Reference{Ref: github.String("refs/heads/Dev")})
	// if err != nil {
	// 	log.Printf("ERROR: %s", err)
	// }
	// // prNum := payload.E.GetIssue().GetNumber()
	// installationID := githubapps.GetInstallationIDFromEvent(&payload.E)
	// installationID := int64(27686641)

	// // ctx, logger := githubapp.PreparePRContext(ctx, installationID, repo, payload.E.GetIssue().GetNumber())
	// log.Printf("InstallationID %d", installationID)
	// _, err := githubapps.
	// // _, err := h.NewInstallationClient(installationID)
	// if err != nil {
	// 	return err
	// }
	// repoOwner := payload.E.GetRepo().GetOwner().GetLogin()
	// repoName := payload.E.GetRepo().GetName()
	// var branch *string
	// *branch = "refs/heads/dev"
	// client.Git.CreateRef(ctx, repoOwner, repoName, &github.Reference{Ref: branch})

	// client, err := h.NewInstallationClient(installationID)
	// if err != nil {
	// 	return err
	// }
	// repoOwner := payload.E.GetRepo().GetOwner().GetLogin()
	// repoName := payload.E.GetRepo().GetName()
	// var branch *string
	// *branch = "refs/heads/dev"
	// client.Git.CreateRef(ctx, repoOwner, repoName, &github.Reference{Ref: branch})
	// if _, _, err := client.Issues.CreateComment(ctx, repoOwner, repoName, prNum, &prComment); err != nil {
	// 	log.Printf("Failed to comment on pull request")
	// }

	return nil
}
