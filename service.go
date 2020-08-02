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

import (
	"context"
)

// Service is the interface for a majordomo.
// A majordomo takes key requests in the form of custom URLs and returns the related value.
type Service interface {
	// Fetch fetches a value given its key.
	// The key is usually defined as a URL, although it is possible that some
	// confidants may require or accept a different format.
	Fetch(ctx context.Context, key string) ([]byte, error)
}
