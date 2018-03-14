// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package rpcplugin

import (
	"testing"

	"github.com/SoulDemon/mattermostp/plugin/rpcplugin/rpcplugintest"
)

func TestSupervisorProvider(t *testing.T) {
	rpcplugintest.TestSupervisorProvider(t, SupervisorProvider)
}
