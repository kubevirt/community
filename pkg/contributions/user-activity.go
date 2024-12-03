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
	"context"
	"fmt"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

type ContributionReportGenerator struct {
	client *githubv4.Client
	opts   ContributionReportGeneratorOptions
}

func NewContributionReportGenerator(opts ContributionReportGeneratorOptions) (*ContributionReportGenerator, error) {
	token, err := os.ReadFile(opts.GithubTokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to use github token path %s: %v", opts.GithubTokenPath, err)
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)
	return &ContributionReportGenerator{client: client, opts: opts}, nil
}

func (g ContributionReportGenerator) GenerateReport(userName string) error {
	var activity ActivityReport
	var err error
	if g.opts.Repo != "" {
		activity, err = generateUserActivityReportInRepository(g.client, g.opts.Org, g.opts.Repo, userName, g.opts.startFrom())
	} else {
		activity, err = generateUserActivityReportInOrganization(g.client, g.opts.Org, userName, g.opts.startFrom())
	}
	if err != nil {
		return fmt.Errorf("failed to query: %v", err)
	}
	fmt.Printf(activity.GenerateActivityLog())
	err = writeActivityToFile(activity, "/tmp", activity.GenerateLogFileName(userName))
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}

type ContributionReportGeneratorOptions struct {
	Org             string
	Repo            string
	GithubTokenPath string
	Months          int
}

func (o ContributionReportGeneratorOptions) validate() error {
	if o.GithubTokenPath == "" {
		return fmt.Errorf("github token path is required")
	}
	return nil
}

func (o ContributionReportGeneratorOptions) startFrom() time.Time {
	return time.Now().AddDate(0, -1*o.Months, 0)
}

func generateUserActivityReportInRepository(client *githubv4.Client, org, repo, username string, startFrom time.Time) (*UserActivityReportInRepository, error) {
	userid, err := getUserId(client, username)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %v", err)
	}

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

	err = client.Query(context.Background(), &query, variables)
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
		Org:                   org,
		Repo:                  repo,
		UserName:              username,
		UserID:                userid,
		StartFrom:             startFrom,
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

func generateUserActivityReportInOrganization(client *githubv4.Client, org, username string, startFrom time.Time) (*UserActivityReportInOrg, error) {
	organizationId, err := getOrganizationId(client, org)
	if err != nil {
		return nil, err
	}

	var query struct {
		UserContributionsInOrg UserContributionsInOrg `graphql:"userContributionsInOrg: user(login: $username)"`
	}

	variables := map[string]interface{}{
		"username":       githubv4.String(username),
		"organizationID": githubv4.ID(organizationId),
		"startFrom":      githubv4.DateTime{Time: startFrom},
	}

	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}

	collection := &query.UserContributionsInOrg.ContributionsCollection
	result := &UserActivityReportInOrg{
		Collection: collection,
		Org:        org,
		UserName:   username,
		StartFrom:  startFrom,
	}
	return result, nil
}

func writeActivityToFile(yamlObject interface{}, dir, fileName string) error {
	tempFile, err := os.CreateTemp(dir, fileName)
	defer tempFile.Close()
	encoder := yaml.NewEncoder(tempFile)
	err = encoder.Encode(&yamlObject)
	if err != nil {
		return err
	}
	fmt.Printf(`user activity log: %q`, tempFile.Name())
	return nil
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
