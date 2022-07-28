package webhooks

import "time"

type EventPayload struct {
	Ref        string `json:"ref"`
	RefType    string `json:"ref_type"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository `json:"repository"`
	Pusher     `json:"pusher"`
	Commits    `json:"commits"`
	HeadCommit `json:"head_commit"`
}

type Repository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    `json:"owner"`
}

type Owner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Login string `json:"login"`
	ID    int    `json:"id"`
}

type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Commits []struct {
	Idef
	Author    `json:"author"`
	Committer `json:"committer"`
	Action
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Committer struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type HeadCommit struct {
	Idef
	Author    `json:"author"`
	Committer `json:"committer"`
	Action
}

type Idef struct {
	ID        string    `json:"id"`
	TreeID    string    `json:"tree_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Action struct {
	Added    []string      `json:"added"`
	Removed  []interface{} `json:"removed"`
	Modified []string      `json:"modified"`
}
