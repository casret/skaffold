/*
Copyright 2020 The Skaffold Authors

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

package build

import (
	"context"
	"errors"
	"io"
	"testing"

	latest_v1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestNewBuilderMux(t *testing.T) {
	tests := []struct {
		description         string
		pipelines           []latest_v1.Pipeline
		pipeBuilder         func(latest_v1.Pipeline) (PipelineBuilder, error)
		shouldErr           bool
		expectedBuilders    []string
		expectedConcurrency int
	}{
		{
			description: "only local builder",
			pipelines: []latest_v1.Pipeline{
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{LocalBuild: &latest_v1.LocalBuild{Concurrency: util.IntPtr(1)}}}},
			},
			pipeBuilder:         newMockPipelineBuilder,
			expectedBuilders:    []string{"local"},
			expectedConcurrency: 1,
		},
		{
			description: "only cluster builder",
			pipelines: []latest_v1.Pipeline{
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{Cluster: &latest_v1.ClusterDetails{}}}},
			},
			pipeBuilder:      newMockPipelineBuilder,
			expectedBuilders: []string{"cluster"},
		},
		{
			description: "only gcb builder",
			pipelines: []latest_v1.Pipeline{
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{GoogleCloudBuild: &latest_v1.GoogleCloudBuild{}}}},
			},
			pipeBuilder:      newMockPipelineBuilder,
			expectedBuilders: []string{"gcb"},
		},
		{
			description: "min non-zero concurrency",
			pipelines: []latest_v1.Pipeline{
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{LocalBuild: &latest_v1.LocalBuild{Concurrency: util.IntPtr(0)}}}},
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{LocalBuild: &latest_v1.LocalBuild{Concurrency: util.IntPtr(3)}}}},
				{Build: latest_v1.BuildConfig{BuildType: latest_v1.BuildType{Cluster: &latest_v1.ClusterDetails{Concurrency: 2}}}},
			},
			pipeBuilder:         newMockPipelineBuilder,
			expectedBuilders:    []string{"local", "local", "cluster"},
			expectedConcurrency: 2,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			cfg := &mockConfig{pipelines: test.pipelines}

			b, err := NewBuilderMux(cfg, nil, test.pipeBuilder)
			t.CheckError(test.shouldErr, err)
			if test.shouldErr {
				return
			}
			t.CheckTrue(len(b.builders) == len(test.expectedBuilders))
			for i := range b.builders {
				t.CheckDeepEqual(test.expectedBuilders[i], b.builders[i].(*mockPipelineBuilder).builderType)
			}
			t.CheckDeepEqual(test.expectedConcurrency, b.concurrency)
		})
	}
}

type mockConfig struct {
	pipelines []latest_v1.Pipeline
	optRepo   string
}

func (m *mockConfig) GetPipelines() []latest_v1.Pipeline { return m.pipelines }
func (m *mockConfig) GlobalConfig() string               { return "" }
func (m *mockConfig) DefaultRepo() *string {
	if m.optRepo != "" {
		return &m.optRepo
	}
	return nil
}
func (m *mockConfig) BuildConcurrency() int { return -1 }

type mockPipelineBuilder struct {
	concurrency int
	builderType string
}

func (m *mockPipelineBuilder) PreBuild(ctx context.Context, out io.Writer) error { return nil }

func (m *mockPipelineBuilder) Build(ctx context.Context, out io.Writer, artifact *latest_v1.Artifact) ArtifactBuilder {
	return nil
}

func (m *mockPipelineBuilder) PostBuild(ctx context.Context, out io.Writer) error { return nil }

func (m *mockPipelineBuilder) Concurrency() int { return m.concurrency }

func (m *mockPipelineBuilder) Prune(context.Context, io.Writer) error { return nil }

func newMockPipelineBuilder(p latest_v1.Pipeline) (PipelineBuilder, error) {
	switch {
	case p.Build.BuildType.LocalBuild != nil:
		c := 0
		if p.Build.LocalBuild.Concurrency != nil {
			c = *p.Build.LocalBuild.Concurrency
		}
		return &mockPipelineBuilder{builderType: "local", concurrency: c}, nil
	case p.Build.BuildType.Cluster != nil:
		return &mockPipelineBuilder{builderType: "cluster", concurrency: p.Build.Cluster.Concurrency}, nil
	case p.Build.BuildType.GoogleCloudBuild != nil:
		return &mockPipelineBuilder{builderType: "gcb", concurrency: p.Build.GoogleCloudBuild.Concurrency}, nil
	default:
		return nil, errors.New("invalid config")
	}
}
