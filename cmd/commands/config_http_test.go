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

package commands

import (
	"reflect"
	"testing"
	"time"

	"github.com/version-fox/vfox/internal/config"
)

func TestConfigSetHTTPTimeout(t *testing.T) {
	c := &config.Config{Plugin: config.Plugin{HTTP: &config.HTTP{}}}
	for _, key := range []string{"timeout", "downloadTimeout"} {
		keys := []string{"plugin", "http", key}
		if err := configSet(reflect.ValueOf(c), keys, "45m"); err != nil {
			t.Fatal(err)
		}
		if err := c.SaveConfig(t.TempDir()); err != nil {
			t.Fatal(err)
		}
		for _, value := range []string{"invalid", "-1s"} {
			if err := configSet(reflect.ValueOf(c), keys, value); err == nil {
				t.Fatalf("accepted %q", value)
			}
		}
	}
	if c.Plugin.HTTP.Timeout != 45*time.Minute || c.Plugin.HTTP.DownloadTimeout != 45*time.Minute {
		t.Fatalf("HTTP = %+v", c.Plugin.HTTP)
	}
	if err := configSet(reflect.ValueOf(c), []string{"plugin", "http", "timeout"}, defaultConfig([]string{"plugin", "http", "timeout"})); err != nil {
		t.Fatal(err)
	}
	request, _ := c.Plugin.HTTP.Timeouts()
	if request != 30*time.Second {
		t.Fatalf("unset timeout = %v", request)
	}
}
