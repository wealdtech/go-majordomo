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

package file_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-majordomo"
	"github.com/wealdtech/go-majordomo/confidants/file"
	"github.com/wealdtech/go-majordomo/standard"
)

func TestFetch(t *testing.T) {
	// Set up some files to read.
	base, err := ioutil.TempDir("", "TestFetch")
	require.NoError(t, err)
	secretPath := filepath.Join(base, "secret-key")
	err = ioutil.WriteFile(secretPath, []byte("secret value"), 0600)
	require.NoError(t, err)

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
			name:  "Good",
			key:   fmt.Sprintf("file://%s", secretPath),
			value: []byte("secret value"),
		},
		{
			name: "Unknown",
			key:  fmt.Sprintf("file://%s2", secretPath),
			err:  majordomo.ErrNotFound.Error(),
		},
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := file.New(ctx, file.WithLogLevel(zerolog.Disabled))
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
