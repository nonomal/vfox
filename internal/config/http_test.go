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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHTTPConfig(t *testing.T) {
	for _, tc := range []struct {
		name, yaml        string
		request, download time.Duration
		invalid           bool
	}{
		{"old config", "proxy: {}", 30 * time.Second, 30 * time.Minute, false},
		{"custom", "plugin:\n  http:\n    timeout: 60s\n    downloadTimeout: 45m", time.Minute, 45 * time.Minute, false},
		{"partial", "plugin:\n  http:\n    timeout: 5s", 5 * time.Second, 30 * time.Minute, false},
		{"zero defaults", "plugin:\n  http:\n    timeout: 0s", 30 * time.Second, 30 * time.Minute, false},
		{"negative", "plugin:\n  http:\n    timeout: -1s", 0, 0, true},
		{"negative download", "plugin:\n  http:\n    downloadTimeout: -1s", 0, 0, true},
		{"invalid", "plugin:\n  http:\n    timeout: invalid", 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0600); err != nil {
				t.Fatal(err)
			}
			c, err := NewConfigWithPath(path)
			if tc.invalid {
				if err == nil {
					t.Fatal("expected invalid timeout error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			check := func(c *Config) {
				t.Helper()
				request, download := c.Plugin.HTTP.Timeouts()
				if request != tc.request || download != tc.download {
					t.Fatalf("timeouts = %v, %v", request, download)
				}
			}
			check(c)
			if err := c.SaveConfig(dir); err != nil {
				t.Fatal(err)
			}
			c, err = NewConfigWithPath(path)
			if err != nil {
				t.Fatal(err)
			}
			check(c)
		})
	}
}

func TestHTTPMerge(t *testing.T) {
	shared := &Config{Plugin: Plugin{HTTP: &HTTP{Timeout: time.Minute, DownloadTimeout: time.Hour}}}
	for _, tc := range []struct {
		name              string
		user              *Config
		request, download time.Duration
	}{
		{"missing", &Config{}, time.Minute, time.Hour},
		{"empty", &Config{Plugin: Plugin{HTTP: &HTTP{}}}, time.Minute, time.Hour},
		{"partial", &Config{Plugin: Plugin{HTTP: &HTTP{Timeout: 5 * time.Second}}}, 5 * time.Second, time.Hour},
		{"explicit default", &Config{Plugin: Plugin{HTTP: &HTTP{Timeout: 30 * time.Second}}}, 30 * time.Second, time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Merge(shared, tc.user)
			request, download := got.Plugin.HTTP.Timeouts()
			if request != tc.request || download != tc.download {
				t.Fatalf("timeouts = %v, %v", request, download)
			}
		})
	}
	if shared.Plugin.HTTP.Timeout != time.Minute {
		t.Fatal("merge mutated shared config")
	}
}
