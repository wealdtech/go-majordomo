// Copyright © 2022 Weald Technology Trading.
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

package http_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-majordomo"
	httpconfidant "github.com/wealdtech/go-majordomo/confidants/http"
	"github.com/wealdtech/go-majordomo/standard"
	"github.com/wealdtech/go-majordomo/testing/logger"
)

func TestFetch(t *testing.T) {
	// Set up a simple HTTP server.
	http.HandleFunc("/path1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "response 1")
	})
	http.HandleFunc("/nodata", func(w http.ResponseWriter, r *http.Request) {
	})

	// listen to port
	go func() {
		if err := http.ListenAndServe(":5050", nil); err != nil {
			t.Fail()
		}
	}()
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name       string
		key        string
		value      []byte
		err        string
		logEntries []string
		ctxMap     map[interface{}]interface{}
	}{
		{
			name: "Nil",
			err:  majordomo.ErrNotFound.Error(),
		},
		{
			name: "BadHostname",
			key:  "http://bad_hostname/",
			err:  majordomo.ErrNotFound.Error(),
			logEntries: []string{
				`Failed to call endpoint`,
			},
		},
		{
			name: "Unknown",
			key:  "http://localhost:5050/badpath",
			err:  majordomo.ErrNotFound.Error(),
			logEntries: []string{
				`Request failed`,
			},
		},
		{
			name: "UnknownWithOptions",
			key:  "http://localhost:5050/badpath",
			ctxMap: map[interface{}]interface{}{
				&httpconfidant.HTTPMethod{}: http.MethodGet,
			},
			err: majordomo.ErrNotFound.Error(),
			logEntries: []string{
				`Request failed`,
			},
		},
		{
			name: "NoData",
			key:  "http://localhost:5050/nodata",
			err:  majordomo.ErrNotFound.Error(),
			logEntries: []string{
				`No data in response`,
			},
		},
		{
			name: "NoDataWithOptions",
			key:  "http://localhost:5050/nodata",
			ctxMap: map[interface{}]interface{}{
				&httpconfidant.HTTPMethod{}: http.MethodGet,
			},
			err: majordomo.ErrNotFound.Error(),
			logEntries: []string{
				`No data in response`,
			},
		},
		{
			name:  "Good",
			key:   "http://localhost:5050/path1",
			value: []byte("response 1"),
		},
		{
			name:  "GoodWithMethod",
			key:   "http://localhost:5050/path1",
			value: []byte("response 1"),
			ctxMap: map[interface{}]interface{}{
				&httpconfidant.HTTPMethod{}: http.MethodGet,
			},
		},
		{
			name:  "GoodWithBody",
			key:   "http://localhost:5050/path1",
			value: []byte("response 1"),
			ctxMap: map[interface{}]interface{}{
				&httpconfidant.HTTPMethod{}: http.MethodPost,
				&httpconfidant.MIMEType{}:   "text/plain",
				&httpconfidant.Body{}:       []byte("test"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			for k, v := range test.ctxMap {
				ctx = context.WithValue(ctx, k, v)
			}

			capture := logger.NewLogCapture()
			service, err := standard.New(ctx)
			require.NoError(t, err)
			confidant, err := httpconfidant.New(ctx)
			require.NoError(t, err)
			require.NoError(t, service.RegisterConfidant(ctx, confidant))

			value, err := service.Fetch(ctx, test.key)
			if test.err != "" {
				require.EqualError(t, err, test.err)
				for _, logEntry := range test.logEntries {
					if !capture.HasLog(map[string]interface{}{
						"message": logEntry,
					}) {
						t.Errorf("missing log entry %s in %v", logEntry, capture.Entries())
					}
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, test.value, value)
			}
		})
	}
}
