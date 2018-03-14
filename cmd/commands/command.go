// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"errors"
	"github.com/SoulDemon/mattermostp/app"
	"github.com/SoulDemon/mattermostp/cmd"
	"github.com/SoulDemon/mattermostp/model"
	"github.com/spf13/cobra"
)

var CommandCmd = &cobra.Command{
	Use:   "command",
	Short: "Management of slash commands",
}

var CommandMoveCmd = &cobra.Command{
	Use:     "move",
	Short:   "Move a slash command to a different team",
	Long:    `Move a slash command to a different team. Commands can be specified by [team]:[command-trigger-word]. ie. myteam:trigger or by command ID.`,
	Example: `  command move newteam oldteam:command`,
	RunE:    moveCommandCmdF,
}

func init() {
	CommandCmd.AddCommand(
		CommandMoveCmd,
	)
	cmd.RootCmd.AddCommand(CommandCmd)
}

func moveCommandCmdF(command *cobra.Command, args []string) error {
	a, err := cmd.InitDBCommandContextCobra(command)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.New("Enter the destination team and at least one comamnd to move.")
	}

	team := getTeamFromTeamArg(a, args[0])
	if team == nil {
		return errors.New("Unable to find destination team '" + args[0] + "'")
	}

	commands := getCommandsFromCommandArgs(a, args[1:])
	cmd.CommandPrintErrorln(commands)
	for i, command := range commands {
		if command == nil {
			cmd.CommandPrintErrorln("Unable to find command '" + args[i+1] + "'")
			continue
		}
		if err := moveCommand(a, team, command); err != nil {
			cmd.CommandPrintErrorln("Unable to move command '" + command.Trigger + "' error: " + err.Error())
		} else {
			cmd.CommandPrettyPrintln("Moved command '" + command.Trigger + "'")
		}
	}

	return nil
}

func moveCommand(a *app.App, team *model.Team, command *model.Command) *model.AppError {
	return a.MoveCommand(team, command)
}
