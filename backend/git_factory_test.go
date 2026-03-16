// Copyright 2021-2023 Charles Francoise
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestGitFactory(t *testing.T) {
	f := gitFactory{}

	testGenericFactory(t, f)

	_, err := f.New(map[string]interface{}{})
	if err == nil {
		t.Error("expected error")
	}

	_, err = f.New(map[string]interface{}{
		"git-url":  "/tmp/repo.git",
		"git-path": "store.scrt",
	})
	if err == nil {
		t.Error("expected error")
	}

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	_, err = git.PlainInit(remotePath, true)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.New(map[string]interface{}{
		"git-url":        remotePath,
		"git-path":       "store.scrt",
		"git-local-path": filepath.Join(root, "clone-flat"),
	})
	if err != nil {
		t.Error(err)
	}

	_, err = f.New(map[string]interface{}{
		"git": map[string]interface{}{
			"url":        remotePath,
			"path":       "store.scrt",
			"local-path": filepath.Join(root, "clone-nested"),
		},
	})
	if err != nil {
		t.Error(err)
	}
}
