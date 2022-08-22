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

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/wealdtech/go-majordomo"
)

// Service returns the values from an HTTP connection.
// This service handles URLs with the scheme "http" and "https".
// It returns the file at the URL as the value.
// For example a URL "http://www.example.com/secret.txt" will return the contents
// of the file "secret.txt" on the server "www.example.com"
// Additional information, such as certificates, can be passed as context values.  The available values are:
// - CACert a certificate authority certificate, as a byte slice
// - ClientCert a client certificate, as a byte slice
// - ClientKey a client key, as a byte slice
// - HTTPMethod the HTTP method, as a string (e.g. http.MethodPost)
// - MIMEType the MIME type for request and response, as a string (e.g. application/json)
// - Body the request body, as a byte slice
type Service struct{}

// CaCert is a context tag for the CA certificate.
type CACert struct{}

// ClientCert is a context tag for the client certificate.
type ClientCert struct{}

// ClientKey is a context tag for the client key.
type ClientKey struct{}

// HTTPMethod is a context tag for the HTTP method.
type HTTPMethod struct{}

// MIMEType is a context tag for the MIME type.
type MIMEType struct{}

// Body is a context tag for the request body.
type Body struct{}

// module-wide log.
var log zerolog.Logger

// New creates a new file confidant.
func New(ctx context.Context, params ...Parameter) (*Service, error) {
	parameters, err := parseAndCheckParameters(params...)
	if err != nil {
		return nil, errors.Wrap(err, "problem with parameters")
	}

	// Set logging.
	log = zerologger.With().Str("service", "confidant").Str("impl", "http").Logger()
	if parameters.logLevel != log.GetLevel() {
		log = log.Level(parameters.logLevel)
	}

	s := &Service{}

	return s, nil
}

// SupportedURLSchemes provides the list of schemes supported by this confidant.
func (s *Service) SupportedURLSchemes(ctx context.Context) ([]string, error) {
	return []string{"http", "https"}, nil
}

// Fetch fetches a value given its https URL.
func (s *Service) Fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	_, clientCertExists := ctx.Value(&ClientCert{}).([]byte)
	_, httpMethodExists := ctx.Value(&HTTPMethod{}).(string)
	_, mimeTypeExists := ctx.Value(&MIMEType{}).(string)
	_, bodyExists := ctx.Value(&Body{}).([]byte)
	if clientCertExists || httpMethodExists || mimeTypeExists || bodyExists {
		return s.fetchWithOptions(ctx, url)
	}
	return s.fetch(ctx, url)
}

func (s *Service) fetch(ctx context.Context, url *url.URL) ([]byte, error) {
	resp, err := http.Get(url.String())
	if err != nil {
		log.Debug().Err(err).Msg("Failed to call endpoint")
		return nil, majordomo.ErrNotFound
	}
	if resp == nil {
		log.Debug().Err(err).Msg("No body returned for endpoint")
		return nil, majordomo.ErrNotFound
	}

	data, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		log.Debug().Err(closeErr).Msg("Response close() returned an error")
	}
	if err != nil {
		log.Debug().Err(err).Msg("Failed to read response")
		return nil, majordomo.ErrNotFound
	}
	if len(data) == 0 {
		log.Debug().Err(err).Msg("No data in response")
		return nil, majordomo.ErrNotFound
	}

	statusFamily := resp.StatusCode / 100
	if statusFamily != 2 {
		log.Debug().Int("status_code", resp.StatusCode).Str("data", string(data)).Msg("Request failed")
		return nil, majordomo.ErrNotFound
	}

	return data, nil
}

func (s *Service) fetchWithOptions(ctx context.Context, url *url.URL) ([]byte, error) {
	caCert, caCertExists := ctx.Value(&CACert{}).([]byte)
	clientCert, clientCertExists := ctx.Value(&ClientCert{}).([]byte)
	clientKey, clientKeyExists := ctx.Value(&ClientKey{}).([]byte)
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	if caCertExists {
		log.Trace().Msg("Adding CA certificate")
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	if clientCertExists && !clientKeyExists {
		return nil, errors.New("both or neither of client certificate and client key must be specified")
	}
	if clientCertExists {
		log.Trace().Msg("Adding client certificate")
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, errors.New("invalid client certificate or key")
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	httpMethod, httpMethodExists := ctx.Value(&HTTPMethod{}).(string)
	if !httpMethodExists {
		httpMethod = http.MethodGet
	} else {
		httpMethod = strings.ToUpper(httpMethod)
	}

	body, bodyExists := ctx.Value(&Body{}).([]byte)
	var bodyReader io.Reader
	if bodyExists {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, url.String(), bodyReader)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create request")
		return nil, majordomo.ErrNotFound
	}

	mimeType, mimeTypeExists := ctx.Value(&MIMEType{}).(string)
	if mimeTypeExists {
		req.Header.Set("Content-type", strings.ToLower(mimeType))
		req.Header.Set("Accept", strings.ToLower(mimeType))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Debug().Msg("Failed to call endpoint")
		return nil, majordomo.ErrNotFound
	}
	if resp == nil {
		log.Debug().Msg("No body returned for endpoint")
		return nil, majordomo.ErrNotFound
	}

	data, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		log.Debug().Err(closeErr).Msg("Response close() returned an error")
	}
	if err != nil {
		log.Debug().Err(err).Msg("Failed to read response")
		return nil, majordomo.ErrNotFound
	}
	if len(data) == 0 {
		log.Debug().Err(err).Msg("No data in response")
		return nil, majordomo.ErrNotFound
	}
	// Because we are using our own client for this call we close it here to avoid connection leaks.
	client.CloseIdleConnections()

	statusFamily := resp.StatusCode / 100
	if statusFamily != 2 {
		log.Debug().Int("status_code", resp.StatusCode).Str("data", string(data)).Msg("Request failed")
		return nil, majordomo.ErrNotFound
	}

	return data, nil
}
