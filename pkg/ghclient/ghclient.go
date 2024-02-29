package ghclient

/**
   For github enterprise, the endpoint MUST include /api/v3, for example, https://github.example.com/api/v3
**/
import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

	"github.com/go-logr/logr"

	"github.com/google/go-github/v59/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

const (
	GITHUB_TOEKN_ENV_VAR = "GITHUB_TOKEN"
)

var clients map[string]*Client = make(map[string]*Client)

var ErrGithubTokenNotFound = fmt.Errorf("github token is not found, unable to init github client")
var ErrGithubClientInitFailed = fmt.Errorf("failed to init github client")

// Repository general infomration about repository
type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	URL      string `json:"url"`
}

type ClientOptions struct {
	Endpoint    string `json:"endpoint" yaml:"endpoint"`
	TokenEnvVar string `json:"tokenEnvVar" yaml:"tokenEnvVar"`
	Author      string `json:"author" yaml:"author"`
	Email       string `json:"email" yaml:"email"`
}

// NewClient create github client, init if not exist
func NewClient(options *ClientOptions, logger *logr.Logger) *Client {
	key := fmt.Sprintf("%s_%s", options.Endpoint, options.TokenEnvVar)
	c, ok := clients[key]
	if !ok {
		logger.V(0).Info("failed to find client, creating", "key", key)
		c = &Client{options: *options, logger: logger}
		err := c.init()
		if err != nil {
			logger.Error(err, "failed to init client")
			return nil
		}
		clients[key] = c
	}
	logger.V(0).Info("ghclient is ready", "key", key)
	return c
}

// Client github client
type Client struct {
	logger  *logr.Logger
	options ClientOptions
	client  *github.Client
}

func (c *Client) init() error {
	var err error
	tokenEnvVar := GITHUB_TOEKN_ENV_VAR
	if c.options.TokenEnvVar != "" {
		tokenEnvVar = c.options.TokenEnvVar
	}
	c.logger.V(0).Info("using env var for ghclient", "env", tokenEnvVar)
	// Get github token from env
	token, ok := os.LookupEnv(tokenEnvVar)
	if !ok {
		c.logger.Error(ErrGithubTokenNotFound, "failed to get github token from environment", "var", tokenEnvVar)
		return ErrGithubTokenNotFound
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	if c.options.Endpoint == "" || c.options.Endpoint == "github.com" {
		c.client = github.NewClient(tc)
	} else {
		c.client, err = github.NewEnterpriseClient(c.options.Endpoint, c.options.Endpoint, tc)
		if err != nil {
			c.logger.Error(ErrGithubClientInitFailed, "failed to init github enterprise client", "endpoint", c.options.Endpoint)
			return err
		}
	}
	return nil
}

// CreateRepository create repository
func (c *Client) CreateRepository(ctx context.Context, owner string, name string, options map[string]string) (*Repository, error) {
	repo := &github.Repository{
		Name:    github.String(name),
		Private: github.Bool(true),
	}
	repoOwner := owner
	if owner == c.GetCurrentUser(ctx) {
		c.logger.V(0).Info("this repository belong to user directly", "name", name)
		repoOwner = ""
	}
	createdRepo, _, err := c.client.Repositories.Create(ctx, repoOwner, repo)
	if err != nil {
		c.logger.Error(err, "failed to create repository", "owner", repoOwner, "name", name)
		return nil, err
	}
	return &Repository{Name: name, URL: *createdRepo.SSHURL, FullName: buildFullNameIncludeHost(c.options.Endpoint, *createdRepo.FullName)}, nil
}

// ListRepositories list repository
func (c *Client) ListRepositories(ctx context.Context, owner string, filter string) ([]Repository, error) {
	repos, _, err := c.client.Repositories.List(ctx, owner, nil)
	if err != nil {
		return nil, err
	}
	repositories := []Repository{}
	for _, repo := range repos {
		repositories = append(repositories, Repository{Name: *repo.Name, URL: *repo.SSHURL, FullName: buildFullNameIncludeHost(c.options.Endpoint, *repo.FullName)})
	}
	return repositories, nil
}

// GetRepository get repository by name
func (c *Client) GetRepository(ctx context.Context, owner string, name string) (*Repository, error) {

	repo, err := c.getRepositroy(ctx, owner, name)
	if err != nil {
		c.logger.Error(err, "faile to find repository", "owner", owner, "name", name)
		return nil, err
	}
	return &Repository{Name: *repo.Name, URL: *repo.SSHURL, FullName: buildFullNameIncludeHost(c.options.Endpoint, *repo.FullName)}, nil
}

// CreatePullRequest create pull request
func (c *Client) CreatePullRequest(ctx context.Context, owner string, name string, subject string, description string, baseBranch string, headBranch string) (int, error) {
	if baseBranch == "" {
		repo, err := c.getRepositroy(ctx, owner, name)
		if err != nil {
			c.logger.Error(err, "faile to find repository", "owner", owner, "name", name)
			return 0, err
		}
		baseBranch = *repo.DefaultBranch
	}
	newPR := &github.NewPullRequest{
		Title:               &subject,
		Head:                &headBranch,
		Base:                &baseBranch,
		Body:                &description,
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := c.client.PullRequests.Create(ctx, owner, name, newPR)
	if err != nil {
		c.logger.Error(err, "failed to create pull request", "owner", owner, "name", name, "subject", subject)
		return 0, err
	}
	c.logger.Info("pull request is created", "url", pr.GetHTMLURL())
	return *pr.Number, nil
}

// ClosePullRequest close pull request
func (c *Client) ClosePullRequest(ctx context.Context, owner string, name string, number int) error {
	closedState := "closed"
	input := &github.PullRequest{State: &closedState}
	_, _, err := c.client.PullRequests.Edit(ctx, owner, name, number, input)
	if err != nil {
		c.logger.Error(err, "failed to clost pull request", "owner", owner, "name", name, "num", number)
		return err
	}
	c.logger.Info("pull request is closed", "owner", owner, "name", name, "num", number)
	return nil
}

// AddPullRequestComment add comment to pull request
func (c *Client) AddPullRequestComment(ctx context.Context, owner string, name string, number int, content string) error {
	input := &github.IssueComment{
		Body: github.String(content),
	}
	// A PullRequest is an issue, but not all issues are pull requests.
	// Pulls.CreateComment doesn't work. Use issues API to manage comment in pull request.
	// See https://docs.github.com/en/rest/reference/pulls for detail
	_, _, err := c.client.Issues.CreateComment(ctx, owner, name, number, input)
	if err != nil {
		c.logger.Error(err, "failed to add comment to pull request", "owner", owner, "name", name, "num", number)
		return err
	}
	c.logger.Info("one comment has been added to pull request", "owner", owner, "name", name, "num", number)
	return nil
}

// AddReviewComment add review comment
func (c *Client) AddReviewComment(ctx context.Context, owner, name string, number int, content string) (int64, error) {
	input := &github.PullRequestReviewRequest{
		Body:  github.String(content),
		Event: github.String("COMMENT"),
	}
	review, _, err := c.client.PullRequests.CreateReview(ctx, owner, name, number, input)
	if err != nil {
		c.logger.Error(err, "failed to add review to pull request", "owner", owner, "name", name, "num", number)
		return 0, err
	}
	c.logger.Info("one review has been added to pull request", "owner", owner, "name", name, "num", number)
	return review.GetID(), nil
}

// SubmitReview submit a review: comment, approve or request change
func (c *Client) SubmitReview(ctx context.Context, owner string, name string, number int, content string) error {
	input := &github.PullRequestReviewRequest{
		Body:  github.String(content),
		Event: github.String("APPROVE"),
	}
	_, _, err := c.client.PullRequests.CreateReview(ctx, owner, name, number, input)

	if err != nil {
		c.logger.Error(err, "failed to approve review for pull request", "owner", owner, "name", name, "num", number)
		return err
	}
	c.logger.Info("pull request is approved", "owner", owner, "name", name, "num", number)
	return nil
}

// MergePullRequest merge pull request
func (c *Client) MergePullRequest(ctx context.Context, owner string, name string, number int) {

}

// CreateAndPushCommit create a commit with files (tree) to a ref(branch)
func (c *Client) CreateAndPushCommit(ctx context.Context, owner string, name string, branch string, message string, contents map[string]string) error {
	ref, _, err := c.client.Git.GetRef(ctx, owner, name, fmt.Sprintf("refs/heads/%s", branch))
	if err != nil {
		c.logger.Error(err, "failed to get ref (branch)", "owner", owner, "name", name, "branch", branch)
		return err
	}
	tree, err := c.getTree(ctx, owner, name, ref, contents)
	if err != nil {
		c.logger.Error(err, "failed to build tree in repository", "owner", owner, "name", name)
		return err
	}
	parent, _, err := c.client.Repositories.GetCommit(ctx, owner, name, *ref.Object.SHA, &github.ListOptions{})
	if err != nil {
		c.logger.Error(err, "failed to get commit", "owner", owner, "name", name)
		return err
	}
	parent.Commit.SHA = parent.SHA

	date := github.Timestamp{}
	author := &github.CommitAuthor{Date: &date, Name: &c.options.Author, Email: &c.options.Email}
	parents := []*github.Commit{parent.Commit}
	commit := &github.Commit{Author: author, Message: &message, Tree: tree, Parents: parents}
	newCommit, _, err := c.client.Git.CreateCommit(ctx, owner, name, commit, &github.CreateCommitOptions{})
	if err != nil {
		c.logger.Error(err, "failed to create commit", "owner", owner, "name", name)
		return err
	}

	// Attach the commit to the master branch.
	ref.Object.SHA = newCommit.SHA
	_, _, err = c.client.Git.UpdateRef(ctx, owner, name, ref, false)
	if err != nil {
		c.logger.Error(err, "failed to update ref", "owner", owner, "name", name)
		return err
	}
	return nil

}

// CreateBranch create a branch
func (c *Client) CreateBranch(ctx context.Context, owner string, name string, branch string) error {
	_, err := c.getRef(ctx, owner, name, branch)
	if err != nil {
		c.logger.Error(err, "failed to create branch", "owner", owner, "name", name, "branch", branch)
		return err
	}
	c.logger.Info("branch is ready", "owner", owner, "name", name, "branch", branch)
	return nil
}

// DeleteBranch delete a branch TODO support delete branch
func (c *Client) DeleteBranch(ctx context.Context, owner string, name string, branch string) error {
	_, err := c.client.Git.DeleteRef(ctx, owner, name, fmt.Sprintf("refs/heads/%s", branch))
	if err != nil {
		c.logger.Error(err, "failed to delete ref (branch)", "owner", owner, "name", name, "branch", branch)
		return err
	}
	c.logger.Info("branch is deleted", "owner", owner, "name", name, "branch", branch)
	return nil
}

func (c *Client) CreateSecretInRepository(ctx context.Context, owner string, name string, secretName string, secretValue string) error {
	publicKey, _, err := c.client.Actions.GetRepoPublicKey(ctx, owner, name)
	if err != nil {
		c.logger.Error(err, "no resource found", "resource", "repositroyPublicKey")
		return err
	}

	encryptedSecret, err := encryptSecretWithPublicKey(publicKey, secretName, secretValue)
	if err != nil {
		c.logger.Error(err, "failed encrypt secret", "secret", secretName)
		return err
	}

	if _, err := c.client.Actions.CreateOrUpdateRepoSecret(ctx, owner, name, encryptedSecret); err != nil {
		c.logger.Error(err, "failed to create secret", "repositorySecret", secretName)
		return err
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, owner string, name string, branchName string, path string, customDomain string) error {
	input := &github.Pages{
		Source: &github.PagesSource{
			Branch: &branchName,
			Path:   &path,
		},
		CNAME: &customDomain,
	}
	pages, response, err := c.client.Repositories.EnablePages(ctx, owner, name, input)
	fmt.Printf("%v\n%v\n%v\n", pages, response.Status, err)
	if err != nil && response.StatusCode != 409 {
		c.logger.Error(err, "failed to publish pages")
		return err
	}

	return nil
}

func (c *Client) getRepositroy(ctx context.Context, owner string, name string) (*github.Repository, error) {
	repo, _, err := c.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		c.logger.Error(err, "failed to find repository", "owner", owner, "name", name)
		return nil, err
	}
	c.logger.Info("repository is found", "owner", owner, "name", name)
	return repo, nil
}

// getTree get tree TODO support binary content
func (c *Client) getTree(ctx context.Context, owner string, name string, ref *github.Reference, contents map[string]string) (*github.Tree, error) {
	entries := []*github.TreeEntry{}
	for path, content := range contents {
		entries = append(entries, &github.TreeEntry{Path: github.String(path), Type: github.String("blob"), Content: github.String(string(content)), Mode: github.String("100644")})
	}
	tree, _, err := c.client.Git.CreateTree(ctx, owner, name, *ref.Object.SHA, entries)
	if err != nil {
		c.logger.Error(err, "failed to create tree", "owner", owner, "name", name)
		return nil, err
	}
	return tree, nil
}

// getRef get ref (brnach), create if not exist
func (c *Client) getRef(ctx context.Context, owner string, name string, branch string) (*github.Reference, error) {
	ref, _, err := c.client.Git.GetRef(ctx, owner, name, fmt.Sprintf("refs/heads/%s", branch))
	if err == nil {
		c.logger.Info("found ref (branch)", "owner", owner, "name", name, "branch", branch)
		return ref, nil
	}
	c.logger.Info("failed to find ref (branch), creating", "owner", owner, "name", name, "branch", branch)
	// Get default branch
	repo, err := c.getRepositroy(ctx, owner, name)
	if err != nil {
		c.logger.Error(err, "failed to find repository", "owner", owner, "name", name, "branch", branch)
		return nil, err
	}
	c.logger.Info("try default branch", "owner", owner, "name", name, "branch", *repo.DefaultBranch)
	baseRef, _, err := c.client.Git.GetRef(ctx, owner, name, fmt.Sprintf("refs/heads/%s", *repo.DefaultBranch))
	if err != nil {
		c.logger.Error(err, "faild found default branch", "owner", owner, "name", name, "branch", *repo.DefaultBranch)
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + branch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = c.client.Git.CreateRef(ctx, owner, name, newRef)
	if err != nil {
		c.logger.Error(err, "faild to create new ref (branch)", "owner", owner, "name", name, "branch", branch)
		return nil, err
	}
	c.logger.Error(err, "ref (branch) is created", "owner", owner, "name", name, "branch", branch)
	return ref, nil
}

func (c *Client) GetOwner(ctx context.Context, organizationName string) string {
	var owner = organizationName
	username := ""
	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		c.logger.Error(err, "failed to get user from context")
		return ""
	}
	if organizationName == "" {
		owner = user.GetLogin()
	}
	return owner
}

func (c *Client) GetCurrentUser(ctx context.Context) string {
	user, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		c.logger.Error(err, "failed to get user from context")
		return ""
	}
	return user.GetLogin()
}

func buildFullNameIncludeHost(endpoint string, fullName string) string {
	p := endpoint
	if p == "" || p == "github.com" {
		p = "https://github.com"
	}
	u, err := url.Parse(p)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v%v%v", u.Host, string(os.PathSeparator), fullName)
}

func encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName string, secretValue string) (*github.EncryptedSecret, error) {

	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, err
	}

	var boxKey [32]byte
	copy(boxKey[:], decodedPublicKey)
	secretBytes := []byte(secretValue)
	encryptedBytes, err := box.SealAnonymous([]byte{}, secretBytes, &boxKey, rand.Reader)
	if err != nil {
		return nil, err
	}

	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
	keyID := publicKey.GetKeyID()
	encryptedSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}
