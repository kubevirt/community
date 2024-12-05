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
	_ "embed"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"kubevirt.io/community/pkg/contributions"
	"kubevirt.io/community/pkg/orgs"
	"kubevirt.io/community/pkg/owners"
	"os"
	"sort"
	"strings"
)

type ContributionReportOptions struct {
	Org                  string `yaml:"org"`
	Repo                 string `yaml:"repo"`
	Username             string `yaml:"username"`
	GithubTokenPath      string `yaml:"githubTokenPath"`
	Months               int    `yaml:"months"`
	OrgsConfigFilePath   string `yaml:"orgsConfigFilePath"`
	OwnersFilePath       string `yaml:"ownersFilePath"`
	ReportAll            bool   `yaml:"reportAll"`
	ReportOutputFilePath string `yaml:"reportOutputFilePath"`
}

type SkipInactiveCheckConfig struct {
	Name   string   `yaml:"name"`
	Github []string `yaml:"github"`
}

type ContributionReportConfig struct {
	SkipInactive map[string][]SkipInactiveCheckConfig `yaml:"skipInactive"`
}

func (c *ContributionReportConfig) ShouldSkip(org, repo, userName string) (bool, string) {
	var skipInactiveKey string
	if repo != "" {
		skipInactiveKey = fmt.Sprintf("%s/%s", org, repo)
	} else {
		skipInactiveKey = org
	}
	configs, exists := c.SkipInactive[skipInactiveKey]
	if !exists {
		return false, ""
	}
	for _, config := range configs {
		for _, github := range config.Github {
			if strings.ToLower(userName) == strings.ToLower(github) {
				return true, config.Name
			}
		}
	}
	return false, ""
}

var (
	//go:embed default-config.yaml
	defaultConfigContent []byte

	defaultConfig *ContributionReportConfig
)

func (o ContributionReportOptions) validate() error {
	if o.Username != "" {
		log.Infof("creating report for user %q", o.Username)
	} else if o.OrgsConfigFilePath == "" && o.OwnersFilePath == "" {
		return fmt.Errorf("username or orgs-config-file-path or owners-file-path is required")
	}
	if o.GithubTokenPath == "" {
		return fmt.Errorf("github token path is required")
	}
	return nil
}

func (o ContributionReportOptions) makeGeneratorOptions() contributions.ContributionReportGeneratorOptions {
	return contributions.ContributionReportGeneratorOptions{
		Org:             o.Org,
		Repo:            o.Repo,
		GithubTokenPath: o.GithubTokenPath,
		Months:          o.Months,
	}
}

func gatherContributionReportOptions() (*ContributionReportOptions, error) {
	o := ContributionReportOptions{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.Org, "org", "kubevirt", "org name")
	fs.StringVar(&o.Repo, "repo", "", "repo name (leave empty to create an org activity report)")
	fs.StringVar(&o.Username, "username", "", "github handle")
	fs.IntVar(&o.Months, "months", 6, "months to look back for fetching data")
	fs.StringVar(&o.GithubTokenPath, "github-token", "/etc/github/oauth", "path to github token to use")
	fs.StringVar(&o.OrgsConfigFilePath, "orgs-file-path", "../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml", "file path to the orgs.yaml file to check")
	fs.StringVar(&o.OwnersFilePath, "owners-file-path", "", "file path to the OWNERS file to check")
	fs.BoolVar(&o.ReportAll, "report-all", false, "whether to only report inactive users or all users")
	fs.StringVar(&o.ReportOutputFilePath, "report-output-file-path", "", "file path to write the report output into")
	err := fs.Parse(os.Args[1:])
	return &o, err
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	err := yaml.Unmarshal(defaultConfigContent, &defaultConfig)
	if err != nil {
		log.Fatalf("error unmarshalling default config: %v", err)
	}
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

	reporter := NewDefaultReporter(contributionReportOptions, defaultConfig)
	userNames := []string{contributionReportOptions.Username}
	if contributionReportOptions.Username == "" {
		if !contributionReportOptions.ReportAll {
			reporter = NewInactiveOnlyReporter(contributionReportOptions, defaultConfig)
		}
		if contributionReportOptions.OwnersFilePath != "" {
			ownersYAML, err := owners.ReadFile(contributionReportOptions.OwnersFilePath)
			if err != nil {
				log.Fatalf("invalid arguments: %v", err)
			}
			userNames = uniq(ownersYAML.Reviewers, ownersYAML.Approvers)
			sort.Strings(userNames)
		} else if contributionReportOptions.OrgsConfigFilePath != "" {
			orgsYAML, err := orgs.ReadFile(contributionReportOptions.OrgsConfigFilePath)
			if err != nil {
				log.Fatalf("invalid arguments: %v", err)
			}
			userNames = orgsYAML.Orgs[contributionReportOptions.Org].Members
		}
	}

	for _, userName := range userNames {
		shouldSkip, reason := defaultConfig.ShouldSkip(contributionReportOptions.Org, contributionReportOptions.Repo, userName)
		if shouldSkip {
			log.Debugf("skipping user %s (reason: %s)", userName, reason)
			reporter.Skip(userName, reason)
			continue
		}
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
	if contributionReportOptions.ReportOutputFilePath != "" {
		reportBytes, err := yaml.Marshal(reporter.Full())
		if err != nil {
			log.Fatalf("failed to write report: %v", err)
		}
		err = os.WriteFile(contributionReportOptions.ReportOutputFilePath, reportBytes, 0666)
		if err != nil {
			log.Fatalf("failed to write report: %v", err)
		}
	}
}

func uniq(elements ...[]string) []string {
	uniqMap := make(map[string]struct{})
	for _, values := range elements {
		for _, value := range values {
			uniqMap[value] = struct{}{}
		}
	}
	var uniqueValues []string
	for uniqueValue := range uniqMap {
		uniqueValues = append(uniqueValues, uniqueValue)
	}
	return uniqueValues
}
