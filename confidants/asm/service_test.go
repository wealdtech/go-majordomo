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

package asm_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-majordomo"
	"github.com/wealdtech/go-majordomo/confidants/asm"
	"github.com/wealdtech/go-majordomo/standard"
	"gotest.tools/assert"
)

func TestFetch(t *testing.T) {
	id := os.Getenv("MAJORDOMO_ASM_TEST_ID")
	secret := os.Getenv("MAJORDOMO_ASM_TEST_SECRET")
	region := os.Getenv("MAJORDOMO_ASM_TEST_REGION")
	secretKey := os.Getenv("MAJORDOMO_ASM_TEST_SECRET_KEY")
	if id == "" || secret == "" || region == "" || secretKey == "" {
		t.Skip("Environment variables MAJORDOMO_ASM_TEST_ID MAJORDOMO_ASM_TEST_SECRET MAJORDOMO_ASM_TEST_REGION MAJORDOMO_ASM_TEST_SECRET_KEY required for this test")
	}

	tests := []struct {
		name  string
		key   string
		value []byte
		err   string
	}{
		{
			name: "Nil",
			err:  majordomo.ErrNotFound.Error(),
		},
		{
			name: "RegionMissing",
			key:  fmt.Sprintf("asm://%s:%s//%s", id, secret, secretKey),
			err:  majordomo.ErrURLInvalid.Error(),
		},
		{
			name: "IDMissing",
			key:  fmt.Sprintf("asm://:%s@%s/%s", secret, region, secretKey),
			err:  "failed to obtain secret: EmptyStaticCreds: static credentials are empty",
		},
		{
			name: "SecretMissing",
			key:  fmt.Sprintf("asm://%s@%s/%s", id, region, secretKey),
			err:  "failed to obtain secret: EnvAccessKeyNotFound: AWS_ACCESS_KEY_ID or AWS_ACCESS_KEY not found in environment",
		},
		{
			name: "PathMissing",
			key:  fmt.Sprintf("asm://%s:%s@%s", id, secret, region),
			err:  "no secret specified",
		},
		{
			name:  "Good",
			key:   fmt.Sprintf("asm://%s:%s@%s/%s", id, secret, region, secretKey),
			value: []byte(`secret value`),
		},
		{
			name: "Unknown",
			key:  fmt.Sprintf("asm://%s:%s@%s/%s2", id, secret, region, secretKey),
			err:  majordomo.ErrNotFound.Error(),
		},
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := asm.New(ctx, asm.WithLogLevel(zerolog.Disabled))
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
	id := os.Getenv("MAJORDOMO_ASM_TEST_ID")
	secret := os.Getenv("MAJORDOMO_ASM_TEST_SECRET")
	region := os.Getenv("MAJORDOMO_ASM_TEST_REGION")
	secretKey := os.Getenv("MAJORDOMO_ASM_TEST_SECRET_KEY")
	if id == "" || secret == "" || region == "" || secretKey == "" {
		t.Skip("Environment variables MAJORDOMO_ASM_TEST_ID MAJORDOMO_ASM_TEST_SECRET MAJORDOMO_ASM_TEST_REGION MAJORDOMO_ASM_TEST_SECRET_KEY required for this test")
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := asm.New(ctx,
		asm.WithLogLevel(zerolog.Disabled),
		asm.WithRegion(region),
		asm.WithCredentials(credentials.NewStaticCredentials(id, secret, "")),
	)
	require.NoError(t, err)
	require.NoError(t, service.RegisterConfidant(ctx, confidant))

	// Fetch with all defaults.
	url1 := fmt.Sprintf("asm:///%s", secretKey)
	require.NoError(t, err)
	secret1, err := service.Fetch(ctx, url1)
	require.NoError(t, err)

	// Fetch with default region.
	url2 := fmt.Sprintf("asm://%s:%s@/%s", id, secret, secretKey)
	require.NoError(t, err)
	secret2, err := service.Fetch(ctx, url2)
	require.NoError(t, err)

	// Fetch with default credentials.
	url3 := fmt.Sprintf("asm://%s/%s", region, secretKey)
	require.NoError(t, err)
	secret3, err := service.Fetch(ctx, url3)
	require.NoError(t, err)

	// Fetch with no defaults.
	url4 := fmt.Sprintf("asm://%s:%s@%s/%s", id, secret, region, secretKey)
	require.NoError(t, err)
	secret4, err := service.Fetch(ctx, url4)
	require.NoError(t, err)

	assert.Equal(t, string(secret1), string(secret2))
	assert.Equal(t, string(secret2), string(secret3))
	assert.Equal(t, string(secret3), string(secret4))
}
