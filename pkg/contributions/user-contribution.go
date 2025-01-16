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
	"github.com/avast/retry-go"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

type ContributionReportGenerator struct {
	client        *githubv4.Client
	opts          ContributionReportGeneratorOptions
	ReportingMode interface{}
}

func NewContributionReportGenerator(opts ContributionReportGeneratorOptions) (*ContributionReportGenerator, error) {
	err := opts.validate()
	if err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}
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

func (g ContributionReportGenerator) GenerateReport(userName string) (ContributionReport, error) {
	var contributionReport ContributionReport
	err := retry.Do(
		func() error {
			var err error
			if g.opts.Repo != "" {
				contributionReport, err = generateUserActivityReportForRepository(g.client, g.opts.Org, g.opts.Repo, userName, g.opts.startFrom())
			} else {
				contributionReport, err = generateUserContributionReportForOrganization(g.client, g.opts.Org, userName, g.opts.startFrom())
			}
			if err != nil {
				log.Errorf("query failed (will retry): %v", err)
			}
			return err
		},
		retry.LastErrorOnly(true),
	)
	if contributionReport == nil && err != nil {
		return nil, fmt.Errorf("query failed (aborting): %v", err)
	}
	return contributionReport, nil
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

func generateUserActivityReportForRepository(client *githubv4.Client, org, repo, username string, startFrom time.Time) (*UserContributionReportForRepository, error) {
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
	return &UserContributionReportForRepository{
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

func generateUserContributionReportForOrganization(client *githubv4.Client, org, username string, startFrom time.Time) (*UserContributionReportForOrganization, error) {
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
		CommitsByUserInOrg    CommitsByUserInOrg    `graphql:"commitsByUserInOrg: organization(login: $org)"`
	}

	fromDate := startFrom.Format("2006-01-02")

	variables := map[string]interface{}{
		"org":       githubv4.String(org),
		"username":  githubv4.String(username),
		"userID":    githubv4.ID(userid),
		"startFrom": githubv4.GitTimestamp{Time: startFrom},
		"authorSearchQuery": githubv4.String(fmt.Sprintf(
			"org:%s author:%s is:issue created:>=%s",
			org,
			username,
			fromDate,
		)),
		"commenterSearchQuery": githubv4.String(fmt.Sprintf(
			"org:%s commenter:%s is:issue created:>=%s",
			org,
			username,
			fromDate,
		)),
		"pullRequestsCreatedQuery": githubv4.String(fmt.Sprintf(
			"org:%s author:%s is:pr created:>=%s",
			org,
			username,
			fromDate,
		)),
		"pullRequestsReviewedQuery": githubv4.String(fmt.Sprintf(
			"org:%s reviewed-by:%s is:pr updated:>=%s",
			org,
			username,
			fromDate,
		)),
		"pullRequestsCommentedQuery": githubv4.String(fmt.Sprintf(
			"org:%s commenter:%s is:pr updated:>=%s",
			org,
			username,
			fromDate,
		)),
	}

	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to use github query %+v with variables %v: %w", query, variables, err)
	}
	return &UserContributionReportForOrganization{
		IssuesCreated:         query.IssuesCreated,
		IssuesCommented:       query.IssuesCommented,
		PullRequestsCreated:   query.PullRequestsCreated,
		PullRequestsReviewed:  query.PullRequestsReviewed,
		PullRequestsCommented: query.PullRequestsCommented,
		CommitsByUserInOrg:    query.CommitsByUserInOrg,
		Org:                   org,
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

func writeActivityToFile(yamlObject interface{}, dir, fileName string) error {
	tempFile, err := os.CreateTemp(dir, fileName)
	if err != nil {
		return err
	}
	defer tempFile.Close()
	encoder := yaml.NewEncoder(tempFile)
	err = encoder.Encode(&yamlObject)
	if err != nil {
		return err
	}
	log.Debugf(`user activity log: %q`, tempFile.Name())
	return nil
}
