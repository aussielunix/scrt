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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/loderunner/scrt/store"
)

func TestGitExistsLoadSave(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	s := store.NewStore()
	err := s.Set("hello", []byte("world"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := store.WriteStore([]byte("password"), s)
	if err != nil {
		t.Fatal(err)
	}

	remotePath := createSeededGitRemote(t, "nested/store.scrt", data)
	clonePath := filepath.Join(t.TempDir(), "clone")

	backendIntf, err := newGit(t.Context(), map[string]interface{}{
		"git-url":        remotePath,
		"git-path":       "nested/store.scrt",
		"git-local-path": clonePath,
	})
	if err != nil {
		t.Fatal(err)
	}

	b, ok := backendIntf.(gitBackend)
	if !ok {
		t.Fatal("expected git backend")
	}

	exists, err := b.Exists()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected store to exist")
	}

	got, err := b.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, got) {
		t.Fatalf("expected %#v, got %#v", data, got)
	}

	_, err = os.Stat(filepath.Join(clonePath, ".git"))
	if err != nil {
		t.Fatal(err)
	}

	updated := []byte("updated store data")
	err = b.Save(updated)
	if err != nil {
		t.Fatal(err)
	}

	localData, err := os.ReadFile(filepath.Join(clonePath, "nested/store.scrt"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(updated, localData) {
		t.Fatalf("expected %#v, got %#v", updated, localData)
	}

	checkoutPath := filepath.Join(t.TempDir(), "checkout")
	_, err = git.PlainClone(checkoutPath, false, &git.CloneOptions{URL: remotePath})
	if err != nil {
		t.Fatal(err)
	}

	remoteData, err := os.ReadFile(filepath.Join(checkoutPath, "nested/store.scrt"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(updated, remoteData) {
		t.Fatalf("expected %#v, got %#v", updated, remoteData)
	}
}

func createSeededGitRemote(t *testing.T, storePath string, data []byte) string {
	t.Helper()

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	_, err := git.PlainInit(remotePath, true)
	if err != nil {
		t.Fatal(err)
	}

	workPath := filepath.Join(root, "work")
	repo, err := git.PlainInit(workPath, false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remotePath},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(filepath.Dir(filepath.Join(workPath, storePath)), 0o700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(workPath, storePath), data, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Add(storePath)
	if err != nil {
		t.Fatal(err)
	}

	sig := &object.Signature{
		Name:  "scrt test",
		Email: "scrt@example.com",
		When:  time.Now(),
	}
	_, err = w.Commit("seed", &git.CommitOptions{
		Author:    sig,
		Committer: sig,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Push(&git.PushOptions{RemoteName: git.DefaultRemoteName})
	if err != nil {
		t.Fatal(err)
	}

	return remotePath
}
