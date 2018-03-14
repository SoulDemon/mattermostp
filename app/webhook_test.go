// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
)

func TestCreateIncomingWebhookForChannel(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	type TestCase struct {
		EnableIncomingHooks        bool
		EnablePostUsernameOverride bool
		EnablePostIconOverride     bool
		IncomingWebhook            model.IncomingWebhook

		ExpectedError           bool
		ExpectedIncomingWebhook *model.IncomingWebhook
	}

	for name, tc := range map[string]TestCase{
		"webhooks not enabled": {
			EnableIncomingHooks:        false,
			EnablePostUsernameOverride: false,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},

			ExpectedError:           true,
			ExpectedIncomingWebhook: nil,
		},
		"valid: username and post icon url ignored, since override not enabled": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: false,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    ":invalid and ignored:",
				IconURL:     "ignored",
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},
		},
		"invalid username, override enabled": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    ":invalid:",
			},

			ExpectedError:           true,
			ExpectedIncomingWebhook: nil,
		},
		"valid, no username or post icon url provided": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     true,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},
		},
		"valid, with username and post icon": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     true,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    "valid",
				IconURL:     "http://example.com/icon",
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    "valid",
				IconURL:     "http://example.com/icon",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableIncomingWebhooks = tc.EnableIncomingHooks })
			th.App.UpdateConfig(func(cfg *model.Config) {
				cfg.ServiceSettings.EnablePostUsernameOverride = tc.EnablePostUsernameOverride
			})
			th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnablePostIconOverride = tc.EnablePostIconOverride })

			createdHook, err := th.App.CreateIncomingWebhookForChannel(th.BasicUser.Id, th.BasicChannel, &tc.IncomingWebhook)
			if tc.ExpectedError && err == nil {
				t.Fatal("should have failed")
			} else if !tc.ExpectedError && err != nil {
				t.Fatalf("should not have failed: %v", err.Error())
			}
			if createdHook != nil {
				defer th.App.DeleteIncomingWebhook(createdHook.Id)
			}
			if tc.ExpectedIncomingWebhook == nil {
				assert.Nil(createdHook, "expected nil webhook")
			} else if assert.NotNil(createdHook, "expected non-nil webhook") {
				assert.Equal(tc.ExpectedIncomingWebhook.DisplayName, createdHook.DisplayName)
				assert.Equal(tc.ExpectedIncomingWebhook.Description, createdHook.Description)
				assert.Equal(tc.ExpectedIncomingWebhook.ChannelId, createdHook.ChannelId)
				assert.Equal(tc.ExpectedIncomingWebhook.Username, createdHook.Username)
				assert.Equal(tc.ExpectedIncomingWebhook.IconURL, createdHook.IconURL)
			}
		})
	}
}

func TestUpdateIncomingWebhook(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	type TestCase struct {
		EnableIncomingHooks        bool
		EnablePostUsernameOverride bool
		EnablePostIconOverride     bool
		IncomingWebhook            model.IncomingWebhook

		ExpectedError           bool
		ExpectedIncomingWebhook *model.IncomingWebhook
	}

	for name, tc := range map[string]TestCase{
		"webhooks not enabled": {
			EnableIncomingHooks:        false,
			EnablePostUsernameOverride: false,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},

			ExpectedError:           true,
			ExpectedIncomingWebhook: nil,
		},
		"valid: username and post icon url ignored, since override not enabled": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: false,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    ":invalid and ignored:",
				IconURL:     "ignored",
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},
		},
		"invalid username, override enabled": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     false,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    ":invalid:",
			},

			ExpectedError:           true,
			ExpectedIncomingWebhook: nil,
		},
		"valid, no username or post icon url provided": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     true,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
			},
		},
		"valid, with username and post icon": {
			EnableIncomingHooks:        true,
			EnablePostUsernameOverride: true,
			EnablePostIconOverride:     true,
			IncomingWebhook: model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    "valid",
				IconURL:     "http://example.com/icon",
			},

			ExpectedError: false,
			ExpectedIncomingWebhook: &model.IncomingWebhook{
				DisplayName: "title",
				Description: "description",
				ChannelId:   th.BasicChannel.Id,
				Username:    "valid",
				IconURL:     "http://example.com/icon",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableIncomingWebhooks = true })

			hook, err := th.App.CreateIncomingWebhookForChannel(th.BasicUser.Id, th.BasicChannel, &model.IncomingWebhook{
				ChannelId: th.BasicChannel.Id,
			})
			if err != nil {
				t.Fatal(err.Error())
			}
			defer th.App.DeleteIncomingWebhook(hook.Id)

			th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableIncomingWebhooks = tc.EnableIncomingHooks })
			th.App.UpdateConfig(func(cfg *model.Config) {
				cfg.ServiceSettings.EnablePostUsernameOverride = tc.EnablePostUsernameOverride
			})
			th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnablePostIconOverride = tc.EnablePostIconOverride })

			updatedHook, err := th.App.UpdateIncomingWebhook(hook, &tc.IncomingWebhook)
			if tc.ExpectedError && err == nil {
				t.Fatal("should have failed")
			} else if !tc.ExpectedError && err != nil {
				t.Fatalf("should not have failed: %v", err.Error())
			}
			if tc.ExpectedIncomingWebhook == nil {
				assert.Nil(updatedHook, "expected nil webhook")
			} else if assert.NotNil(updatedHook, "expected non-nil webhook") {
				assert.Equal(tc.ExpectedIncomingWebhook.DisplayName, updatedHook.DisplayName)
				assert.Equal(tc.ExpectedIncomingWebhook.Description, updatedHook.Description)
				assert.Equal(tc.ExpectedIncomingWebhook.ChannelId, updatedHook.ChannelId)
				assert.Equal(tc.ExpectedIncomingWebhook.Username, updatedHook.Username)
				assert.Equal(tc.ExpectedIncomingWebhook.IconURL, updatedHook.IconURL)
			}
		})
	}
}

func TestCreateWebhookPost(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()

	th.App.UpdateConfig(func(cfg *model.Config) { cfg.ServiceSettings.EnableIncomingWebhooks = true })

	hook, err := th.App.CreateIncomingWebhookForChannel(th.BasicUser.Id, th.BasicChannel, &model.IncomingWebhook{ChannelId: th.BasicChannel.Id})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer th.App.DeleteIncomingWebhook(hook.Id)

	post, err := th.App.CreateWebhookPost(hook.UserId, th.BasicChannel, "foo", "user", "http://iconurl", model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "text",
			},
		},
		"webhook_display_name": hook.DisplayName,
	}, model.POST_SLACK_ATTACHMENT, "")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, k := range []string{"from_webhook", "attachments", "webhook_display_name"} {
		if _, ok := post.Props[k]; !ok {
			t.Log("missing one props: " + k)
			t.Fatal(k)
		}
	}

	_, err = th.App.CreateWebhookPost(hook.UserId, th.BasicChannel, "foo", "user", "http://iconurl", nil, model.POST_SYSTEM_GENERIC, "")
	if err == nil {
		t.Fatal("should have failed - bad post type")
	}

	expectedText := "`<>|<>|`"
	post, err = th.App.CreateWebhookPost(hook.UserId, th.BasicChannel, expectedText, "user", "http://iconurl", model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "text",
			},
		},
		"webhook_display_name": hook.DisplayName,
	}, model.POST_SLACK_ATTACHMENT, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, expectedText, post.Message)

	expectedText = "< | \n|\n>"
	post, err = th.App.CreateWebhookPost(hook.UserId, th.BasicChannel, expectedText, "user", "http://iconurl", model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "text",
			},
		},
		"webhook_display_name": hook.DisplayName,
	}, model.POST_SLACK_ATTACHMENT, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, expectedText, post.Message)

	expectedText = `commit bc95839e4a430ace453e8b209a3723c000c1729a
Author: foo <foo@example.org>
Date:   Thu Mar 1 19:46:54 2018 +0300

    commit message 2

  test | 1 +
 1 file changed, 1 insertion(+)

commit 5df78b7139b543997838071cd912e375d8bd69b2
Author: foo <foo@example.org>
Date:   Thu Mar 1 19:46:48 2018 +0300

    commit message 1

 test | 3 +++
 1 file changed, 3 insertions(+)`
	post, err = th.App.CreateWebhookPost(hook.UserId, th.BasicChannel, expectedText, "user", "http://iconurl", model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "text",
			},
		},
		"webhook_display_name": hook.DisplayName,
	}, model.POST_SLACK_ATTACHMENT, "")
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, expectedText, post.Message)
}

func TestSplitWebhookPost(t *testing.T) {
	type TestCase struct {
		Post     *model.Post
		Expected []*model.Post
	}

	for name, tc := range map[string]TestCase{
		"LongPost": {
			Post: &model.Post{
				Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES*3/2),
			},
			Expected: []*model.Post{
				{
					Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES),
				},
				{
					Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES/2),
				},
			},
		},
		"LongPostAndMultipleAttachments": {
			Post: &model.Post{
				Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES*3/2),
				Props: map[string]interface{}{
					"attachments": []*model.SlackAttachment{
						&model.SlackAttachment{
							Text: strings.Repeat("本", 1000),
						},
						&model.SlackAttachment{
							Text: strings.Repeat("本", 2000),
						},
						&model.SlackAttachment{
							Text: strings.Repeat("本", model.POST_PROPS_MAX_USER_RUNES-1000),
						},
					},
				},
			},
			Expected: []*model.Post{
				{
					Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES),
				},
				{
					Message: strings.Repeat("本", model.POST_MESSAGE_MAX_RUNES/2),
					Props: map[string]interface{}{
						"attachments": []*model.SlackAttachment{
							&model.SlackAttachment{
								Text: strings.Repeat("本", 1000),
							},
							&model.SlackAttachment{
								Text: strings.Repeat("本", 2000),
							},
						},
					},
				},
				{
					Props: map[string]interface{}{
						"attachments": []*model.SlackAttachment{
							&model.SlackAttachment{
								Text: strings.Repeat("本", model.POST_PROPS_MAX_USER_RUNES-1000),
							},
						},
					},
				},
			},
		},
		"UnsplittableProps": {
			Post: &model.Post{
				Message: "foo",
				Props: map[string]interface{}{
					"foo": strings.Repeat("x", model.POST_PROPS_MAX_USER_RUNES*2),
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			splits, err := SplitWebhookPost(tc.Post)
			if tc.Expected == nil {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
			assert.Equal(t, len(tc.Expected), len(splits))
			for i, split := range splits {
				if i < len(tc.Expected) {
					assert.Equal(t, tc.Expected[i].Message, split.Message)
					assert.Equal(t, tc.Expected[i].Props["attachments"], split.Props["attachments"])
				}
			}
		})
	}
}
