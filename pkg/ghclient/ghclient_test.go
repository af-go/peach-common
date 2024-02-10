package ghclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/af-go/peach-common/pkg/log"
)

var (
	organizationName = ""
	logger           = log.NewLogger(true)
)

func TestGet(t *testing.T) {
	ctx := context.TODO()
	client := NewClient(&ClientOptions{}, logger)
	if client == nil {
		t.Fatalf("failed to creat ghclient")
	}
	owner := client.GetOwner(ctx, organizationName)
	repo, err := client.GetRepository(ctx, owner, "git-test")
	if err != nil {
		t.Fatalf("failed to get repository")
	}
	fmt.Printf("Get Repository:\n")
	fmt.Printf("  Name: %v\n", repo.Name)
	fmt.Printf("  FullName: %v\n", repo.FullName)
	fmt.Printf("  URL: %v\n", repo.URL)
}

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	client := NewClient(&ClientOptions{}, logger)
	if client == nil {
		t.Fatalf("failed to creat ghclient")
	}
	owner := client.GetOwner(ctx, organizationName)
	repo, err := client.CreateRepository(ctx, owner, "git-test2", nil)
	if err != nil {
		t.Fatalf("failed to create repository")
	}
	fmt.Printf("Get Repository:\n")
	fmt.Printf("  Name: %v\n", repo.Name)
	fmt.Printf("  FullName: %v\n", repo.FullName)
	fmt.Printf("  URL: %v\n", repo.URL)
}

/*
func TestBranchCommitReview(t *testing.T) {
	branchName := "test5"
	repositoryName := "git-test"
	submiter := NewClient(&ClientOptions{Author: "Joe", Email: "joe@example.com", TokenEnvVar: "GITHUB_KEY"}, logger)
	approver := NewClient(&ClientOptions{TokenEnvVar: "GITHUB_APPROVER_KEY"}, logger)

	ctx := context.TODO()
	owner := submiter.GetOwner(ctx, organizationName)

	err := submiter.CreateBranch(ctx, owner, repositoryName, branchName)
	if err != nil {
		t.Fatalf("create branch %s failed %v", branchName, err)
	}
	contents := make(map[string]string)
	contents[".trigger"] = "Hello Trigger"

	err = submiter.CreateAndPushCommit(ctx, owner, repositoryName, branchName, "create commit test", contents)
	if err != nil {
		t.Fatalf("failed to create and push commit")
	}
	number, err := submiter.CreatePullRequest(ctx, owner, repositoryName, "test pull request", "just for test", "", branchName)
	if err != nil {
		t.Fatalf("failed to create pull request")
	}
	err = submiter.AddPullRequestComment(ctx, owner, repositoryName, number, "new comment")
	if err != nil {
		t.Fatalf("failed to add comment to pull request")
	}
	_, err = submiter.AddReviewComment(ctx, owner, repositoryName, number, "new review comment")
	if err != nil {
		t.Fatalf("failed to add review comment")
	}

	err = approver.SubmitReview(ctx, owner, repositoryName, number, "approved")
	if err != nil {
		t.Fatalf("failed to approve review %v", err)
	}
	err = submiter.ClosePullRequest(ctx, owner, repositoryName, number)
	if err != nil {
		t.Fatalf("falied to close pull request")
	}

}
*/
