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
	"flag"
	"fmt"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

type options struct {
	org             string
	repo            string
	username        string
	githubTokenPath string
	months          int
}

func (o options) validate() error {
	if o.username == "" {
		return fmt.Errorf("username is required")
	}
	if o.githubTokenPath == "" {
		return fmt.Errorf("github token path is required")
	}
	return nil
}

func gatherOptions() (options, error) {
	o := options{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.org, "org", "kubevirt", "org name")
	fs.StringVar(&o.repo, "repo", "", "repo name (leave empty to create an org activity report)")
	fs.StringVar(&o.username, "username", "", "github handle")
	fs.IntVar(&o.months, "months", 6, "months to look back for fetching data")
	fs.StringVar(&o.githubTokenPath, "github-token", "/etc/github/oauth", "path to github token to use")
	err := fs.Parse(os.Args[1:])
	return o, err
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	opts, err := gatherOptions()
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	if err = opts.validate(); err != nil {
		log.Fatalf("error validating arguments: %v", err)
	}

	token, err := os.ReadFile(opts.githubTokenPath)
	if err != nil {
		log.Fatalf("failed to use github token path %s: %v", opts.githubTokenPath, err)
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: strings.TrimSpace(string(token))},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	graphqlClient := githubv4.NewClient(httpClient)

	xMonthsAgo := time.Now().AddDate(0, -1*opts.months, 0)

	if opts.repo != "" {
		id, err := getUserId(graphqlClient, opts.username)
		if err != nil {
			log.Fatalf("failed to query: %v", err)
		}

		activity, err := generateUserActivityReportInRepository(graphqlClient, opts.org, opts.repo, opts.username, id, xMonthsAgo)
		if err != nil {
			log.Fatalf("failed to query: %v", err)
		}
		tempFile, err := os.CreateTemp("/tmp", fmt.Sprintf("user-activity-%s-%s_%s-*.yaml", opts.username, opts.org, opts.repo))
		if err != nil {
			log.Fatal(err)
		}
		defer tempFile.Close()
		encoder := yaml.NewEncoder(tempFile)
		err = encoder.Encode(&activity)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(`activity log:
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

user activity log: %q
`, opts.username, opts.org, opts.repo, xMonthsAgo.Format(time.DateTime),
			activity.IssuesCreated.IssueCount,
			activity.IssuesCommented.IssueCount,
			activity.PullRequestsReviewed.IssueCount,
			activity.PullRequestsCreated.IssueCount,
			activity.PullRequestsCommented.IssueCount,
			activity.CommitsByUser.DefaultBranchRef.Target.Fragment.History.TotalCount,
			tempFile.Name(),
		)
	} else {
		id, err := getOrganizationId(graphqlClient, opts.org)
		if err != nil {
			log.Fatalf("failed to query: %v", err)
		}

		activity, err := generateUserActivityReportInOrganization(graphqlClient, id, opts.username, xMonthsAgo)
		if err != nil {
			log.Fatalf("failed to query: %v", err)
		}
		tempFile, err := os.CreateTemp("/tmp", fmt.Sprintf("user-activity-%s-%s-*.yaml", opts.username, opts.org))
		if err != nil {
			log.Fatal(err)
		}
		defer tempFile.Close()
		encoder := yaml.NewEncoder(tempFile)
		err = encoder.Encode(&activity)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf(`activity log:
	user:         %s
	organization: %s
	since:        %s

	hasContributions:                    %t
	totalIssueContributions:             %d
	totalPullRequestContributions:       %d
	totalPullRequestReviewContributions: %d
	totalCommitContributions:            %d

user activity log: %q
`, opts.username, opts.org, xMonthsAgo.Format(time.DateTime),
			activity.HasAnyContributions,
			activity.TotalIssueContributions,
			activity.TotalPullRequestContributions,
			activity.TotalPullRequestReviewContributions,
			activity.TotalCommitContributions,
			tempFile.Name(),
		)
	}
}
