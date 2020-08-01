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

package direct_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-majordomo"
	"github.com/wealdtech/go-majordomo/confidants/direct"
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
		{
			name: "SchemeUnknown",
			key:  "unknown://secret",
			err:  majordomo.ErrSchemeUnknown.Error(),
		},
		{
			name: "PathEmpty",
			key:  "direct://",
		},
		{
			name:  "DirectSimple",
			key:   "direct:///secret",
			value: []byte("secret"),
		},
		{
			name:  "DirectWithSlash",
			key:   "direct:///dev/secret",
			value: []byte("dev/secret"),
		},
		{
			name:  "DirectWithLeadingSlash",
			key:   "direct:////secret",
			value: []byte("/secret"),
		},
	}

	ctx := context.Background()
	service, err := standard.New(ctx)
	require.NoError(t, err)
	confidant, err := direct.New(ctx, direct.WithLogLevel(zerolog.Disabled))
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
