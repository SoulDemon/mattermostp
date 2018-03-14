package api4

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGetPing(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	goRoutineHealthThreshold := *th.App.Config().ServiceSettings.GoroutineHealthThreshold
	defer func() {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.GoroutineHealthThreshold = goRoutineHealthThreshold })
	}()

	status, resp := Client.GetPing()
	CheckNoError(t, resp)
	if status != "OK" {
		t.Fatal("should return OK")
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.GoroutineHealthThreshold = 10 })
	status, resp = th.SystemAdminClient.GetPing()
	CheckInternalErrorStatus(t, resp)
	if status != "unhealthy" {
		t.Fatal("should return unhealthy")
	}
}

func TestGetConfig(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	_, resp := Client.GetConfig()
	CheckForbiddenStatus(t, resp)

	cfg, resp := th.SystemAdminClient.GetConfig()
	CheckNoError(t, resp)

	if len(cfg.TeamSettings.SiteName) == 0 {
		t.Fatal()
	}

	if *cfg.LdapSettings.BindPassword != model.FAKE_SETTING && len(*cfg.LdapSettings.BindPassword) != 0 {
		t.Fatal("did not sanitize properly")
	}
	if *cfg.FileSettings.PublicLinkSalt != model.FAKE_SETTING {
		t.Fatal("did not sanitize properly")
	}
	if cfg.FileSettings.AmazonS3SecretAccessKey != model.FAKE_SETTING && len(cfg.FileSettings.AmazonS3SecretAccessKey) != 0 {
		t.Fatal("did not sanitize properly")
	}
	if cfg.EmailSettings.InviteSalt != model.FAKE_SETTING {
		t.Fatal("did not sanitize properly")
	}
	if cfg.EmailSettings.SMTPPassword != model.FAKE_SETTING && len(cfg.EmailSettings.SMTPPassword) != 0 {
		t.Fatal("did not sanitize properly")
	}
	if cfg.GitLabSettings.Secret != model.FAKE_SETTING && len(cfg.GitLabSettings.Secret) != 0 {
		t.Fatal("did not sanitize properly")
	}
	if *cfg.SqlSettings.DataSource != model.FAKE_SETTING {
		t.Fatal("did not sanitize properly")
	}
	if cfg.SqlSettings.AtRestEncryptKey != model.FAKE_SETTING {
		t.Fatal("did not sanitize properly")
	}
	if !strings.Contains(strings.Join(cfg.SqlSettings.DataSourceReplicas, " "), model.FAKE_SETTING) && len(cfg.SqlSettings.DataSourceReplicas) != 0 {
		t.Fatal("did not sanitize properly")
	}
	if !strings.Contains(strings.Join(cfg.SqlSettings.DataSourceSearchReplicas, " "), model.FAKE_SETTING) && len(cfg.SqlSettings.DataSourceSearchReplicas) != 0 {
		t.Fatal("did not sanitize properly")
	}
}

func TestReloadConfig(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	flag, resp := Client.ReloadConfig()
	CheckForbiddenStatus(t, resp)
	if flag {
		t.Fatal("should not Reload the config due no permission.")
	}

	flag, resp = th.SystemAdminClient.ReloadConfig()
	CheckNoError(t, resp)
	if !flag {
		t.Fatal("should Reload the config")
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.MaxUsersPerTeam = 50 })
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.EnableOpenServer = true })
}

func TestUpdateConfig(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	cfg, resp := th.SystemAdminClient.GetConfig()
	CheckNoError(t, resp)

	_, resp = Client.UpdateConfig(cfg)
	CheckForbiddenStatus(t, resp)

	SiteName := th.App.Config().TeamSettings.SiteName

	cfg.TeamSettings.SiteName = "MyFancyName"
	cfg, resp = th.SystemAdminClient.UpdateConfig(cfg)
	CheckNoError(t, resp)

	if len(cfg.TeamSettings.SiteName) == 0 {
		t.Fatal()
	} else {
		if cfg.TeamSettings.SiteName != "MyFancyName" {
			t.Log("It should update the SiteName")
			t.Fatal()
		}
	}

	//Revert the change
	cfg.TeamSettings.SiteName = SiteName
	cfg, resp = th.SystemAdminClient.UpdateConfig(cfg)
	CheckNoError(t, resp)

	if len(cfg.TeamSettings.SiteName) == 0 {
		t.Fatal()
	} else {
		if cfg.TeamSettings.SiteName != SiteName {
			t.Log("It should update the SiteName")
			t.Fatal()
		}
	}

	t.Run("Should not be able to modify PluginSettings.EnableUploads", func(t *testing.T) {
		oldEnableUploads := *th.App.GetConfig().PluginSettings.EnableUploads
		*cfg.PluginSettings.EnableUploads = !oldEnableUploads

		cfg, resp = th.SystemAdminClient.UpdateConfig(cfg)
		CheckNoError(t, resp)
		assert.Equal(t, oldEnableUploads, *cfg.PluginSettings.EnableUploads)
		assert.Equal(t, oldEnableUploads, *th.App.GetConfig().PluginSettings.EnableUploads)

		cfg.PluginSettings.EnableUploads = nil
		cfg, resp = th.SystemAdminClient.UpdateConfig(cfg)
		CheckNoError(t, resp)
		assert.Equal(t, oldEnableUploads, *cfg.PluginSettings.EnableUploads)
		assert.Equal(t, oldEnableUploads, *th.App.GetConfig().PluginSettings.EnableUploads)
	})
}

func TestGetOldClientConfig(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	config, resp := Client.GetOldClientConfig("")
	CheckNoError(t, resp)

	if len(config["Version"]) == 0 {
		t.Fatal("config not returned correctly")
	}

	Client.Logout()

	_, resp = Client.GetOldClientConfig("")
	CheckNoError(t, resp)

	if _, err := Client.DoApiGet("/config/client", ""); err == nil || err.StatusCode != http.StatusNotImplemented {
		t.Fatal("should have errored with 501")
	}

	if _, err := Client.DoApiGet("/config/client?format=junk", ""); err == nil || err.StatusCode != http.StatusBadRequest {
		t.Fatal("should have errored with 400")
	}
}

func TestGetOldClientLicense(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	license, resp := Client.GetOldClientLicense("")
	CheckNoError(t, resp)

	if len(license["IsLicensed"]) == 0 {
		t.Fatal("license not returned correctly")
	}

	Client.Logout()

	_, resp = Client.GetOldClientLicense("")
	CheckNoError(t, resp)

	if _, err := Client.DoApiGet("/license/client", ""); err == nil || err.StatusCode != http.StatusNotImplemented {
		t.Fatal("should have errored with 501")
	}

	if _, err := Client.DoApiGet("/license/client?format=junk", ""); err == nil || err.StatusCode != http.StatusBadRequest {
		t.Fatal("should have errored with 400")
	}

	license, resp = th.SystemAdminClient.GetOldClientLicense("")
	CheckNoError(t, resp)

	if len(license["IsLicensed"]) == 0 {
		t.Fatal("license not returned correctly")
	}
}

func TestGetAudits(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	audits, resp := th.SystemAdminClient.GetAudits(0, 100, "")
	CheckNoError(t, resp)

	if len(audits) == 0 {
		t.Fatal("should not be empty")
	}

	audits, resp = th.SystemAdminClient.GetAudits(0, 1, "")
	CheckNoError(t, resp)

	if len(audits) != 1 {
		t.Fatal("should only be 1")
	}

	audits, resp = th.SystemAdminClient.GetAudits(1, 1, "")
	CheckNoError(t, resp)

	if len(audits) != 1 {
		t.Fatal("should only be 1")
	}

	_, resp = th.SystemAdminClient.GetAudits(-1, -1, "")
	CheckNoError(t, resp)

	_, resp = Client.GetAudits(0, 100, "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetAudits(0, 100, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestEmailTest(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	config := model.Config{
		EmailSettings: model.EmailSettings{
			SMTPServer: "",
			SMTPPort:   "",
		},
	}

	_, resp := Client.TestEmail(&config)
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.TestEmail(&config)
	CheckErrorMessage(t, resp, "api.admin.test_email.missing_server")
	CheckBadRequestStatus(t, resp)

	inbucket_host := os.Getenv("CI_HOST")
	if inbucket_host == "" {
		inbucket_host = "dockerhost"
	}

	inbucket_port := os.Getenv("CI_INBUCKET_PORT")
	if inbucket_port == "" {
		inbucket_port = "9000"
	}

	config.EmailSettings.SMTPServer = inbucket_host
	config.EmailSettings.SMTPPort = inbucket_port
	_, resp = th.SystemAdminClient.TestEmail(&config)
	CheckOKStatus(t, resp)
}

func TestDatabaseRecycle(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	_, resp := Client.DatabaseRecycle()
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.DatabaseRecycle()
	CheckNoError(t, resp)
}

func TestInvalidateCaches(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	flag, resp := Client.InvalidateCaches()
	CheckForbiddenStatus(t, resp)
	if flag {
		t.Fatal("should not clean the cache due no permission.")
	}

	flag, resp = th.SystemAdminClient.InvalidateCaches()
	CheckNoError(t, resp)
	if !flag {
		t.Fatal("should clean the cache")
	}
}

func TestGetLogs(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	for i := 0; i < 20; i++ {
		l4g.Info(i)
	}

	logs, resp := th.SystemAdminClient.GetLogs(0, 10)
	CheckNoError(t, resp)

	if len(logs) != 10 {
		t.Log(len(logs))
		t.Fatal("wrong length")
	}

	logs, resp = th.SystemAdminClient.GetLogs(1, 10)
	CheckNoError(t, resp)

	if len(logs) != 10 {
		t.Log(len(logs))
		t.Fatal("wrong length")
	}

	logs, resp = th.SystemAdminClient.GetLogs(-1, -1)
	CheckNoError(t, resp)

	if len(logs) == 0 {
		t.Fatal("should not be empty")
	}

	_, resp = Client.GetLogs(0, 10)
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetLogs(0, 10)
	CheckUnauthorizedStatus(t, resp)
}

func TestPostLog(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	enableDev := *th.App.Config().ServiceSettings.EnableDeveloper
	defer func() {
		*th.App.Config().ServiceSettings.EnableDeveloper = enableDev
	}()
	*th.App.Config().ServiceSettings.EnableDeveloper = true

	message := make(map[string]string)
	message["level"] = "ERROR"
	message["message"] = "this is a test"

	_, resp := Client.PostLog(message)
	CheckNoError(t, resp)

	Client.Logout()

	_, resp = Client.PostLog(message)
	CheckNoError(t, resp)

	*th.App.Config().ServiceSettings.EnableDeveloper = false

	_, resp = Client.PostLog(message)
	CheckForbiddenStatus(t, resp)

	logMessage, resp := th.SystemAdminClient.PostLog(message)
	CheckNoError(t, resp)
	if len(logMessage) == 0 {
		t.Fatal("should return the log message")
	}
}

func TestUploadLicenseFile(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	ok, resp := Client.UploadLicenseFile([]byte{})
	CheckForbiddenStatus(t, resp)
	if ok {
		t.Fatal("should fail")
	}

	ok, resp = th.SystemAdminClient.UploadLicenseFile([]byte{})
	CheckBadRequestStatus(t, resp)
	if ok {
		t.Fatal("should fail")
	}
}

func TestRemoveLicenseFile(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	ok, resp := Client.RemoveLicenseFile()
	CheckForbiddenStatus(t, resp)
	if ok {
		t.Fatal("should fail")
	}

	ok, resp = th.SystemAdminClient.RemoveLicenseFile()
	CheckNoError(t, resp)
	if !ok {
		t.Fatal("should pass")
	}
}

func TestGetAnalyticsOld(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	rows, resp := Client.GetAnalyticsOld("", "")
	CheckForbiddenStatus(t, resp)
	if rows != nil {
		t.Fatal("should be nil")
	}

	rows, resp = th.SystemAdminClient.GetAnalyticsOld("", "")
	CheckNoError(t, resp)

	found := false
	for _, row := range rows {
		if row.Name == "unique_user_count" {
			found = true
		}
	}

	if !found {
		t.Fatal("should return unique user count")
	}

	_, resp = th.SystemAdminClient.GetAnalyticsOld("post_counts_day", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetAnalyticsOld("user_counts_with_posts_day", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetAnalyticsOld("extra_counts", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetAnalyticsOld("", th.BasicTeam.Id)
	CheckNoError(t, resp)

	Client.Logout()
	_, resp = Client.GetAnalyticsOld("", th.BasicTeam.Id)
	CheckUnauthorizedStatus(t, resp)
}

func TestS3TestConnection(t *testing.T) {
	th := Setup().InitBasic().InitSystemAdmin()
	defer th.TearDown()
	Client := th.Client

	s3Host := os.Getenv("CI_HOST")
	if s3Host == "" {
		s3Host = "dockerhost"
	}

	s3Port := os.Getenv("CI_MINIO_PORT")
	if s3Port == "" {
		s3Port = "9001"
	}

	s3Endpoint := fmt.Sprintf("%s:%s", s3Host, s3Port)
	config := model.Config{
		FileSettings: model.FileSettings{
			DriverName:              model.NewString(model.IMAGE_DRIVER_S3),
			AmazonS3AccessKeyId:     model.MINIO_ACCESS_KEY,
			AmazonS3SecretAccessKey: model.MINIO_SECRET_KEY,
			AmazonS3Bucket:          "",
			AmazonS3Endpoint:        s3Endpoint,
			AmazonS3SSL:             model.NewBool(false),
		},
	}

	_, resp := Client.TestS3Connection(&config)
	CheckForbiddenStatus(t, resp)

	_, resp = th.SystemAdminClient.TestS3Connection(&config)
	CheckBadRequestStatus(t, resp)
	if resp.Error.Message != "S3 Bucket is required" {
		t.Fatal("should return error - missing s3 bucket")
	}

	config.FileSettings.AmazonS3Bucket = model.MINIO_BUCKET
	config.FileSettings.AmazonS3Region = "us-east-1"
	_, resp = th.SystemAdminClient.TestS3Connection(&config)
	CheckOKStatus(t, resp)

	config.FileSettings.AmazonS3Region = ""
	_, resp = th.SystemAdminClient.TestS3Connection(&config)
	CheckOKStatus(t, resp)

	config.FileSettings.AmazonS3Bucket = "Wrong_bucket"
	_, resp = th.SystemAdminClient.TestS3Connection(&config)
	CheckInternalErrorStatus(t, resp)
	if resp.Error.Message != "Error checking if bucket exists." {
		t.Fatal("should return error ")
	}
}
