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

package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirSymlinkRecognizesWindowsJunctions(t *testing.T) {
	// Go 1.23 and later do not mark junctions with os.ModeSymlink.
	t.Setenv("GODEBUG", "winsymlink=1")
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("ordinary file"), 0600); err != nil {
		t.Fatal(err)
	}
	junction := filepath.Join(root, "junction")
	if err := CreateDirSymlink(target, junction); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(root, "dangling")
	if err := CreateDirSymlink(filepath.Join(root, "missing-target"), dangling); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		path string
		want bool
	}{
		{"junction", junction, true},
		{"dangling junction", dangling, true},
		{"ordinary directory", target, false},
		{"ordinary file", file, false},
		{"missing path", filepath.Join(root, "missing"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsDirSymlink(tc.path); got != tc.want {
				t.Errorf("IsDirSymlink(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
