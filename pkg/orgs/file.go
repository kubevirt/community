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

package orgs

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func ReadFile(path string) (*Orgs, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", path, err)
	}

	orgs := &Orgs{}
	err = yaml.Unmarshal(buf, orgs)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", path, err)
	}
	return orgs, err
}
