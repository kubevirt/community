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
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"kubevirt.io/community/pkg/sigs"
	"log"
	"os"
	"regexp"
	"sort"
	"text/template"
)

type options struct {
	sigsYAMLPath string
	outputPath   string
}

func (o *options) Validate() error {
	if o.sigsYAMLPath == "" {
		return fmt.Errorf("path to sigs.yaml is required")
	}
	if _, err := os.Stat(o.sigsYAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", o.sigsYAMLPath)
	}
	return nil
}

func gatherOptions() options {
	o := options{}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.sigsYAMLPath, "sigs-yaml-path", "./sigs.yaml", "path to file sigs.yaml")
	fs.StringVar(&o.outputPath, "output-path", "/tmp/repo_groups.sql", "path to file to write the output into")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("error parsing arguments %v: %v", os.Args[1:], err)
	}
	return o
}

var repoNameMatcher = regexp.MustCompile(`^https://raw.githubusercontent.com/([^/]+/[^/]+)/.*$`)

func main() {
	opts := gatherOptions()
	if err := opts.Validate(); err != nil {
		log.Fatalf("invalid arguments: %v", err)
	}

	sigsYAML, err := sigs.ReadFile(opts.sigsYAMLPath)
	if err != nil {
		log.Fatalf("failed to read sigs.yaml: %v", err)
	}

	d := extractRepoGroups(sigsYAML)

	sql, err := generateRepoGroupsSQL(d)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to generate sql: %w", err))
	}

	file, err := os.OpenFile(opts.outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to write to file %q, %w", opts.outputPath, err))
	}
	defer file.Close()
	_, err = file.WriteString(sql)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to write to file %q, %w", opts.outputPath, err))
	}

	log.Printf("output written to %q", opts.outputPath)
}

func extractRepoGroups(sigsYAML *sigs.Sigs) RepoGroupsTemplateData {
	var d RepoGroupsTemplateData
	for _, sig := range sigsYAML.Sigs {
		repoGroup := RepoGroup{
			Name:  sig.Name,
			Alias: sig.Dir,
		}
		repoMap := make(map[string]struct{})
		for _, subProject := range sig.SubProjects {
			for _, ownerRef := range subProject.Owners {
				stringSubmatch := repoNameMatcher.FindStringSubmatch(ownerRef)
				if stringSubmatch == nil {
					log.Fatalf("ownerRef %q doesn't match!", ownerRef)
				}
				repoName := stringSubmatch[1]
				if _, exists := repoMap[repoName]; !exists {
					repoMap[repoName] = struct{}{}
				}
			}
		}
		if len(repoMap) == 0 {
			continue
		}
		var repos []string
		for repo := range repoMap {
			repos = append(repos, repo)
		}
		sort.Strings(repos)
		repoGroup.Repos = repos
		d.RepoGroups = append(d.RepoGroups, repoGroup)
	}
	return d
}

//go:embed repo_groups.gosql
var repoGroupsSQLTemplate string

func generateRepoGroupsSQL(d RepoGroupsTemplateData) (string, error) {
	templateInstance, err := template.New("repo_groups").Parse(repoGroupsSQLTemplate)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	err = templateInstance.Execute(&output, d)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
