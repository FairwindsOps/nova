// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_removeDuplicateString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		want  []string
	}{
		{
			name: "test1",
			slice: []string{
				"a",
				"b",
				"c",
				"a",
			},
			want: []string{
				"a",
				"b",
				"c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDuplicateString(tt.slice)
			if !assert.EqualValues(t, got, tt.want) {
				t.Errorf("removeDuplicateString() got: %v, want %v", got, tt.want)
			}
		})
	}
}
