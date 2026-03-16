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

package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/loderunner/scrt/backend"
)

func TestRootCmd(t *testing.T) {
	viper.Reset()

	fs := afero.NewMemMapFs()
	viper.SetFs(fs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockBackend(ctrl)
	backend.Backends["mock"] = newMockFactory(mockBackend)

	err := RootCmd.PersistentPreRunE(RootCmd, []string{})
	if err == nil {
		t.Fatal("expected error")
	}

	viper.Set(configKeyStorage, "mock")
	err = RootCmd.PersistentPreRunE(RootCmd, []string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRootCmdPromptPassword(t *testing.T) {
	hijack()
	defer restore()

	viper.Reset()

	fs := afero.NewMemMapFs()
	viper.SetFs(fs)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockBackend(ctrl)
	backend.Backends["mock"] = newMockFactory(mockBackend)

	viper.Set(configKeyStorage, "mock")
	RootCmd.PersistentFlags().Lookup("password").Changed = false

	err := RootCmd.PersistentFlags().Set("password", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = RootCmd.PersistentFlags().Set("password", "")
		RootCmd.PersistentFlags().Lookup("password").Changed = false
	}()

	_, err = hijackStdin.WriteString("toto\n")
	if err != nil {
		t.Fatal(err)
	}
	_ = hijackStdin.Close()

	err = RootCmd.PersistentPreRunE(RootCmd, []string{})
	if err != nil {
		t.Fatal(err)
	}
	if got := viper.GetString(configKeyPassword); got != "toto" {
		t.Fatalf("expected prompted password, got %q", got)
	}
}

func TestPromptPasswordEmpty(t *testing.T) {
	_, err := promptPassword(strings.NewReader("\n"), io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPromptPassword(t *testing.T) {
	password, err := promptPassword(strings.NewReader("toto\n"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if password != "toto" {
		t.Fatalf("expected toto, got %q", password)
	}
}

func TestRootCmdPasswordFlagWithoutValue(t *testing.T) {
	RootCmd.PersistentFlags().Lookup("password").Changed = false

	err := RootCmd.ParseFlags([]string{"list", "--password"})
	if err != nil {
		t.Fatal(err)
	}

	if !RootCmd.Flag("password").Changed {
		t.Fatal("expected password flag to be marked as changed")
	}
	if got := RootCmd.Flag("password").Value.String(); got != promptPasswordSentinel {
		t.Fatalf("expected prompt sentinel, got %q", got)
	}

	_ = RootCmd.PersistentFlags().Set("password", "")
	RootCmd.PersistentFlags().Lookup("password").Changed = false
}

func TestRootCmdRegistersBackendFlags(t *testing.T) {
	if RootCmd.PersistentFlags().Lookup("git-local-path") == nil {
		t.Fatal("expected git-local-path flag to be registered")
	}
	if RootCmd.PersistentFlags().Lookup("local-path") == nil {
		t.Fatal("expected local-path flag to be registered")
	}
	if RootCmd.PersistentFlags().Lookup("s3-bucket-name") == nil {
		t.Fatal("expected s3-bucket-name flag to be registered")
	}
}

func TestReadConfigAllowsMissingExplicitConfigFile(t *testing.T) {
	viper.Reset()
	viper.SetFs(afero.NewOsFs())

	oldConfigFile := configFile
	configFile = filepath.Join(t.TempDir(), "missing.yml")
	defer func() {
		configFile = oldConfigFile
	}()

	err := readConfig(RootCmd)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadConfigUsesLocalConfigFirst(t *testing.T) {
	viper.Reset()
	viper.SetFs(afero.NewOsFs())
	oldConfigFile := configFile
	configFile = ""
	defer func() {
		configFile = oldConfigFile
	}()

	root := t.TempDir()
	err := os.WriteFile(
		filepath.Join(root, "config.yml"),
		[]byte("storage: local\npassword: localpass\n"),
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}

	homeDir := t.TempDir()
	err = os.MkdirAll(filepath.Join(homeDir, ".scrt"), 0o700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(
		filepath.Join(homeDir, ".scrt", "config.yml"),
		[]byte("storage: git\npassword: homepass\n"),
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	err = os.Chdir(root)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", homeDir)
	homedir.Reset()

	err = readConfig(RootCmd)
	if err != nil {
		t.Fatal(err)
	}

	if got := viper.GetString(configKeyStorage); got != "local" {
		t.Fatalf("expected local config, got %q", got)
	}
	if got := viper.GetString(configKeyPassword); got != "localpass" {
		t.Fatalf("expected local password, got %q", got)
	}
}

func TestReadConfigUsesHomeConfigWhenLocalMissing(t *testing.T) {
	viper.Reset()
	viper.SetFs(afero.NewOsFs())
	oldConfigFile := configFile
	configFile = ""
	defer func() {
		configFile = oldConfigFile
	}()

	root := t.TempDir()
	homeDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(homeDir, ".scrt"), 0o700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(
		filepath.Join(homeDir, ".scrt", "config.yml"),
		[]byte("storage: git\npassword: homepass\n"),
		0o600,
	)
	if err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	err = os.Chdir(root)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", homeDir)
	homedir.Reset()

	err = readConfig(RootCmd)
	if err != nil {
		t.Fatal(err)
	}

	if got := viper.GetString(configKeyStorage); got != "git" {
		t.Fatalf("expected home config, got %q", got)
	}
	if got := viper.GetString(configKeyPassword); got != "homepass" {
		t.Fatalf("expected home password, got %q", got)
	}
}
