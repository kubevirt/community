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
	"gopkg.in/yaml.v3"
	"kubevirt.io/community/pkg/contributions"
	"kubevirt.io/community/pkg/orgs"
	"kubevirt.io/community/pkg/owners"
	"os"
	"sort"
)

type ContributionReportOptions struct {
	org                string
	repo               string
	username           string
	githubTokenPath    string
	months             int
	orgsConfigFilePath string
	ownersFilePath     string
	reportAll          bool
}

func (o ContributionReportOptions) validate() error {
	if o.username != "" {
		log.Infof("creating report for user %q", o.username)
	} else if o.orgsConfigFilePath == "" && o.ownersFilePath == "" {
		return fmt.Errorf("username or orgs-config-file-path or owners-file-path is required")
	}
	if o.githubTokenPath == "" {
		return fmt.Errorf("github token path is required")
	}
	return nil
}

func (o ContributionReportOptions) makeGeneratorOptions() contributions.ContributionReportGeneratorOptions {
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
	fs.StringVar(&o.orgsConfigFilePath, "orgs-file-path", "../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml", "file path to the orgs.yaml file to check")
	fs.StringVar(&o.ownersFilePath, "owners-file-path", "", "file path to the OWNERS file to check")
	fs.BoolVar(&o.reportAll, "report-all", false, "whether to only report inactive users or all users")
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

	generator, err := contributions.NewContributionReportGenerator(contributionReportOptions.makeGeneratorOptions())
	if err != nil {
		log.Fatalf("failed to create report generator: %v", err)
	}

	var reporter Reporter = DefaultReporter{}
	userNames := []string{contributionReportOptions.username}
	if contributionReportOptions.username == "" {
		if contributionReportOptions.ownersFilePath != "" {
			ownersYAML, err := owners.ReadFile(contributionReportOptions.ownersFilePath)
			if err != nil {
				log.Fatalf("invalid arguments: %v", err)
			}
			userNames = ownersYAML.Reviewers
			userNames = append(userNames, ownersYAML.Approvers...)
			sort.Strings(userNames)
			if !contributionReportOptions.reportAll {
				reporter = InactiveOnlyReporter{}
			}
		} else if contributionReportOptions.orgsConfigFilePath != "" {
			orgsYAML, err := orgs.ReadFile(contributionReportOptions.orgsConfigFilePath)
			if err != nil {
				log.Fatalf("invalid arguments: %v", err)
			}
			userNames = orgsYAML.Orgs[contributionReportOptions.org].Members
			if !contributionReportOptions.reportAll {
				reporter = InactiveOnlyReporter{}
			}
		}
	}

	for _, userName := range userNames {
		activity, err := generator.GenerateReport(userName)
		if err != nil {
			log.Fatalf("failed to generate report: %v", err)
		}
		err = reporter.Report(activity, userName)
		if err != nil {
			log.Fatalf("failed to report: %v", err)
		}
	}
	fmt.Printf(reporter.Summary())
}

type Reporter interface {
	Report(r contributions.ActivityReport, userName string) error
	Summary() string
}

type DefaultReporter struct{}

func (d DefaultReporter) Report(r contributions.ActivityReport, userName string) error {
	fmt.Printf(r.GenerateActivityLog())
	_, err := r.WriteToFile("/tmp", userName)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}

func (d DefaultReporter) Summary() string {
	return ""
}

type InactiveOnlyReporter struct {
	inactiveUsers []string
}

func (d InactiveOnlyReporter) Report(r contributions.ActivityReport, userName string) error {
	if r.HasActivity() {
		log.Debugf("active user: %s", userName)
		return nil
	}
	log.Infof("inactive user: %s", userName)
	fmt.Printf(r.GenerateActivityLog())
	_, err := r.WriteToFile("/tmp", userName)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	d.inactiveUsers = append(d.inactiveUsers, userName)
	return nil
}

func (d InactiveOnlyReporter) Summary() string {
	out, err := yaml.Marshal(d.inactiveUsers)
	if err != nil {
		log.Fatalf("failed to serialize: %v", err)
	}
	return fmt.Sprintf(`inactive users:
%s`, string(out))
}
