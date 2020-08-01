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

package gsm

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/wealdtech/go-majordomo"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Service returns values from Google secrets manager.
// This service handles URLs with the scheme "gsm".
// A full URL is of the form "gsm://id:secret@project/secret".
// ID and secret can be supplied at creation time if preferred.
// region can also be supplied at creation time if preferred.
// If both are supplied URLs are of the form "gsm:///secret".
// Any provision of ID and secret or of region will override the defaults.
// N.B. the project value is the project _ID_ not the project name.
type Service struct {
	credentialsPath string
	project         string
}

// module-wide log.
var log zerolog.Logger

// New creates a new Google secrets manager confidant.
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
		credentialsPath: parameters.credentialsPath,
		project:         parameters.project,
	}

	return s, nil
}

// SupportedURLSchemes provides the list of schemes supported by this confidant.
func (s *Service) SupportedURLSchemes(ctx context.Context) ([]string, error) {
	return []string{"gsm"}, nil
}

// Fetch fetches a value given its key.
func (s *Service) Fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	if url.Host == "" {
		url.Host = s.project
	}
	if url.Host == "" {
		return nil, errors.New("no project specified")
	}

	url.Path = strings.TrimPrefix(url.Path, "/")
	if url.Path == "" {
		return nil, errors.New("no secret specified")
	}

	client, err := secretmanager.NewClient(ctx, option.WithCredentialsFile(s.credentialsPath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client connection")
	}

	path := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", url.Host, url.Path)
	log.Trace().Str("path", path).Msg("Secret path")
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}
	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "it may not exist") {
			return nil, majordomo.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to fetch secret")
	}

	return resp.Payload.Data, nil
}
