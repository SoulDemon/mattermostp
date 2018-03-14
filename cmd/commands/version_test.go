// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"testing"

	"github.com/SoulDemon/mattermostp/cmd"
)

func TestVersion(t *testing.T) {
	cmd.CheckCommand(t, "version")
}
