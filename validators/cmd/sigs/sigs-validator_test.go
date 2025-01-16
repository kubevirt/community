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
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"kubevirt.io/community/pkg/labels"
	"kubevirt.io/community/pkg/orgs"
	"kubevirt.io/community/pkg/sigs"
	"reflect"
	"testing"
)

func TestValidateGroups(t *testing.T) {
	type args struct {
		sigsYAML         *sigs.Sigs
		expectedSigsYAML *sigs.Sigs
		labelsYAML       *labels.LabelsYAML
		kubevirtOrg      orgs.Org
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "sig: removes directory",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Dir:  "non-existing-dir",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
						},
					},
				},
			},
		},
		{
			name: "sig: leaves directory",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Dir:  "testdata/existing-dir",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Dir:  "testdata/existing-dir",
						},
					},
				},
			},
		},
		{
			name: "sig: removes directory if not a dir",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Dir:  "testdata/existing-file",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
						},
					},
				},
			},
		},
		{
			name: "sig: removes chair",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leadership: &sigs.Leadership{
								Chairs: []*sigs.Chair{
									{Github: "nonexisting-gh"},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name:       "sig-test",
							Leadership: &sigs.Leadership{},
						},
					},
				},
			},
		},
		{
			name: "sig: leaves chair",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leadership: &sigs.Leadership{
								Chairs: []*sigs.Chair{
									{Github: "existing-gh"},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leadership: &sigs.Leadership{
								Chairs: []*sigs.Chair{
									{Github: "existing-gh"},
								},
							},
						},
					},
				},
				kubevirtOrg: orgs.Org{
					Members: []string{
						"existing-gh",
					},
				},
			},
		},
		{
			name: "sig: removes lead",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leads: []*sigs.Lead{
								{
									Github: "nonexisting-gh",
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
						},
					},
				},
			},
		},
		{
			name: "sig: leaves lead",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leads: []*sigs.Lead{
								{
									Github: "existing-gh",
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							Leads: []*sigs.Lead{
								{
									Github: "existing-gh",
								},
							},
						},
					},
				},
				kubevirtOrg: orgs.Org{
					Members: []string{
						"existing-gh",
					},
				},
			},
		},
		{
			name: "sig: removes label",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name:  "sig-test",
							Label: "nonexisting-label",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
						},
					},
				},
				labelsYAML: &labels.LabelsYAML{
					Default: &labels.Repo{Labels: []*labels.Label{}},
				},
			},
		},
		{
			name: "sig: leaves default label",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name:  "sig-test",
							Label: "existing-label",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name:  "sig-test",
							Label: "existing-label",
						},
					},
				},
				labelsYAML: &labels.LabelsYAML{
					Default: &labels.Repo{Labels: []*labels.Label{{
						Name: "existing-label",
					}}},
				},
			},
		},
		{
			name: "sig: removes repo-specific label",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name:  "sig-test",
							Label: "existing-label",
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
						},
					},
				},
				labelsYAML: &labels.LabelsYAML{
					Default: &labels.Repo{Labels: []*labels.Label{{}}},
					Repos: map[string]*labels.Repo{
						"somerepo": {Labels: []*labels.Label{
							{
								Name: "existing-label",
							},
						}},
					},
				},
			},
		},
		{
			name: "sig: subproject - leaves existing owners references",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Owners: []string{
										"https://raw.githubusercontent.com/kubevirt/community/main/OWNERS",
									},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Owners: []string{
										"https://raw.githubusercontent.com/kubevirt/community/main/OWNERS",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "sig: subproject - remove owners reference if not found",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Owners: []string{
										"https://raw.githubusercontent.com/kubevirt/non-existing-repo/main/OWNERS",
									},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "sig: subproject - leaves lead",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Leads: []*sigs.Lead{
										{
											Github: "existing-gh",
										},
									},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Leads: []*sigs.Lead{
										{
											Github: "existing-gh",
										},
									},
								},
							},
						},
					},
				},
				kubevirtOrg: orgs.Org{
					Members: []string{
						"existing-gh",
					},
				},
			},
		},
		{
			name: "sig: subproject - removes lead if not org member",
			args: args{
				sigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
									Leads: []*sigs.Lead{
										{
											Github: "existing-gh",
										},
									},
								},
							},
						},
					},
				},
				expectedSigsYAML: &sigs.Sigs{
					Sigs: []*sigs.Group{
						{
							Name: "sig-test",
							SubProjects: []*sigs.SubProject{
								{
									Name: "some-subproject",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateGroups(tt.args.sigsYAML, tt.args.labelsYAML, tt.args.kubevirtOrg)
			if !reflect.DeepEqual(tt.args.expectedSigsYAML, tt.args.sigsYAML) {
				t.Errorf("sigs yaml:\n\ngot: %v\n\nwant: %v", printSigsYAML(tt.args.sigsYAML), printSigsYAML(tt.args.expectedSigsYAML))
			}
		})
	}
}

func printSigsYAML(s *sigs.Sigs) string {
	out, err := yaml.Marshal(s)
	if err != nil {
		log.Fatalf("failed to print sigs yaml: %v", err)
	}
	return string(out)
}
