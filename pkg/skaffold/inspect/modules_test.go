/*
Copyright 2021 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inspect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/parser"
	sErrors "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/errors"
	v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestPrintModulesList(t *testing.T) {
	tests := []struct {
		description string
		configSet   parser.SkaffoldConfigSet
		err         error
		expected    string
	}{
		{
			description: "print modules",
			configSet: parser.SkaffoldConfigSet{
				&parser.SkaffoldConfigEntry{SkaffoldConfig: &v1.SkaffoldConfig{Metadata: v1.Metadata{Name: "cfg1"}}, SourceFile: "path/to/cfg1"},
				&parser.SkaffoldConfigEntry{SkaffoldConfig: &v1.SkaffoldConfig{Metadata: v1.Metadata{Name: "cfg2"}}, SourceFile: "path/to/cfg2"},
			},
			expected: `{"modules":[{"name":"cfg1","path":"path/to/cfg1"},{"name":"cfg2","path":"path/to/cfg2"}]}` + "\n",
		},
		{
			description: "actionable error",
			err:         sErrors.MainConfigFileNotFoundErr("path/to/skaffold.yaml", fmt.Errorf("failed to read file : %q", "skaffold.yaml")),
			expected:    `{"errorCode":"CONFIG_FILE_NOT_FOUND_ERR","errorMessage":"unable to find configuration file \"path/to/skaffold.yaml\": failed to read file : \"skaffold.yaml\". Check that the specified configuration file exists at \"path/to/skaffold.yaml\"."}` + "\n",
		},
		{
			description: "generic error",
			err:         errors.New("some error occurred"),
			expected:    `{"errorCode":"UNKNOWN_ERROR","errorMessage":"some error occurred"}` + "\n",
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.Override(&getConfigSetFunc, func(config.SkaffoldOptions) (parser.SkaffoldConfigSet, error) {
				return test.configSet, test.err
			})
			var buf bytes.Buffer
			err := PrintModulesList(context.Background(), &buf, Options{OutFormat: "json"})
			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected, buf.String())
		})
	}
}
