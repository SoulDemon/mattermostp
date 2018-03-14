// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package cmd

import (
	"github.com/SoulDemon/mattermostp/app"
	"github.com/SoulDemon/mattermostp/model"
	"github.com/souldemon/mattermostp/utils"
	"github.com/spf13/cobra"
)

func InitDBCommandContextCobra(cmd *cobra.Command) (*app.App, error) {
	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	a, err := InitDBCommandContext(config)
	if err != nil {
		// Returning an error just prints the usage message, so actually panic
		panic(err)
	}

	return a, nil
}

func InitDBCommandContext(configFileLocation string) (*app.App, error) {
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, err
	}
	model.AppErrorInit(utils.T)

	utils.ConfigureCmdLineLog()

	a, err := app.New(app.ConfigFile(configFileLocation))
	if err != nil {
		return nil, err
	}

	if model.BuildEnterpriseReady == "true" {
		a.LoadLicense()
	}

	return a, nil
}
