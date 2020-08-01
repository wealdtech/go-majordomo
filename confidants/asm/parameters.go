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
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rs/zerolog"
)

type parameters struct {
	logLevel    zerolog.Level
	credentials *credentials.Credentials
	region      string
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

// WithCredentials sets the default credentials for accessing Amazon Secrets Manager.
func WithCredentials(credentials *credentials.Credentials) Parameter {
	return parameterFunc(func(p *parameters) {
		p.credentials = credentials
	})
}

// WithRegion sets the default region for accessing Amazon Secrets Manager.
func WithRegion(region string) Parameter {
	return parameterFunc(func(p *parameters) {
		p.region = region
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
