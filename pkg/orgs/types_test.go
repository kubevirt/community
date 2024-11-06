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

import "testing"

func TestOrg_HasMember(t *testing.T) {
	type fields struct {
		Admins  []string
		Members []string
	}
	type args struct {
		githubHandle string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "has member",
			fields: fields{
				Admins: []string{
					"admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"member1",
			},
			want: true,
		},
		{
			name: "has admin",
			fields: fields{
				Admins: []string{
					"admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"admin3",
			},
			want: true,
		},
		{
			name: "has member (lowercase)",
			fields: fields{
				Admins: []string{
					"admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"Member1",
			},
			want: true,
		},
		{
			name: "has admin (lowercase)",
			fields: fields{
				Admins: []string{
					"admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"Admin1",
			},
			want: true,
		},
		{
			name: "has member (uppercase)",
			fields: fields{
				Admins: []string{
					"admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"Member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"member1",
			},
			want: true,
		},
		{
			name: "has admin (uppercase)",
			fields: fields{
				Admins: []string{
					"Admin1",
					"admin2",
					"admin3",
				},
				Members: []string{
					"member1",
					"member2",
					"member3",
				},
			},
			args: args{
				"admin1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := Org{
				Admins:  tt.fields.Admins,
				Members: tt.fields.Members,
			}
			if got := receiver.HasMember(tt.args.githubHandle); got != tt.want {
				t.Errorf("HasMember() = %v, want %v", got, tt.want)
			}
		})
	}
}
