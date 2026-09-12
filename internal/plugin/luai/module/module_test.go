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

package module

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/version-fox/vfox/internal/config"
	lua "github.com/yuin/gopher-lua"
)

func TestPreloadIncludesFileModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Preload(L, &PreloadOptions{Config: config.DefaultConfig})
	if err := L.DoString(`
		local fs = require("fs")
		assert(type(fs.copy) == "function")
		assert(type(fs.remove) == "function")
		assert(type(fs.move) == "function")
		assert(type(fs.symlink) == "function")
	`); err != nil {
		t.Fatal(err)
	}
}

func TestConfiguredHTTPTimeoutsReachLua(t *testing.T) {
	for _, proxy := range []bool{false, true} {
		t.Run(fmt.Sprintf("proxy=%t", proxy), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Send headers first so the download must use its configured body deadline.
				w.Header().Set("Content-Length", "100")
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				// Bound the server independently so a regression fails instead of hanging.
				select {
				case <-r.Context().Done():
				case <-time.After(2 * time.Second):
				}
			}))
			defer server.Close()
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			data := fmt.Sprintf("plugin:\n  http:\n    timeout: 25ms\n    downloadTimeout: 50ms\nproxy:\n  enable: %t\n  url: %s\n", proxy, server.URL)
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			c, err := config.NewConfigWithPath(path)
			if err != nil {
				t.Fatal(err)
			}
			L := lua.NewState()
			defer L.Close()
			Preload(L, &PreloadOptions{Config: c})
			url := server.URL
			if proxy {
				url = "http://vfox-timeout.invalid/manifest"
			}
			L.SetGlobal("url", lua.LString(url))
			L.SetGlobal("destination", lua.LString(filepath.Join(dir, "download")))
			if err := L.DoString(`
    local http = require("http")
    local resp, err = http.get({url = url})
    assert(resp == nil)
    assert(string.find(err, "deadline exceeded") or string.find(err, "timeout"), err)
    err = http.download_file({url = url}, destination)
    assert(type(err) == "string")
    assert(string.find(err, "deadline exceeded") or string.find(err, "timeout"), err)
   `); err != nil {
				t.Fatal(err)
			}
		})
	}
}
