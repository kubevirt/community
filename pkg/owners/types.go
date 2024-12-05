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

package owners

type Owners struct {
	Reviewers         []string            `yaml:"reviewers"`
	Approvers         []string            `yaml:"approvers"`
	EmeritusApprovers []string            `yaml:"emeritus_approvers"`
	Filters           map[string][]string `yaml:"filters"`
}

type OwnersAliases struct {
	Aliases map[string][]string `yaml:"aliases"`
}

func (a OwnersAliases) Resolve(aliases []string) []string {
	var resolved []string
	for _, alias := range aliases {
		resolvedUserNames, exists := a.Aliases[alias]
		if !exists {
			resolved = append(resolved, alias)
			continue
		}
		resolved = append(resolved, resolvedUserNames...)
	}
	return resolved
}
