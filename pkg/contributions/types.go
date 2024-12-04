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

package contributions

import (
	"fmt"
	"path/filepath"
	"time"
)

type ActivityReport interface {
	GenerateActivityLog() string
	GenerateLogFileName(userName string) string
	HasActivity() bool
	WriteToFile(dir, userName string) (string, error)
}

type UserActivityReportInRepository struct {
	IssuesCreated         IssuesCreated         `yaml:"issuesCreated"`
	IssuesCommented       IssuesCommented       `yaml:"issuesCommented"`
	PullRequestsCreated   PullRequestsCreated   `yaml:"pullRequestsCreated"`
	PullRequestsReviewed  PullRequestsReviewed  `yaml:"pullRequestsReviewed"`
	PullRequestsCommented PullRequestsCommented `yaml:"pullRequestsCommented"`
	CommitsByUser         CommitsByUser         `yaml:"commitsByUser"`
	Org                   string
	Repo                  string
	UserName              string
	UserID                string
	StartFrom             time.Time
}

func (u *UserActivityReportInRepository) GenerateActivityLog() string {
	return fmt.Sprintf(`activity log:
    user:          %s
    repository:    %s/%s
    since:         %s

    issues
        created:   %d
        commented: %d
    pull requests:
        reviewed:  %d
        created:   %d
        commented: %d
    commits:       %d
`, u.UserName, u.Org, u.Repo, u.StartFrom.Format(time.DateTime),
		u.IssuesCreated.IssueCount,
		u.IssuesCommented.IssueCount,
		u.PullRequestsReviewed.IssueCount,
		u.PullRequestsCreated.IssueCount,
		u.PullRequestsCommented.IssueCount,
		u.CommitsByUser.DefaultBranchRef.Target.Fragment.History.TotalCount,
	)
}

func (u *UserActivityReportInRepository) GenerateLogFileName(userName string) string {
	return fmt.Sprintf("user-activity-%s-%s_%s-*.yaml", userName, u.Org, u.Repo)
}

func (u *UserActivityReportInRepository) HasActivity() bool {
	return u.IssuesCreated.IssueCount > 0 ||
		u.IssuesCommented.IssueCount > 0 ||
		u.PullRequestsReviewed.IssueCount > 0 ||
		u.PullRequestsCreated.IssueCount > 0 ||
		u.PullRequestsCommented.IssueCount > 0 ||
		u.CommitsByUser.DefaultBranchRef.Target.Fragment.History.TotalCount > 0
}

func (u *UserActivityReportInRepository) WriteToFile(dir string, userName string) (string, error) {
	logFileName := u.GenerateLogFileName(userName)
	err := writeActivityToFile(u, dir, logFileName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
}

type UserActivityReportInOrg2 struct {
	IssuesCreated         IssuesCreated         `yaml:"issuesCreated"`
	IssuesCommented       IssuesCommented       `yaml:"issuesCommented"`
	PullRequestsCreated   PullRequestsCreated   `yaml:"pullRequestsCreated"`
	PullRequestsReviewed  PullRequestsReviewed  `yaml:"pullRequestsReviewed"`
	PullRequestsCommented PullRequestsCommented `yaml:"pullRequestsCommented"`
	CommitsByUserInOrg    CommitsByUserInOrg    `yaml:"commitsByUserInOrg"`
	Org                   string
	UserName              string
	UserID                string
	StartFrom             time.Time
}

func (u *UserActivityReportInOrg2) GenerateActivityLog() string {
	return fmt.Sprintf(`activity log:
    user:          %s
    org:           %s
    since:         %s

    issues
        created:   %d
        commented: %d
    pull requests:
        reviewed:  %d
        created:   %d
        commented: %d
    commits:       %d
`, u.UserName, u.Org, u.StartFrom.Format(time.DateTime),
		u.IssuesCreated.IssueCount,
		u.IssuesCommented.IssueCount,
		u.PullRequestsReviewed.IssueCount,
		u.PullRequestsCreated.IssueCount,
		u.PullRequestsCommented.IssueCount,
		u.totalCommitCount(),
	)
}

func (u *UserActivityReportInOrg2) totalCommitCount() int {
	totalCommitCount := 0
	for _, node := range u.CommitsByUserInOrg.Repositories.Nodes {
		totalCommitCount += node.DefaultBranchRef.Target.Fragment.History.TotalCount
	}
	return totalCommitCount
}

func (u *UserActivityReportInOrg2) GenerateLogFileName(userName string) string {
	return fmt.Sprintf("user-activity-%s-%s-*.yaml", userName, u.Org)
}

func (u *UserActivityReportInOrg2) HasActivity() bool {
	return u.IssuesCreated.IssueCount > 0 ||
		u.IssuesCommented.IssueCount > 0 ||
		u.PullRequestsReviewed.IssueCount > 0 ||
		u.PullRequestsCreated.IssueCount > 0 ||
		u.PullRequestsCommented.IssueCount > 0 ||
		u.totalCommitCount() > 0
}

func (u *UserActivityReportInOrg2) WriteToFile(dir, userName string) (string, error) {
	logFileName := u.GenerateLogFileName(userName)
	err := writeActivityToFile(u, dir, logFileName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, logFileName), nil
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

type RepositoryNodeRefTargetHistoryNodeAuthorUser struct {
	Name string `yaml:"name"`
}

type RepositoryNodeRefTargetHistoryNodeAuthor struct {
	User RepositoryNodeRefTargetHistoryNodeAuthorUser `yaml:"user"`
}

type RepositoryNodeRefTargetHistoryNode struct {
	URL           string                                   `yaml:"URL"`
	CommittedDate time.Time                                `yaml:"committedDate"`
	Author        RepositoryNodeRefTargetHistoryNodeAuthor `yaml:"author"`
}

type RepositoryNodeRefTargetHistory struct {
	TotalCount int                                  `yaml:"totalCount"`
	Nodes      []RepositoryNodeRefTargetHistoryNode `yaml:"nodes"`
}

type RepositoryNodeRefTargetFragment struct {
	CommitURL string                         `yaml:"commitURL"`
	History   RepositoryNodeRefTargetHistory `graphql:"history(first: 3, author: {id: $userID}, since: $startFrom)" yaml:"history"`
}

type RepositoryNodeRefTargetItem struct {
	Fragment RepositoryNodeRefTargetFragment `graphql:"... on Commit" yaml:"fragment"`
}

type RepositoryNodeRef struct {
	Target RepositoryNodeRefTargetItem `yaml:"target"`
}

type RepositoryNode struct {
	Name             string            `yaml:"name"`
	DefaultBranchRef RepositoryNodeRef `yaml:"defaultBranchRef"`
}

type Repositories struct {
	Nodes []RepositoryNode `yaml:"nodes"`
}

type CommitsByUserInOrg struct {
	Repositories Repositories `graphql:"repositories(first: 25, isArchived: false, visibility: PUBLIC)" yaml:"repositories"`
}
