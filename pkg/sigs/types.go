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
	Sigs       []*Group `yaml:"sigs"`
	Usergroups []*Group `yaml:"usergroups"`
	Committees []*Group `yaml:"committees"`
}

type Group struct {
	Dir              string
	Name             string
	MissionStatement string         `yaml:"mission_statement,omitempty"`
	Label            string         `yaml:",omitempty"`
	Leadership       *Leadership    `yaml:",omitempty"`
	Meetings         []*Meeting     `yaml:",omitempty"`
	Contact          *Contact       `yaml:",omitempty"`
	SubProjects      []*SubProjects `yaml:",omitempty"`
}

type Contact struct {
	Slack       string  `yaml:"slack"`
	MailingList string  `yaml:"mailing_list"`
	Teams       []*Team `yaml:"teams"`
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
	ArchiveURL    string `yaml:"archive_url"`
	RecordingsURL string `yaml:"recordings_url"`
}

type Leadership struct {
	Chairs []*Chair
}

type Chair struct {
	Github  string
	Name    string
	Company string
}

type SubProjects struct {
	Name   string
	Owners []string
}
