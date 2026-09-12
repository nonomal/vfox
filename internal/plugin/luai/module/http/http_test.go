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

package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/version-fox/vfox/internal/plugin/luai/codec"
	"github.com/version-fox/vfox/internal/shared/util"

	"github.com/version-fox/vfox/internal/config"
	lua "github.com/yuin/gopher-lua"
)

const jsonUrl = `https://version-fox.github.io/vfox-plugins/index.json`

func TestWithConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip on windows, the proxy won't error on windows.")
	}

	const str = `
	local http = require("http")
	assert(type(http) == "table")
	assert(type(http.get) == "function")
	local resp, err = http.get({
        url = jsonUrl
    })
	print(err)
	assert(err == 'Get "'.. jsonUrl .. '": proxyconnect tcp: dial tcp 127.0.0.1:80: connect: connection refused')
	`
	s := lua.NewState()
	defer s.Close()

	s.SetGlobal("jsonUrl", lua.LString(jsonUrl))

	Preload(s, &config.Proxy{
		Enable: true,
		Url:    "http://127.0.0.1",
	}, nil)

	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}
func TestGetRequest(t *testing.T) {
	const str = `
	local http = require("http")
	assert(type(http) == "table")
	assert(type(http.get) == "function")
	local resp, err = http.get({
        url = jsonUrl
    })
	assert(err == nil)
	assert(resp.status_code == 200)
	assert(resp.headers['Content-Type'] == 'application/json; charset=utf-8')
	`
	eval(str, t)
}

func TestHeadRequest(t *testing.T) {
	const str = `
	local http = require("http")
	assert(type(http) == "table")
	assert(type(http.get) == "function")
	local resp, err = http.head({
        url = jsonUrl
    })
	assert(err == nil)
	assert(resp.status_code == 200)
	assert(resp.content_length ~= 0)
	`
	eval(str, t)
}

func TestDownloadFile(t *testing.T) {
	const str = `
	local http = require("http")
	assert(type(http) == "table")
	assert(type(http.get) == "function")
	local err = http.download_file({
        url = "https://version-fox.github.io/vfox-plugins/index.json"
    }, "index.json")
	assert(err == nil, [[must be nil]] )
	local err = http.download_file({
        url = "https://version-fox.github.io/vfox-plugins/xxx.json"
    }, "xxx.json")
	assert(err == "file not found")
	`
	defer os.Remove("index.json")
	eval(str, t)
	if !util.FileExists("index.json") {
		t.Error("file not exists")
	}
}

func eval(str string, t *testing.T) {
	s := lua.NewState()
	defer s.Close()

	s.SetGlobal("jsonUrl", lua.LString(jsonUrl))
	Preload(s, config.EmptyProxy, nil)

	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}

func setUserAgent(L *lua.LState, ua string) {
	L.SetGlobal(codec.NavigatorObjKey, codec.MustMarshal(L, codec.Navigator{UserAgent: ua}))
}

func TestUserAgentDefault(t *testing.T) {
	uaCh := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uaCh <- r.UserAgent()
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	ls := lua.NewState()
	ua := "vfox/0.7.0 vfox-nodejs/0.3.0"
	setUserAgent(ls, ua)
	defer ls.Close()
	Preload(ls, config.EmptyProxy, nil)

	script := fmt.Sprintf(`
	local http = require("http")
	local resp, err = http.get({
	    url = "%s"
	})
	assert(err == nil)
	assert(resp.status_code == 200)
	`, server.URL)
	if err := ls.DoString(script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case ua := <-uaCh:
		expected := "vfox/0.7.0 vfox-nodejs/0.3.0"
		if ua != expected {
			t.Fatalf("expected user-agent %q, got %q", expected, ua)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for request")
	}
}

func TestUserAgentAppend(t *testing.T) {
	uaCh := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uaCh <- r.UserAgent()
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	ls := lua.NewState()
	ua := "vfox/0.7.0 vfox-nodejs/0.3.0"
	setUserAgent(ls, ua)
	defer ls.Close()

	Preload(ls, config.EmptyProxy, nil)

	script := fmt.Sprintf(`
	local http = require("http")
	local resp, err = http.get({
	    url = "%s",
	    headers = {
	        ["User-Agent"] = "custom-agent"
	    }
	})
	assert(err == nil)
	assert(resp.status_code == 200)
	`, server.URL)
	if err := ls.DoString(script); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case ua := <-uaCh:
		expected := "custom-agent"
		if ua != expected {
			t.Fatalf("expected user-agent %q, got %q", expected, ua)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for request")
	}
}

// Exercise the same clients used by Lua, with short deadlines for local tests.
func TestRequestTimeouts(t *testing.T) {
	for _, proxyEnabled := range []bool{false, true} {
		for _, method := range []string{"get", "head", "download_file"} {
			for _, stallBody := range []bool{false, true} {
				if method == "head" && stallBody {
					continue
				}
				t.Run(fmt.Sprintf("proxy=%t/%s/body=%t", proxyEnabled, method, stallBody), func(t *testing.T) {
					t.Parallel()
					release := make(chan struct{})
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if stallBody {
							w.Header().Set("Content-Length", "100")
							w.WriteHeader(http.StatusOK)
							w.(http.Flusher).Flush()
						}
						select {
						case <-r.Context().Done():
						case <-release:
						case <-time.After(2 * time.Second):
						}
					}))
					defer server.Close()
					defer close(release)
					m := newModule(&config.Proxy{Enable: proxyEnabled, Url: server.URL}, nil)
					if m.client.Timeout != 30*time.Second || m.downloadClient.Timeout != 30*time.Minute {
						t.Fatal("unexpected default request/download deadlines")
					}
					m.client.Timeout = 100 * time.Millisecond
					m.downloadClient.Timeout = 100 * time.Millisecond
					ls := lua.NewState()
					defer ls.Close()
					ls.SetGlobal("http", ls.SetFuncs(ls.NewTable(), m.luaMap()))
					url := server.URL
					if proxyEnabled {
						url = "http://vfox-timeout.invalid/manifest"
					}
					ls.SetGlobal("url", lua.LString(url))
					ls.SetGlobal("destination", lua.LString(filepath.Join(t.TempDir(), "download")))
					script := fmt.Sprintf(`local resp, err = http.%s({url = url}); assert(resp == nil); assert(string.find(err, "deadline exceeded") or string.find(err, "timeout"), err)`, method)
					if method == "download_file" {
						script = `local err = http.download_file({url = url}, destination); assert(string.find(err, "deadline exceeded") or string.find(err, "timeout"), err)`
					}
					if err := ls.DoString(script); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	}
}

func TestDownloadUsesSeparateDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("complete download"))
	}))
	defer server.Close()
	m := newModule(config.EmptyProxy, nil)
	m.client.Timeout = time.Millisecond
	m.downloadClient.Timeout = time.Second
	ls := lua.NewState()
	defer ls.Close()
	ls.SetGlobal("http", ls.SetFuncs(ls.NewTable(), m.luaMap()))
	ls.SetGlobal("url", lua.LString(server.URL))
	destination := filepath.Join(t.TempDir(), "download")
	ls.SetGlobal("destination", lua.LString(destination))
	if err := ls.DoString(`assert(http.download_file({url = url}, destination) == nil)`); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(destination)
	if err != nil || string(body) != "complete download" {
		t.Fatalf("download = %q, err = %v", body, err)
	}
}
