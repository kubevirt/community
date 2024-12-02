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

func generateUserActivityReportInOrganization(client *githubv4.Client, orgId, username string, startFrom time.Time) (*ContributionsCollection, error) {

	var query struct {
		UserContributionsInOrg UserContributionsInOrg `graphql:"userContributionsInOrg: user(login: $username)"`
	}

	variables := map[string]interface{}{
		"username":       githubv4.String(username),
		"organizationID": githubv4.ID(orgId),
		"startFrom":      githubv4.DateTime{Time: startFrom},
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}

	return &query.UserContributionsInOrg.ContributionsCollection, nil
}

func getOrganizationId(client *githubv4.Client, organizationName string) (string, error) {
	var query struct {
		Organization struct {
			ID string
		} `graphql:"organization(login: $organizationName)"`
	}
	variables := map[string]interface{}{
		"organizationName": githubv4.String(organizationName),
	}
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}
	return query.Organization.ID, nil
}
