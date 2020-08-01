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

package asm

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/wealdtech/go-majordomo"
)

// Service returns values from Amazon secrets manager.
// This service handles URLs with the scheme "asm".
// A full URL is of the form "asm://id:secret@region/secret".
// ID and secret can be supplied at creation time if preferred.
// region can also be supplied at creation time if preferred.
// If both are supplied URLs are of the form "asm:///secret".
// Any provision of ID and secret or of region will override the defaults.
type Service struct {
	credentials *credentials.Credentials
	region      string
}

// module-wide log.
var log zerolog.Logger

// New creates a new Amazon Secrets Manager confidant.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "confidant").Str("impl", "asm").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	s := &Service{
		credentials: parameters.credentials,
		region:      parameters.region,
	}

	return s, nil
}

// SupportedURLSchemes provides the list of schemes supported by this confidant.
func (s *Service) SupportedURLSchemes(ctx context.Context) ([]string, error) {
	return []string{"asm"}, nil
}

// Fetch fetches a value given its key.
func (s *Service) Fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	if url.Host == "" {
		url.Host = s.region
	}
	if url.Host == "" {
		return nil, errors.New("no region specified")
	}

	if url.Path == "" {
		return nil, errors.New("no secret specified")
	}

	var creds *credentials.Credentials
	password, hasPassword := url.User.Password()
	switch {
	case hasPassword:
		creds = credentials.NewStaticCredentials(url.User.Username(), password, "")
	case s.credentials != nil:
		creds = s.credentials
	default:
		creds = credentials.NewEnvCredentials()
	}
	session, err := session.NewSession(aws.NewConfig().WithRegion(url.Host).WithCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "failed to initiate session with Amazon secrets manager")
	}
	svc := secretsmanager.New(session)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(strings.TrimPrefix(url.Path, "/")),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
				return nil, majordomo.ErrNotFound
			}
		}
		return nil, errors.Wrap(err, "failed to obtain secret")
	}

	if result.SecretString != nil {
		return []byte(*result.SecretString), nil
	}
	if result.SecretBinary != nil {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		size, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			return nil, errors.Wrap(err, "invalid secret binary")
		}
		return decodedBinarySecretBytes[:size], nil
	}

	// No value but no error.
	return nil, nil
}
