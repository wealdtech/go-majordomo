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
	"github.com/rs/zerolog"
)

type parameters struct {
	logLevel   zerolog.Level
	clientCert []byte
	clientKey  []byte
	caCert     []byte
}

// Parameter is the interface for service parameters.
type Parameter interface {
	apply(*parameters)
}

type parameterFunc func(*parameters)

func (f parameterFunc) apply(p *parameters) {
	f(p)
}

// WithLogLevel sets the log level for the module.
func WithLogLevel(logLevel zerolog.Level) Parameter {
	return parameterFunc(func(p *parameters) {
		p.logLevel = logLevel
	})
}

// WithClientCert sets the client certificate for HTTPS requests.
func WithClientCert(cert []byte) Parameter {
	return parameterFunc(func(p *parameters) {
		p.clientCert = cert
	})
}

// WithClientKey sets the client Key for HTTPS requests.
func WithClientKey(key []byte) Parameter {
	return parameterFunc(func(p *parameters) {
		p.clientKey = key
	})
}

// WithCACert sets the certificate authority certificate for HTTPS requests.
func WithCACert(cert []byte) Parameter {
	return parameterFunc(func(p *parameters) {
		p.caCert = cert
	})
}

// parseAndCheckParameters parses and checks parameters to ensure that mandatory parameters are present and correct.
func parseAndCheckParameters(params ...Parameter) (*parameters, error) {
	parameters := parameters{
		logLevel: zerolog.GlobalLevel(),
	}
	for _, p := range params {
		if params != nil {
			p.apply(&parameters)
		}
	}

	return &parameters, nil
}
