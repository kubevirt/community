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

package sigs

type Sigs struct {
	Sigs          []*Group `yaml:"sigs"`
	Workinggroups []*Group `yaml:"workinggroups"`
	Usergroups    []*Group `yaml:"usergroups"`
	Committees    []*Group `yaml:"committees"`
}

type Group struct {
	Name             string
	Dir              string        `yaml:",omitempty"`
	Description      string        `yaml:",omitempty"`
	MissionStatement string        `yaml:"mission_statement,omitempty"`
	Label            string        `yaml:",omitempty"`
	Leads            []*Lead       `yaml:",omitempty"`
	Leadership       *Leadership   `yaml:",omitempty"`
	Meetings         []*Meeting    `yaml:",omitempty"`
	Contact          *Contact      `yaml:",omitempty"`
	SubProjects      []*SubProject `yaml:",omitempty"`
}

type Contact struct {
	Slack       string     `yaml:"slack"`
	MailingList string     `yaml:"mailing_list"`
	Teams       []*Team    `yaml:"teams,omitempty"`
	Liaison     *OrgMember `yaml:"liaison,omitempty"`
}

type Team struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Meeting struct {
	Description   string
	Day           string
	Time          string
	TZ            string
	Frequency     string
	URL           string
	ArchiveURL    string `yaml:"archive_url,omitempty"`
	RecordingsURL string `yaml:"recordings_url,omitempty"`
}

type Leadership struct {
	Chairs []*Chair `yaml:",omitempty"`
}

type OrgMember struct {
	Github  string
	Name    string `yaml:",omitempty"`
	Company string `yaml:",omitempty"`
}

type Chair OrgMember
type Lead OrgMember

type SubProject struct {
	Name        string
	Description string `yaml:",omitempty"`
	Owners      []string
	Leads       []*Lead `yaml:",omitempty"`
}
