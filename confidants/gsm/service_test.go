// Copyright © 2020 Weald Technology Trading.
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

package gsm_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-majordomo"
	"github.com/wealdtech/go-majordomo/confidants/gsm"
	"github.com/wealdtech/go-majordomo/standard"
	"gotest.tools/assert"
)

func TestFetch(t *testing.T) {
	credentialsPath := os.Getenv("MAJORDOMO_GSM_TEST_CREDENTIALS_PATH")
	project := os.Getenv("MAJORDOMO_GSM_TEST_PROJECT")
	secretKey := os.Getenv("MAJORDOMO_GSM_TEST_SECRET_KEY")
	if credentialsPath == "" || project == "" || secretKey == "" {
		t.Skip("Environment variables MAJORDOMO_GSM_TEST_CREDENTIALS_PATH MAJORDOMO_GSM_TEST_PROJECT MAJORDOMO_GSM_TEST_SECRET_KEY required for this test")
	}

	tests := []struct {
		name  string
		key   string
		value []byte
		err   string
	}{
		{
			name: "ProjectMissing",
			key:  fmt.Sprintf("gsm:///%s", secretKey),
			err:  "no project specified",
		},
		{
			name: "KeyMissing",
			key:  fmt.Sprintf("gsm://%s", project),
			err:  "no secret specified",
		},
		{
			name: "KeyMissing2",
			key:  fmt.Sprintf("gsm://%s/", project),
			err:  "no secret specified",
		},
		{
			name:  "Good",
			key:   fmt.Sprintf("gsm://%s/%s", project, secretKey),
			value: []byte(`secret value`),
		},
		{
			name: "Unknown",
			key:  fmt.Sprintf("gsm://%s/%s2", project, secretKey),
			err:  majordomo.ErrNotFound.Error(),
		},
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := gsm.New(ctx,
		gsm.WithLogLevel(zerolog.Disabled),
		gsm.WithCredentialsPath(credentialsPath),
	)
	require.NoError(t, err)
	require.NoError(t, service.RegisterConfidant(ctx, confidant))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := service.Fetch(ctx, test.key)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.value, value)
			}
		})
	}
}

func TestOverrides(t *testing.T) {
	credentialsPath := os.Getenv("MAJORDOMO_GSM_TEST_CREDENTIALS_PATH")
	project := os.Getenv("MAJORDOMO_GSM_TEST_PROJECT")
	secretKey := os.Getenv("MAJORDOMO_GSM_TEST_SECRET_KEY")
	if credentialsPath == "" || project == "" || secretKey == "" {
		t.Skip("Environment variables MAJORDOMO_GSM_TEST_CREDENTIALS_PATH MAJORDOMO_GSM_TEST_PROJECT MAJORDOMO_GSM_TEST_SECRET_KEY required for this test")
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := gsm.New(ctx,
		gsm.WithLogLevel(zerolog.Disabled),
		gsm.WithProject(project),
		gsm.WithCredentialsPath(credentialsPath),
	)
	require.NoError(t, err)
	require.NoError(t, service.RegisterConfidant(ctx, confidant))

	// Fetch with all defaults.
	url1 := fmt.Sprintf("gsm:///%s", secretKey)
	require.NoError(t, err)
	secret1, err := service.Fetch(ctx, url1)
	require.NoError(t, err)

	// Fetch with default project.
	url2 := fmt.Sprintf("gsm:///%s", secretKey)
	require.NoError(t, err)
	secret2, err := service.Fetch(ctx, url2)
	require.NoError(t, err)

	assert.Equal(t, string(secret1), string(secret2))
}
