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

package standard

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	majordomo "github.com/wealdtech/go-majordomo"
)

// Service is the standard majordomo service.
type Service struct {
	confidants map[string]majordomo.Confidant
}

// module-wide log.
var log zerolog.Logger

// New creates a new majordomo instance.
// Confidants must be added to the instance with
// `RegisterConfidant()`
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "confidant").Str("impl", "gsm").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	s := &Service{
		confidants: make(map[string]majordomo.Confidant),
	}

	return s, nil
}

// RegisterConfidant registers a confidant.
// The confidant will register whichever URL schemes it supports.
func (s *Service) RegisterConfidant(ctx context.Context, confidant majordomo.Confidant) error {
	schemes, err := confidant.SupportedURLSchemes(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to obtain supported URL schemes from confidant")
	}
	for _, scheme := range schemes {
		if _, exists := s.confidants[scheme]; exists {
			return fmt.Errorf("scheme %s already registered by another confidant", scheme)
		}
		s.confidants[scheme] = confidant
	}
	return nil
}

// Fetch fetches a URL from a confidant.
func (s *Service) Fetch(ctx context.Context, req string) ([]byte, error) {
	// We short-circuit anything that isn't a URL as a direct value.
	if req == "" {
		// Empty req is never found.
		return nil, majordomo.ErrNotFound
	}
	if !strings.Contains(req, "://") {
		return []byte(req), nil
	}

	url, err := url.Parse(req)
	if err != nil {
		return nil, majordomo.ErrURLInvalid
	}
	if url.Scheme == "" {
		return nil, majordomo.ErrURLInvalid
	}

	confidant, exists := s.confidants[url.Scheme]
	if !exists {
		return nil, majordomo.ErrSchemeUnknown
	}

	val, err := confidant.Fetch(ctx, url)
	if err != nil {
		// We return this error without wrapping it to allow comparison to majordomo well-known errors.
		return nil, err
	}
	return val, nil
}
