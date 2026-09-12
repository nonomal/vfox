/*
 *    Copyright 2026 Han Li and contributors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package config

import (
	"fmt"
	"time"
)

// Plugin contains settings for Lua plugins.
type Plugin struct {
	HTTP *HTTP `yaml:"http"`
}

// HTTP configures the Lua HTTP module. Zero inherits the shared setting, or
// uses the built-in default when no shared setting is provided.
type HTTP struct {
	Timeout         time.Duration `yaml:"timeout"`
	DownloadTimeout time.Duration `yaml:"downloadTimeout"`
}

func (h *HTTP) Validate() error {
	if h == nil {
		return nil
	}
	if h.Timeout < 0 {
		return fmt.Errorf("plugin.http.timeout must not be negative")
	}
	if h.DownloadTimeout < 0 {
		return fmt.Errorf("plugin.http.downloadTimeout must not be negative")
	}
	return nil
}

// Timeouts resolves defaults without changing the stored configuration, so
// omitted user fields can still inherit values from the shared configuration.
func (h *HTTP) Timeouts() (request, download time.Duration) {
	request, download = 30*time.Second, 30*time.Minute
	if h != nil {
		if h.Timeout > 0 {
			request = h.Timeout
		}
		if h.DownloadTimeout > 0 {
			download = h.DownloadTimeout
		}
	}
	return
}

func mergeHTTP(shared, user *HTTP) *HTTP {
	result := &HTTP{}
	if shared != nil {
		*result = *shared
	}
	if user != nil {
		if user.Timeout != 0 {
			result.Timeout = user.Timeout
		}
		if user.DownloadTimeout != 0 {
			result.DownloadTimeout = user.DownloadTimeout
		}
	}
	return result
}
