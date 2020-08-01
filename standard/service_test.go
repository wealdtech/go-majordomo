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

package standard_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	majordomo "github.com/wealdtech/go-majordomo"
	"github.com/wealdtech/go-majordomo/standard"
)

func TestFetch(t *testing.T) {
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
			name:  "Direct",
			key:   "foo",
			value: []byte("foo"),
		},
	}

	ctx := context.Background()
	service, err := standard.New(ctx, standard.WithLogLevel(zerolog.Disabled))
	require.NoError(t, err)
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

func TestRegister(t *testing.T) {
	ctx := context.Background()
	service, err := standard.New(ctx, standard.WithLogLevel(zerolog.Disabled))
	require.NoError(t, err)
	confidant := &MockConfidant{}
	require.NoError(t, service.RegisterConfidant(ctx, confidant))

	// Cannot re-register.
	require.Error(t, service.RegisterConfidant(ctx, confidant))

	// Scheme supported.
	value, err := service.Fetch(ctx, "mock://")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), value)

	// Scheme not supported.
	_, err = service.Fetch(ctx, "nomock://")
	require.EqualError(t, err, majordomo.ErrSchemeUnknown.Error())
}

// MockConfidant is a mock implementation of confidant.
type MockConfidant struct{}

func (s *MockConfidant) SupportedURLSchemes(ctx context.Context) ([]string, error) {
	return []string{"mock"}, nil
}

// Fetch says hello.
func (s *MockConfidant) Fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	return []byte("hello"), nil
}
