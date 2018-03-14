// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SoulDemon/mattermostp/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SoulDemon/mattermostp/model"
)

func TestConfigValidate(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "config.json")
	config := &model.Config{}
	config.SetDefaults()
	require.NoError(t, ioutil.WriteFile(path, []byte(config.ToJson()), 0600))

	assert.Error(t, cmd.RunCommand(t, "--config", "foo.json", "config", "validate"))
	assert.NoError(t, cmd.RunCommand(t, "--config", path, "config", "validate"))
}
