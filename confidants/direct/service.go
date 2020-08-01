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

package direct

import (
	"context"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
)

// Service returns returns values reflected from the key.
// This service handles URLs with the scheme "direct".
// It returns the path as the value, minus the leading "/".
// For example a URL "direct:///secret" will return "secret".
type Service struct{}

// module-wide log.
var log zerolog.Logger

// New creates a new Amazon Secrets Manager confidant.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "confidant").Str("impl", "direct").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	s := &Service{}

	return s, nil
}

// SupportedURLSchemes provides the list of schemes supported by this confidant.
func (s *Service) SupportedURLSchemes(ctx context.Context) ([]string, error) {
	return []string{"direct"}, nil
}

// Fetch fetches a value given its key URL.
func (s *Service) Fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	value := strings.TrimPrefix(url.Path, "/")
	if value == "" {
		return nil, nil
	}
	return []byte(value), nil
}
