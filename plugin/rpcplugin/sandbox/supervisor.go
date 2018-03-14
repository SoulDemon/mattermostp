// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sandbox

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/SoulDemon/mattermostp/model"
	"github.com/SoulDemon/mattermostp/plugin"
	"github.com/SoulDemon/mattermostp/plugin/rpcplugin"
)

func SupervisorProvider(bundle *model.BundleInfo) (plugin.Supervisor, error) {
	return rpcplugin.SupervisorWithNewProcessFunc(bundle, func(ctx context.Context) (rpcplugin.Process, io.ReadWriteCloser, error) {
		executable := filepath.Clean(filepath.Join(".", bundle.Manifest.Backend.Executable))
		if strings.HasPrefix(executable, "..") {
			return nil, nil, fmt.Errorf("invalid backend executable")
		}
		return NewProcess(ctx, &Configuration{
			MountPoints: []*MountPoint{{
				Source:      bundle.Path,
				Destination: "/plugin",
				ReadOnly:    true,
			}},
			WorkingDirectory: "/plugin",
		}, filepath.Join("/plugin", executable))
	})
}
