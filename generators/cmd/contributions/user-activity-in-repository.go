/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright the KubeVirt Authors.
 *
 */

package main

import (
	"context"
	"fmt"
	"github.com/shurcooL/githubv4"
	"time"
)

type UserActivityReportInRepository struct {
	IssuesCreated         IssuesCreated         `yaml:"issuesCreated"`
	IssuesCommented       IssuesCommented       `yaml:"issuesCommented"`
	PullRequestsCreated   PullRequestsCreated   `yaml:"pullRequestsCreated"`
	PullRequestsReviewed  PullRequestsReviewed  `yaml:"pullRequestsReviewed"`
	PullRequestsCommented PullRequestsCommented `yaml:"pullRequestsCommented"`
	CommitsByUser         CommitsByUser         `yaml:"commitsByUser"`
}

type Repository struct {
	Name string `yaml:"name"`
}
type Author struct {
	Login string `yaml:"login"`
}
type IssueFragment struct {
	Number     int        `yaml:"number"`
	Title      string     `yaml:"title"`
	URL        string     `yaml:"URL"`
	Repository Repository `yaml:"repository"`
	Author     Author     `yaml:"author"`
	CreatedAt  time.Time  `yaml:"createdAt"`
}

type IssuesCreatedNodeItem struct {
	Issue IssueFragment `graphql:"... on Issue" yaml:"issue"`
}

type IssuesCreated struct {
	IssueCount int                     `yaml:"issueCount"`
	Nodes      []IssuesCreatedNodeItem `yaml:"nodes"`
}

type CommentAuthor struct {
	Login string `yaml:"login"`
}
type CommentItem struct {
	Author    CommentAuthor `yaml:"author"`
	CreatedAt time.Time     `yaml:"createdAt"`
	URL       string        `yaml:"URL"`
}

type Comments struct {
	Nodes []CommentItem `yaml:"nodes"`
}

type IssueWithCommentFragment struct {
	Number     int        `yaml:"number"`
	Title      string     `yaml:"title"`
	URL        string     `yaml:"URL"`
	Repository Repository `yaml:"repository"`
	Author     Author     `yaml:"author"`
	Comments   Comments   `graphql:"comments(first:100, orderBy: {field: UPDATED_AT, direction: ASC} )" yaml:"comments"`
}

type IssuesCommentedNodeItem struct {
	Issue IssueWithCommentFragment `graphql:"... on Issue" yaml:"issue"`
}

type IssuesCommented struct {
	IssueCount int                       `yaml:"issueCount"`
	Nodes      []IssuesCommentedNodeItem `yaml:"nodes"`
}

type PullRequestAuthor struct {
	Login string `yaml:"login"`
}

type PullRequestFragment struct {
	Number    int               `yaml:"number"`
	Title     string            `yaml:"title"`
	URL       string            `yaml:"URL"`
	CreatedAt time.Time         `yaml:"createdAt"`
	Author    PullRequestAuthor `yaml:"author"`
}

type PullRequestNodeItem struct {
	PullRequest PullRequestFragment `graphql:"... on PullRequest" yaml:"pullRequest"`
}

type PullRequestsCreated struct {
	IssueCount int                   `yaml:"issueCount"`
	Nodes      []PullRequestNodeItem `yaml:"nodes"`
}

type PullRequestReviewItem struct {
	State string `yaml:"state"`
	URL   string `yaml:"URL"`
}

type PullRequestReviews struct {
	TotalCount int                     `yaml:"totalCount"`
	Nodes      []PullRequestReviewItem `yaml:"nodes"`
}

type PullRequestReviewFragment struct {
	Title     string             `yaml:"title"`
	Number    int                `yaml:"number"`
	URL       string             `yaml:"URL"`
	CreatedAt time.Time          `yaml:"createdAt"`
	Reviews   PullRequestReviews `graphql:"reviews(first:5, author: $username)" yaml:"reviews"`
}

type PullRequestReviewNodeItem struct {
	PullRequestReview PullRequestReviewFragment `graphql:"... on PullRequest" yaml:"pullRequestReview"`
}

type PullRequestsReviewed struct {
	IssueCount int                         `yaml:"issueCount"`
	Nodes      []PullRequestReviewNodeItem `yaml:"nodes"`
}

type PullRequestCommentAuthor struct {
	Login string `yaml:"login"`
}

type PullRequestComment struct {
	Author    PullRequestCommentAuthor `yaml:"author"`
	CreatedAt time.Time                `yaml:"createdAt"`
	URL       string                   `yaml:"URL"`
}
type PullRequestCommentsItem struct {
	Nodes []PullRequestComment `yaml:"nodes"`
}

type PullRequestCommentedRepository struct {
	Name string `yaml:"name"`
}

type PullRequestCommentedAuthor struct {
	Login string `yaml:"login"`
}

type PullRequestCommentedFragment struct {
	Number     int                            `yaml:"number"`
	Title      string                         `yaml:"title"`
	URL        string                         `yaml:"URL"`
	Repository PullRequestCommentedRepository `yaml:"repository"`
	Author     PullRequestCommentedAuthor     `yaml:"author"`
	Comments   PullRequestCommentsItem        `graphql:"comments(first:100, orderBy: {field: UPDATED_AT, direction: ASC} )" yaml:"comments"`
}

type PullRequestCommentedItem struct {
	PullRequest PullRequestCommentedFragment `graphql:"... on PullRequest" yaml:"pullRequest"`
}

type PullRequestsCommented struct {
	IssueCount int                        `yaml:"issueCount"`
	Nodes      []PullRequestCommentedItem `yaml:"nodes"`
}

type AssociatedPullRequest struct {
	Number int    `yaml:"number"`
	Title  string `yaml:"title"`
	URL    string `yaml:"URL"`
}

type AssociatedPullRequests struct {
	Nodes []AssociatedPullRequest `yaml:"nodes"`
}

type CommitsByUserTargetHistoryNode struct {
	CommitUrl              string `yaml:"commitUrl"`
	AssociatedPullRequests `graphql:"associatedPullRequests(first: 5)" yaml:"associatedPullRequests"`
}

type CommitsByUserTargetHistory struct {
	TotalCount int                              `yaml:"totalCount"`
	Nodes      []CommitsByUserTargetHistoryNode `yaml:"nodes"`
}

type CommitsByUserTargetFragment struct {
	History CommitsByUserTargetHistory `graphql:"history(author: {id: $userID}, since: $startFrom)" yaml:"history"`
}

type CommitsByUserTargetItem struct {
	Fragment CommitsByUserTargetFragment `graphql:"... on Commit" yaml:"fragment"`
}

type CommitsByUserRef struct {
	Target CommitsByUserTargetItem `yaml:"target"`
}

type CommitsByUser struct {
	DefaultBranchRef CommitsByUserRef `yaml:"defaultBranchRef"`
}

func generateUserActivityReportInRepository(client *githubv4.Client, org, repo, username, userid string, startFrom time.Time) (*UserActivityReportInRepository, error) {

	var query struct {
		IssuesCreated         IssuesCreated         `graphql:"issuesCreated: search(first: 5, type: ISSUE, query: $authorSearchQuery)"`
		IssuesCommented       IssuesCommented       `graphql:"issuesCommented: search(first: 5, type: ISSUE, query: $commenterSearchQuery)"`
		PullRequestsCreated   PullRequestsCreated   `graphql:"prsCreated: search(type: ISSUE, first: 5, query: $pullRequestsCreatedQuery)"`
		PullRequestsReviewed  PullRequestsReviewed  `graphql:"prsReviewed: search(type: ISSUE, first: 5, query: $pullRequestsReviewedQuery)"`
		PullRequestsCommented PullRequestsCommented `graphql:"prsCommented: search(last: 100, type: ISSUE, query: $pullRequestsCommentedQuery)"`
		CommitsByUser         CommitsByUser         `graphql:"commitsByUser: repository(owner: $org, name: $repo)"`
	}

	fromDate := startFrom.Format("2006-01-02")

	variables := map[string]interface{}{
		"org":       githubv4.String(org),
		"repo":      githubv4.String(repo),
		"username":  githubv4.String(username),
		"userID":    githubv4.ID(userid),
		"startFrom": githubv4.GitTimestamp{Time: startFrom},
		"authorSearchQuery": githubv4.String(fmt.Sprintf(
			"repo:%s/%s author:%s is:issue created:>=%s",
			org,
			repo,
			username,
			fromDate,
		)),
		"commenterSearchQuery": githubv4.String(fmt.Sprintf(
			"repo:%s/%s commenter:%s is:issue created:>=%s",
			org,
			repo,
			username,
			fromDate,
		)),
		"pullRequestsCreatedQuery": githubv4.String(fmt.Sprintf(
			"repo:%s/%s author:%s is:pr created:>=%s",
			org,
			repo,
			username,
			fromDate,
		)),
		"pullRequestsReviewedQuery": githubv4.String(fmt.Sprintf(
			"repo:%s/%s reviewed-by:%s is:pr updated:>=%s",
			org,
			repo,
			username,
			fromDate,
		)),
		"pullRequestsCommentedQuery": githubv4.String(fmt.Sprintf(
			"repo:%s/%s commenter:%s is:pr updated:>=%s",
			org,
			repo,
			username,
			fromDate,
		)),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}
	return &UserActivityReportInRepository{
		IssuesCreated:         query.IssuesCreated,
		IssuesCommented:       query.IssuesCommented,
		PullRequestsCreated:   query.PullRequestsCreated,
		PullRequestsReviewed:  query.PullRequestsReviewed,
		PullRequestsCommented: query.PullRequestsCommented,
		CommitsByUser:         query.CommitsByUser,
	}, nil
}

func getUserId(client *githubv4.Client, username string) (string, error) {
	var query struct {
		User struct {
			ID string
		} `graphql:"user(login: $username)"`
	}
	variables := map[string]interface{}{
		"username": githubv4.String(username),
	}
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}
	return query.User.ID, nil
}
