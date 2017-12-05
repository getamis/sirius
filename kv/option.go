// Copyright 2017 AMIS Technologies
//
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

package kv

import (
	"encoding/json"
	"time"
)

type putOptions struct {
	IsPrefix bool          // Optional, indicate if the given key is a prefix
	TTL      time.Duration // Optional, expiration time associated with the key
}

type PutOption func(*putOptions)

func IsPrefix(isPrefix bool) PutOption {
	return func(o *putOptions) {
		o.IsPrefix = isPrefix
	}
}

func PutExpiration(t time.Duration) PutOption {
	return func(o *putOptions) {
		o.TTL = t
	}
}

// ----------------------------------------------------------------------------

type lockOptions struct {
	Value     []byte        // Optional, value to associate with the lock
	TTL       time.Duration // Optional, expiration time associated with the lock
	RenewLock chan struct{} // Optional, chan used to control and stop the session ttl renewal for the lock
}

type LockOption func(*lockOptions)

func LockExpiration(t time.Duration) LockOption {
	return func(o *lockOptions) {
		o.TTL = t
	}
}

func LockValue(value interface{}) LockOption {
	return func(o *lockOptions) {
		o.Value, _ = json.Marshal(value)
	}
}
