package github

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	"github.com/mikkeloscar/maze-repo/model"
	"golang.org/x/oauth2"
)

const (
	defaultURL = "https://github.com"
	defaultAPI = "https://api.github.com"
)

var defaultScope = []string{"repo"}

// Github defines a github remote.
type Github struct {
	URL    string
	API    string
	Client string
	Secret string
}

// TODO: config
// Load loads the github remote.
func Load(client, secret string) *Github {
	github := Github{
		URL:    defaultURL,
		API:    defaultAPI,
		Client: client,
		Secret: secret,
	}

	return &github
}

// newClient returns a oauth2 authenticated github client.
func newClient(uri, token string) *github.Client {
	t := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	// tc := config.Cliento(oauth2.NoContext, &oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, t)

	c := github.NewClient(tc)
	c.BaseURL, _ = url.Parse(uri)
	return c
}

// Login authenticates the session and returns the remoter user details.
func (g *Github) Login(res http.ResponseWriter, req *http.Request) (*model.User, error) {
	var config = &oauth2.Config{
		ClientID:     g.Client,
		ClientSecret: g.Secret,
		Scopes:       defaultScope,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", g.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", g.URL),
		},
		// RedirectURL: fmt.Sprintf("%s/login/authorize", httputil.GetURL(req)),
	}

	// get the OAuth code
	var code = req.FormValue("code")

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}

	client := newClient(g.API, tok.AccessToken)
	userInfo, _, err := client.Users.Get("")
	if err != nil {
		return nil, err
	}

	user := model.User{}
	user.Login = *userInfo.Login
	user.Token = tok.AccessToken
	return &user, nil
}

// EmptyCommit creates/adds a new empty commit to a branch of a repo.
// if srcBranch and dstBranch are different then the commit will include the
// state of srcbranch effectively rebasing dstBranch onto srcBranch.
func (g *Github) EmptyCommit(u *model.User, repo, srcBranch, dstBranch, msg string) error {
	client := newClient(g.API, u.Token)
	// Get head of Srcbranch
	r, _, err := client.Git.GetRef(u.Login, repo, fmt.Sprintf("heads/%s", srcBranch))
	if err != nil {
		return err
	}

	// get the last commit (head of branch)
	c, _, err := client.Git.GetCommit(u.Login, repo, *r.Object.SHA)
	if err != nil {
		return err
	}

	// Get tree of latest commit
	t, _, err := client.Git.GetTree(u.Login, repo, *r.Object.SHA, false)
	if err != nil {
		return err
	}

	// create a new tree identical to the parent (no changes empty commit)
	t, _, err = client.Git.CreateTree(u.Login, repo, *c.Tree.SHA, t.Entries)
	if err != nil {
		return err
	}

	if srcBranch != dstBranch {
		// Get head of branch
		r, _, err = client.Git.GetRef(u.Login, repo, fmt.Sprintf("heads/%s", dstBranch))
		if err != nil {
			return err
		}

		// get the last commit (head of branch)
		c, _, err = client.Git.GetCommit(u.Login, repo, *r.Object.SHA)
		if err != nil {
			return err
		}
	}

	// create new commit based on the unchanged tree
	commit := &github.Commit{
		Message: &msg,
		Tree:    t,
		Parents: []github.Commit{*c},
	}
	c2, _, err := client.Git.CreateCommit(u.Login, repo, commit)
	if err != nil {
		return err
	}

	// point head of branch to the new commit
	ref := &github.Reference{
		Ref: r.Ref,
		Object: &github.GitObject{
			SHA: c2.SHA,
		},
	}
	_, _, err = client.Git.UpdateRef(u.Login, repo, ref, false)
	if err != nil {
		return err
	}

	return nil
}

// SetupBranch sets up a new branch based on srcBranch. If the branch already
// exists nothing happens.
func (g *Github) SetupBranch(u *model.User, repo, srcBranch, dstBranch string) error {
	client := newClient(g.API, u.Token)
	// check if dstBranch exists
	_, resp, err := client.Git.GetRef(u.Login, repo, fmt.Sprintf("heads/%s", dstBranch))
	if err != nil {
		if resp.StatusCode != 404 {
			return err
		}
	}

	// branch already exist
	if resp.StatusCode == 200 {
		return nil
	}

	// Get head of Srcbranch
	r, _, err := client.Git.GetRef(u.Login, repo, fmt.Sprintf("heads/%s", srcBranch))
	if err != nil {
		return err
	}

	// Create new branch name
	branchRef := fmt.Sprintf("refs/heads/%s", dstBranch)

	// create new branch
	ref := &github.Reference{
		Ref:    &branchRef,
		Object: r.Object,
	}
	_, _, err = client.Git.CreateRef(u.Login, repo, ref)
	if err != nil {
		return err
	}

	return nil
}
