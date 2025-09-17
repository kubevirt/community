#!/usr/bin/env bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright the KubeVirt Authors.
#
#

set -e
set -u
set -o pipefail

if ! command -V covreport; then
    echo "covreport binary required to be installed locally"
    echo "run > GOBIN=${LOCAL_BIN} go install github.com/cancue/covreport@latest"
    exit 1
fi

go test ./... -coverprofile=/tmp/coverage.out
covreport -i /tmp/coverage.out -o "${COVERAGE_OUTPUT_PATH}"
echo "coverage written to ${COVERAGE_OUTPUT_PATH}"
