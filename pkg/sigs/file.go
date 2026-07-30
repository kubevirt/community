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

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dirNameRegex = `^(sig|wg|ug|committee)-`
)

var (
	dirNameMatcher         = regexp.MustCompile(dirNameRegex)
	templateDirNameMatcher = regexp.MustCompile(dirNameRegex + `TEMPLATE`)
)

func ReadFile(path string) (*Sigs, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", path, err)
	}

	sigs := &Sigs{}
	err = yaml.Unmarshal(buf, sigs)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", path, err)
	}

	err = appendGroupsFromGroupDirectories(path, sigs)
	if err != nil {
		return nil, err
	}
	return sigs, err
}

// appendGroupsFromGroupDirectories looks at each group subdirectory candidate
// (sig-, wg-, ug-, committee-) for a file group.yaml and if that exists it
// appends its content to the respective group.
func appendGroupsFromGroupDirectories(path string, sigs *Sigs) error {
	dir := filepath.Dir(path)
	dirCandidates, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %q: %w", dir, err)
	}
	for _, groupDir := range dirCandidates {
		if !groupDir.IsDir() {
			continue
		}
		groupDirName := groupDir.Name()
		if templateDirNameMatcher.MatchString(groupDirName) || !dirNameMatcher.MatchString(groupDirName) {
			continue
		}

		groupFile := filepath.Join(dir, groupDirName, "group.yaml")
		buf, err := os.ReadFile(groupFile)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		groupEntry := &Group{}
		err = yaml.Unmarshal(buf, groupEntry)
		if err != nil {
			return fmt.Errorf("in file %q: %w", groupFile, err)
		}
		err = appendGroup(groupDirName, sigs, groupEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func appendGroup(groupDirName string, sigs *Sigs, groupEntry *Group) error {
	switch {
	case strings.HasPrefix(groupDirName, "sig-"):
		sigs.Sigs = append(sigs.Sigs, groupEntry)
	case strings.HasPrefix(groupDirName, "wg-"):
		sigs.Workinggroups = append(sigs.Workinggroups, groupEntry)
	case strings.HasPrefix(groupDirName, "ug-"):
		sigs.Usergroups = append(sigs.Usergroups, groupEntry)
	case strings.HasPrefix(groupDirName, "committee-"):
		sigs.Committees = append(sigs.Committees, groupEntry)
	default:
		return fmt.Errorf("no acceptable type in directory %q", groupDirName)
	}
	return nil
}
