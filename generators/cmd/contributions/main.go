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
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"kubevirt.io/community/pkg/contributions"
	"os"
)

type ContributionReportOptions struct {
	org             string
	repo            string
	username        string
	githubTokenPath string
	months          int
}

func (o ContributionReportOptions) validate() error {
	if o.username == "" {
		return fmt.Errorf("username is required")
	}
	if o.githubTokenPath == "" {
		return fmt.Errorf("github token path is required")
	}
	return nil
}

func (o ContributionReportOptions) MakeGeneratorOptions() contributions.ContributionReportGeneratorOptions {
	return contributions.ContributionReportGeneratorOptions{
		Org:             o.org,
		Repo:            o.repo,
		GithubTokenPath: o.githubTokenPath,
		Months:          o.months,
	}
}

func gatherContributionReportOptions() (ContributionReportOptions, error) {
	o := ContributionReportOptions{}
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
	contributionReportOptions, err := gatherContributionReportOptions()
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	if err = contributionReportOptions.validate(); err != nil {
		log.Fatalf("error validating arguments: %v", err)
	}
	err = generateReport(
		[]string{contributionReportOptions.username},
		contributionReportOptions.MakeGeneratorOptions(),
	)
	if err != nil {
		log.Fatalf("failed to generate report: %v", err)
	}
}

func generateReport(userNames []string, opts contributions.ContributionReportGeneratorOptions) error {
	generator, err := contributions.NewContributionReportGenerator(opts)
	if err != nil {
		return fmt.Errorf("failed to create report generator: %v", err)
	}
	for _, userName := range userNames {
		if err := generator.GenerateReport(userName); err != nil {
			return err
		}
	}
	return nil
}
