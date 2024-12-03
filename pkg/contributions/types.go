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
	"time"
)

type ActivityReport interface {
	GenerateActivityLog() string
	GenerateLogFileName(userName string) string
}

type UserActivityReportInOrg struct {
	Collection *ContributionsCollection
	Org        string
	UserName   string
	StartFrom  time.Time
}

func (u *UserActivityReportInOrg) GenerateActivityLog() string {
	return fmt.Sprintf(`activity log:
	user:         %s
	organization: %s
	since:        %s

	hasContributions:                    %t
	totalIssueContributions:             %d
	totalPullRequestContributions:       %d
	totalPullRequestReviewContributions: %d
	totalCommitContributions:            %d

`, u.UserName, u.Org, u.StartFrom.Format(time.DateTime),
		u.Collection.HasAnyContributions,
		u.Collection.TotalIssueContributions,
		u.Collection.TotalPullRequestContributions,
		u.Collection.TotalPullRequestReviewContributions,
		u.Collection.TotalCommitContributions,
	)
}

func (u *UserActivityReportInOrg) GenerateLogFileName(userName string) string {
	return fmt.Sprintf("user-activity-%s-%s-*.yaml", userName, u.Org)
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
	user:       %s
	repository: %s/%s
	since:      %s

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

type IssueContributionNodeFragment struct {
	URL string `yaml:"URL"`
}

type IssueContributionNode struct {
	Issue      IssueContributionNodeFragment `yaml:"issue"`
	OccurredAt time.Time                     `yaml:"occurredAt"`
}

type IssueContributions struct {
	TotalCount int                     `yaml:"totalCount"`
	Nodes      []IssueContributionNode `yaml:"nodes"`
}

type PullRequestContributionNodeFragment struct {
	URL string `yaml:"URL"`
}

type PullRequestContributionNode struct {
	PullRequest PullRequestContributionNodeFragment `yaml:"pullRequest"`
	OccurredAt  time.Time                           `yaml:"occurredAt"`
}

type PullRequestContributions struct {
	TotalCount int                           `yaml:"totalCount"`
	Nodes      []PullRequestContributionNode `yaml:"nodes"`
}

type PullRequestReviewContributionNodePullRequest struct {
	URL string `yaml:"URL"`
}
type PullRequestReviewContributionNodeRepository struct {
	NameWithOwner string `yaml:"nameWithOwner"`
}
type PullRequestReviewContributionNodeFragment struct {
	Repository  PullRequestReviewContributionNodeRepository  `yaml:"repository"`
	PullRequest PullRequestReviewContributionNodePullRequest `yaml:"pullRequest"`
	CreatedAt   time.Time                                    `yaml:"createdAt"`
	State       string                                       `yaml:"state"`
}
type PullRequestReviewContributionNode struct {
	PullRequestReview PullRequestReviewContributionNodeFragment `yaml:"pullRequestReview"`
}
type PullRequestReviewContributions struct {
	TotalCount int                                 `yaml:"totalCount"`
	Nodes      []PullRequestReviewContributionNode `yaml:"nodes"`
}

type CommitContributionsByRepositoryContributionUser struct {
	Name string `yaml:"name"`
}
type CommitContributionsByRepositoryContributionRepository struct {
	NameWithOwner string `yaml:"nameWithOwner"`
}

type CommitContributionsByRepositoryContributionsNode struct {
	Repository CommitContributionsByRepositoryContributionRepository `yaml:"repository"`
	User       CommitContributionsByRepositoryContributionUser       `yaml:"user"`
	OccurredAt time.Time                                             `yaml:"occurredAt"`
}

type CommitContributionsByRepositoryContributions struct {
	TotalCount int                                                `yaml:"totalCount"`
	Nodes      []CommitContributionsByRepositoryContributionsNode `yaml:"nodes"`
}
type CommitContributionsByRepository struct {
	Contributions CommitContributionsByRepositoryContributions `graphql:"contributions(first: 10,orderBy: {field: OCCURRED_AT, direction: DESC})"`
}

type ContributionsCollection struct {
	HasAnyContributions                 bool `yaml:"hasAnyContributions"`
	TotalCommitContributions            int  `yaml:"totalCommitContributions"`
	TotalIssueContributions             int  `yaml:"totalIssueContributions"`
	TotalPullRequestContributions       int  `yaml:"totalPullRequestContributions"`
	TotalPullRequestReviewContributions int  `yaml:"totalPullRequestReviewContributions"`
	IssueContributions                  `graphql:"issueContributions(first: 1, orderBy: {direction: DESC})" yaml:"issueContributions"`
	PullRequestContributions            `graphql:"pullRequestContributions(first: 1, orderBy: {direction: DESC})" yaml:"pullRequestContributions"`
	PullRequestReviewContributions      `graphql:"pullRequestReviewContributions(first: 1, orderBy: {direction: DESC})" yaml:"pullRequestReviewContributions"`
	CommitContributionsByRepository     []CommitContributionsByRepository `graphql:"commitContributionsByRepository(maxRepositories: 10)" yaml:"commitContributionsByRepository"`
}

type UserContributionsInOrg struct {
	ContributionsCollection `graphql:"contributionsCollection(organizationID: $organizationID, from: $startFrom)" yaml:"contributionsCollection"`
}
