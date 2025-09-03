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
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"kubevirt.io/community/pkg/contributions"
)

type ReportResult struct {
	ActiveUsers   []string            `yaml:"activeUsers"`
	InactiveUsers []string            `yaml:"inactiveUsers"`
	SkippedUsers  map[string][]string `yaml:"skippedUsers"`
}

func (receiver *ReportResult) SkipUser(reason, userName string) {
	if receiver.SkippedUsers == nil {
		receiver.SkippedUsers = make(map[string][]string)
	}
	receiver.SkippedUsers[reason] = append(receiver.SkippedUsers[reason], userName)
}

type Report struct {
	ReportOptions *contributionReportOptions `yaml:"reportOptions"`
	ReportConfig  *contributionReportConfig  `yaml:"reportConfig"`
	Result        *ReportResult              `yaml:"result"`
	Log           []string                   `yaml:"log"`
}

func NewReportWithConfiguration(options *contributionReportOptions, config *contributionReportConfig) *Report {
	return &Report{
		ReportConfig:  config,
		ReportOptions: options,
		Result:        &ReportResult{},
	}
}

type Reporter interface {
	Report(r contributions.ContributionReport, userName string) error
	Summary() string
	Full() *Report
	Skip(userName string, reason string)
}

type DefaultReporter struct {
	report *Report
}

func NewDefaultReporter(options *contributionReportOptions, config *contributionReportConfig) Reporter {
	d := &DefaultReporter{}
	d.report = NewReportWithConfiguration(options, config)
	return d
}

func (d *DefaultReporter) Skip(userName string, reason string) {
	d.report.Result.SkipUser(reason, userName)
}

func (d *DefaultReporter) Report(r contributions.ContributionReport, userName string) error {
	fmt.Print(r.Summary())
	_, err := r.WriteToFile("/tmp", userName)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}

func (d *DefaultReporter) Summary() string {
	return ""
}

func (d *DefaultReporter) Full() *Report {
	return d.report
}

type InactiveOnlyReporter struct {
	report *Report
}

func (d *InactiveOnlyReporter) Skip(userName string, reason string) {
	d.report.Result.SkipUser(reason, userName)
}

func NewInactiveOnlyReporter(options *contributionReportOptions, config *contributionReportConfig) Reporter {
	i := &InactiveOnlyReporter{}
	i.report = NewReportWithConfiguration(options, config)
	return i
}

func (d *InactiveOnlyReporter) Report(r contributions.ContributionReport, userName string) error {
	if r.HasContributions() {
		log.Debugf("active user: %s", userName)
		d.report.Result.ActiveUsers = append(d.report.Result.ActiveUsers, userName)
		return nil
	}
	log.Infof("inactive user: %s", userName)
	d.report.Log = append(d.report.Log, r.Summary())
	fileName, err := r.WriteToFile("/tmp", userName)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	d.report.Log = append(d.report.Log, fmt.Sprintf("activity log written to %q", fileName))
	d.report.Result.InactiveUsers = append(d.report.Result.InactiveUsers, userName)
	return nil
}

func (d *InactiveOnlyReporter) Summary() string {
	out, err := yaml.Marshal(d.report.Result.InactiveUsers)
	if err != nil {
		log.Fatalf("failed to serialize: %v", err)
	}
	return fmt.Sprintf(`inactive users:
%s`, string(out))
}

func (d *InactiveOnlyReporter) Full() *Report {
	return d.report
}
