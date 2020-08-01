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

package majordomo

import "github.com/pkg/errors"

// ErrNotFound is returned when a key cannot be found.
var ErrNotFound = errors.New("key not known")

// ErrURLInvalid is returned when a URL is in some way invalid.
var ErrURLInvalid = errors.New("supplied URL is invalid")

// ErrSchemeUnknown is returned when a confidant scheme is not found.
var ErrSchemeUnknown = errors.New("no confidants registered to handle that scheme")
