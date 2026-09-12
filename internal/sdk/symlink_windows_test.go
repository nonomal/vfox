//go:build windows

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

package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"

	"github.com/version-fox/vfox/internal/env"
)

func TestSdk_CreateDirSymlinksReusesWindowsJunction(t *testing.T) {
	t.Setenv("GODEBUG", "winsymlink=1")
	root := t.TempDir()
	target := filepath.Join(root, "installed")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	rt := &Runtime{Name: "test-sdk", Version: "1.0.0", Path: target}
	linkDir := filepath.Join(root, "sdks")
	sdk := &impl{}
	if err := sdk.createDirSymlinks(rt, linkDir); err != nil {
		t.Fatal(err)
	}

	// Keep the junction open without FILE_SHARE_DELETE. Reusing the existing
	// link must work even when Windows prevents deleting and recreating it.
	linkPath, err := windows.UTF16PtrFromString(filepath.Join(linkDir, rt.Name))
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(linkPath, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.CloseHandle(handle); err != nil {
			t.Error(err)
		}
	})

	if err := sdk.createDirSymlinks(rt, linkDir); err != nil {
		t.Fatalf("existing junction should be reused without deletion: %v", err)
	}
}

func TestSdk_CreateDirSymlinksRetargetsWindowsJunction(t *testing.T) {
	root := t.TempDir()
	linkDir := filepath.Join(root, "sdks")
	sdk := &impl{}
	for _, version := range []Version{"1.0.0", "2.0.0"} {
		target := filepath.Join(root, string(version))
		if err := os.Mkdir(target, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, "version"), []byte(version), 0600); err != nil {
			t.Fatal(err)
		}
		rt := &Runtime{Name: "test-sdk", Version: version, Path: target}
		if err := sdk.createDirSymlinks(rt, linkDir); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(linkDir, rt.Name)
		if got, err := env.ReadDirSymlink(link); err != nil || got != target {
			t.Fatalf("junction target = %q, %v; want %q", got, err, target)
		}
		if got, err := os.ReadFile(filepath.Join(link, "version")); err != nil || string(got) != string(version) {
			t.Fatalf("version through junction = %q, %v; want %q", got, err, version)
		}
	}
	if got, err := os.ReadFile(filepath.Join(root, "1.0.0", "version")); err != nil || string(got) != "1.0.0" {
		t.Fatalf("switching the junction changed the old installation: %q, %v", got, err)
	}
}
