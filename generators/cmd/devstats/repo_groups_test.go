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
	"strings"
	"testing"
)

func TestRepoGroupsTemplate(t *testing.T) {
	testCases := []struct {
		name                    string
		templateData            RepoGroupsTemplateData
		expectedOutputContained string
		expectedErr             error
	}{
		{
			name: "one group",
			templateData: RepoGroupsTemplateData{
				RepoGroups: []RepoGroup{
					{
						Name:  "sig-testing",
						Alias: "blah",
						Repos: []string{
							"kubevirt/ci",
							"kubevirt/test",
						},
					},
				},
			},
			expectedOutputContained: `insert into gha_repo_groups(id, name, repo_group, alias, org_id, org_login)
select id, name, 'sig-testing', alias, org_id, org_login
  from gha_repo_groups
 where lower(name) in (
          'kubevirt/ci',
          'kubevirt/test',
       )
  on conflict update;`,
			expectedErr: nil,
		},
		{
			name: "two groups",
			templateData: RepoGroupsTemplateData{
				RepoGroups: []RepoGroup{
					{
						Name:  "sig-testing",
						Alias: "blah",
						Repos: []string{
							"kubevirt/kubevirt",
							"kubevirt/test",
						},
					},
					{
						Name:  "sig-ci",
						Alias: "bled",
						Repos: []string{
							"kubevirt/ci-health",
							"kubevirt/kubevirtci",
						},
					},
				},
			},
			expectedOutputContained: `
insert into gha_repo_groups(id, name, repo_group, alias, org_id, org_login)
select id, name, 'sig-testing', alias, org_id, org_login
  from gha_repo_groups
 where lower(name) in (
          'kubevirt/kubevirt',
          'kubevirt/test',
       )
  on conflict update;

insert into gha_repo_groups(id, name, repo_group, alias, org_id, org_login)
select id, name, 'sig-ci', alias, org_id, org_login
  from gha_repo_groups
 where lower(name) in (
          'kubevirt/ci-health',
          'kubevirt/kubevirtci',
       )
  on conflict update;`,
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sql, err := generateRepoGroupsSQL(testCase.templateData)
			if !strings.Contains(sql, testCase.expectedOutputContained) {
				t.Log(sql)
				t.Errorf(`wanted output to contain:
%s`, testCase.expectedOutputContained)
			}
			if testCase.expectedErr != err {
				t.Errorf("got %q, want %q", err, testCase.expectedErr)
			}
		})
	}
}
