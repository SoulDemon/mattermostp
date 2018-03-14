// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"testing"

	"github.com/mattermost/mattermost-server/api"
	"github.com/mattermost/mattermost-server/cmd"
	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/require"
)

func TestCreateUserWithTeam(t *testing.T) {
	th := api.Setup().InitSystemAdmin()
	defer th.TearDown()

	id := model.NewId()
	email := "success+" + id + "@simulator.amazonses.com"
	username := "name" + id

	cmd.CheckCommand(t, "user", "create", "--email", email, "--password", "mypassword1", "--username", username)

	cmd.CheckCommand(t, "team", "add", th.SystemAdminTeam.Id, email)

	profiles := th.SystemAdminClient.Must(th.SystemAdminClient.GetProfilesInTeam(th.SystemAdminTeam.Id, 0, 1000, "")).Data.(map[string]*model.User)

	found := false

	for _, user := range profiles {
		if user.Email == email {
			found = true
		}

	}

	if !found {
		t.Fatal("Failed to create User")
	}
}

func TestCreateUserWithoutTeam(t *testing.T) {
	th := api.Setup()
	defer th.TearDown()

	id := model.NewId()
	email := "success+" + id + "@simulator.amazonses.com"
	username := "name" + id

	cmd.CheckCommand(t, "user", "create", "--email", email, "--password", "mypassword1", "--username", username)

	if result := <-th.App.Srv.Store.User().GetByEmail(email); result.Err != nil {
		t.Fatal()
	} else {
		user := result.Data.(*model.User)
		if user.Email != email {
			t.Fatal()
		}
	}
}

func TestResetPassword(t *testing.T) {
	th := api.Setup().InitBasic()
	defer th.TearDown()

	cmd.CheckCommand(t, "user", "password", th.BasicUser.Email, "password2")

	th.BasicClient.Logout()
	th.BasicUser.Password = "password2"
	th.LoginBasic()
}

func TestMakeUserActiveAndInactive(t *testing.T) {
	th := api.Setup().InitBasic()
	defer th.TearDown()

	// first inactivate the user
	cmd.CheckCommand(t, "user", "deactivate", th.BasicUser.Email)

	// activate the inactive user
	cmd.CheckCommand(t, "user", "activate", th.BasicUser.Email)
}

func TestChangeUserEmail(t *testing.T) {
	th := api.Setup().InitBasic()
	defer th.TearDown()

	newEmail := model.NewId() + "@mattermost-test.com"

	cmd.CheckCommand(t, "user", "email", th.BasicUser.Username, newEmail)
	if result := <-th.App.Srv.Store.User().GetByEmail(th.BasicUser.Email); result.Err == nil {
		t.Fatal("should've updated to the new email")
	}
	if result := <-th.App.Srv.Store.User().GetByEmail(newEmail); result.Err != nil {
		t.Fatal()
	} else {
		user := result.Data.(*model.User)
		if user.Email != newEmail {
			t.Fatal("should've updated to the new email")
		}
	}

	// should fail because using an invalid email
	require.Error(t, cmd.RunCommand(t, "user", "email", th.BasicUser.Username, "wrong$email.com"))

	// should fail because user not found
	require.Error(t, cmd.RunCommand(t, "user", "email", "invalidUser", newEmail))

}
