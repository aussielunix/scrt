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

//go:generate mockgen -destination mock_backend.go -package cmd "github.com/loderunner/scrt/backend" Backend

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/loderunner/scrt/backend"
)

func TestInitCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockBackend(ctrl)
	backend.Backends["mock"] = newMockFactory(mockBackend)

	viper.Reset()
	viper.Set(configKeyPassword, "toto")
	viper.Set(configKeyStorage, "mock")

	mockBackend.EXPECT().ExistsContext(ctxMatcher).Return(false, nil)
	mockBackend.EXPECT().SaveContext(ctxMatcher, gomock.Any())

	args := []string{"path"}
	err := initCmd.RunE(initCmd, args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitOverwrite(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBackend := NewMockBackend(ctrl)
	backend.Backends["mock"] = newMockFactory(mockBackend)

	viper.Reset()
	viper.Set(configKeyPassword, "toto")
	viper.Set(configKeyStorage, "mock")

	mockBackend.EXPECT().ExistsContext(ctxMatcher).Return(true, nil)

	args := []string{"path"}
	err := initCmd.RunE(initCmd, args)
	if err == nil {
		t.Fatal("expected error")
	}

	mockBackend.EXPECT().ExistsContext(ctxMatcher).Return(true, nil)
	mockBackend.EXPECT().SaveContext(ctxMatcher, gomock.Any())

	err = initCmd.Flags().Set("overwrite", "true")
	if err != nil {
		t.Fatal(err)
	}

	err = initCmd.RunE(initCmd, args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitWritesConfigFileFromFlags(t *testing.T) {
	viper.Reset()

	root := t.TempDir()
	storePath := filepath.Join(root, "store.scrt")
	configPath := filepath.Join(root, "scrt.yml")

	oldConfigFile := configFile
	configFile = configPath
	defer func() {
		configFile = oldConfigFile
		for _, name := range []string{
			configKeyStorage,
			configKeyPassword,
			"local-path",
			"verbose",
		} {
			flag := RootCmd.PersistentFlags().Lookup(name)
			if flag != nil {
				flag.Changed = false
			}
		}
		_ = RootCmd.PersistentFlags().Set(configKeyStorage, "")
		_ = RootCmd.PersistentFlags().Set(configKeyPassword, "")
		_ = RootCmd.PersistentFlags().Set("local-path", "")
		_ = RootCmd.PersistentFlags().Set("verbose", "false")
	}()

	viper.Set(configKeyStorage, "local")
	viper.Set(configKeyPassword, "toto")
	viper.Set("local-path", storePath)
	viper.Set("verbose", true)

	err := RootCmd.PersistentFlags().Set(configKeyStorage, "local")
	if err != nil {
		t.Fatal(err)
	}
	err = RootCmd.PersistentFlags().Set(configKeyPassword, "toto")
	if err != nil {
		t.Fatal(err)
	}
	err = RootCmd.PersistentFlags().Set("local-path", storePath)
	if err != nil {
		t.Fatal(err)
	}
	err = RootCmd.PersistentFlags().Set("verbose", "true")
	if err != nil {
		t.Fatal(err)
	}

	err = initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]interface{}
	err = yaml.Unmarshal(data, &got)
	if err != nil {
		t.Fatal(err)
	}

	if got[configKeyStorage] != "local" {
		t.Fatalf("expected storage local, got %#v", got[configKeyStorage])
	}
	if got[configKeyPassword] != "toto" {
		t.Fatalf("expected password toto, got %#v", got[configKeyPassword])
	}
	if got["verbose"] != true {
		t.Fatalf("expected verbose true, got %#v", got["verbose"])
	}

	local, ok := got["local"].(map[interface{}]interface{})
	if !ok {
		t.Fatalf("expected local config section, got %#v", got["local"])
	}
	if local["path"] != storePath {
		t.Fatalf("expected local path %q, got %#v", storePath, local["path"])
	}
}

func TestInitWritesConfigFileWithTildePath(t *testing.T) {
	viper.Reset()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	storePath := filepath.Join(homeDir, "store.scrt")
	configRelPath := "~/.scrt/config.yml"
	configPath := filepath.Join(homeDir, ".scrt", "config.yml")

	oldConfigFile := configFile
	configFile = configRelPath
	defer func() {
		configFile = oldConfigFile
		for _, name := range []string{
			configKeyStorage,
			configKeyPassword,
			"local-path",
		} {
			flag := RootCmd.PersistentFlags().Lookup(name)
			if flag != nil {
				flag.Changed = false
			}
		}
		_ = RootCmd.PersistentFlags().Set(configKeyStorage, "")
		_ = RootCmd.PersistentFlags().Set(configKeyPassword, "")
		_ = RootCmd.PersistentFlags().Set("local-path", "")
	}()

	viper.Set(configKeyStorage, "local")
	viper.Set(configKeyPassword, "toto")
	viper.Set("local-path", storePath)

	err := RootCmd.PersistentFlags().Set(configKeyStorage, "local")
	if err != nil {
		t.Fatal(err)
	}
	err = RootCmd.PersistentFlags().Set(configKeyPassword, "toto")
	if err != nil {
		t.Fatal(err)
	}
	err = RootCmd.PersistentFlags().Set("local-path", storePath)
	if err != nil {
		t.Fatal(err)
	}

	err = initCmd.RunE(initCmd, []string{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitExecuteWritesConfigFileFromCLIFlags(t *testing.T) {
	viper.Reset()
	err := viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	storePath := filepath.Join(root, "store.scrt")
	configPath := filepath.Join(root, "scrt.yml")

	RootCmd.SetArgs([]string{
		"init",
		"--config=" + configPath,
		"--storage=local",
		"--password=toto",
		"--local-path=" + storePath,
		"--verbose",
	})
	defer func() {
		RootCmd.SetArgs(nil)
		viper.Reset()
		for _, name := range []string{
			configKeyStorage,
			configKeyPassword,
			"local-path",
			"verbose",
			"config",
		} {
			flag := RootCmd.PersistentFlags().Lookup(name)
			if flag != nil {
				flag.Changed = false
			}
		}
		_ = RootCmd.PersistentFlags().Set(configKeyStorage, "")
		_ = RootCmd.PersistentFlags().Set(configKeyPassword, "")
		_ = RootCmd.PersistentFlags().Set("local-path", "")
		_ = RootCmd.PersistentFlags().Set("verbose", "false")
		_ = RootCmd.PersistentFlags().Set("config", "")
	}()

	err = RootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "{}\n" {
		t.Fatal("expected config file to contain CLI options")
	}
}
