// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	l4g "github.com/alecthomas/log4go"
	"github.com/SoulDemon/mattermostp/api"
	"github.com/SoulDemon/mattermostp/api4"
	"github.com/SoulDemon/mattermostp/app"
	"github.com/SoulDemon/mattermostp/cmd"
	"github.com/SoulDemon/mattermostp/manualtesting"
	"github.com/SoulDemon/mattermostp/model"
	"github.com/souldemon/mattermostp/utils"
	"github.com/SoulDemon/mattermostp/web"
	"github.com/SoulDemon/mattermostp/wsapi"
	"github.com/spf13/cobra"
)

const (
	SESSIONS_CLEANUP_BATCH_SIZE = 1000
)

var MaxNotificationsPerChannelDefault int64 = 1000000

var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the Mattermost server",
	RunE:         serverCmdF,
	SilenceUsage: true,
}

func init() {
	cmd.RootCmd.AddCommand(serverCmd)
	cmd.RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
	config, err := command.Flags().GetString("config")
	if err != nil {
		return err
	}

	disableConfigWatch, _ := command.Flags().GetBool("disableconfigwatch")

	interruptChan := make(chan os.Signal, 1)
	return runServer(config, disableConfigWatch, interruptChan)
}

func runServer(configFileLocation string, disableConfigWatch bool, interruptChan chan os.Signal) error {
	options := []app.Option{app.ConfigFile(configFileLocation)}
	if disableConfigWatch {
		options = append(options, app.DisableConfigWatch)
	}

	a, err := app.New(options...)
	if err != nil {
		l4g.Critical(err.Error())
		return err
	}
	defer a.Shutdown()

	utils.TestConnection(a.Config())

	pwd, _ := os.Getwd()
	l4g.Info(utils.T("mattermost.current_version"), model.CurrentVersion, model.BuildNumber, model.BuildDate, model.BuildHash, model.BuildHashEnterprise)
	l4g.Info(utils.T("mattermost.entreprise_enabled"), model.BuildEnterpriseReady)
	l4g.Info(utils.T("mattermost.working_dir"), pwd)
	l4g.Info(utils.T("mattermost.config_file"), utils.FindConfigFile(configFileLocation))

	backend, appErr := a.FileBackend()
	if appErr == nil {
		appErr = backend.TestConnection()
	}
	if appErr != nil {
		l4g.Error("Problem with file storage settings: " + appErr.Error())
	}

	if model.BuildEnterpriseReady == "true" {
		a.LoadLicense()
	}

	a.InitPlugins(*a.Config().PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory, nil)
	a.AddConfigListener(func(prevCfg, cfg *model.Config) {
		if *cfg.PluginSettings.Enable {
			a.InitPlugins(*cfg.PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory, nil)
		} else {
			a.ShutDownPlugins()
		}
	})

	serverErr := a.StartServer()
	if serverErr != nil {
		l4g.Critical(serverErr.Error())
		return serverErr
	}

	api4.Init(a, a.Srv.Router, false)
	api3 := api.Init(a, a.Srv.Router)
	wsapi.Init(a, a.Srv.WebSocketRouter)
	web.Init(api3)

	license := a.License()

	if license == nil && len(a.Config().SqlSettings.DataSourceReplicas) > 1 {
		l4g.Warn(utils.T("store.sql.read_replicas_not_licensed.critical"))
		a.UpdateConfig(func(cfg *model.Config) {
			cfg.SqlSettings.DataSourceReplicas = cfg.SqlSettings.DataSourceReplicas[:1]
		})
	}

	if license == nil {
		a.UpdateConfig(func(cfg *model.Config) {
			cfg.TeamSettings.MaxNotificationsPerChannel = &MaxNotificationsPerChannelDefault
		})
	}

	a.ReloadConfig()

	// Enable developer settings if this is a "dev" build
	if model.BuildNumber == "dev" {
		a.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableDeveloper = true })
	}

	resetStatuses(a)

	// If we allow testing then listen for manual testing URL hits
	if a.Config().ServiceSettings.EnableTesting {
		manualtesting.Init(api3)
	}

	a.EnsureDiagnosticId()

	a.Go(func() {
		runSecurityJob(a)
	})
	a.Go(func() {
		runDiagnosticsJob(a)
	})
	a.Go(func() {
		runSessionCleanupJob(a)
	})
	a.Go(func() {
		runTokenCleanupJob(a)
	})
	a.Go(func() {
		runCommandWebhookCleanupJob(a)
	})

	if complianceI := a.Compliance; complianceI != nil {
		complianceI.StartComplianceDailyJob()
	}

	if a.Cluster != nil {
		a.RegisterAllClusterMessageHandlers()
		a.Cluster.StartInterNodeCommunication()
	}

	if a.Metrics != nil {
		a.Metrics.StartServer()
	}

	if a.Elasticsearch != nil {
		a.Go(func() {
			if err := a.Elasticsearch.Start(); err != nil {
				l4g.Error(err.Error())
			}
		})
	}

	if *a.Config().JobSettings.RunJobs {
		a.Jobs.StartWorkers()
	}
	if *a.Config().JobSettings.RunScheduler {
		a.Jobs.StartSchedulers()
	}

	notifyReady()

	// wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	if a.Cluster != nil {
		a.Cluster.StopInterNodeCommunication()
	}

	if a.Metrics != nil {
		a.Metrics.StopServer()
	}

	a.Jobs.StopSchedulers()
	a.Jobs.StopWorkers()

	return nil
}

func runSecurityJob(a *app.App) {
	doSecurity(a)
	model.CreateRecurringTask("Security", func() {
		doSecurity(a)
	}, time.Hour*4)
}

func runDiagnosticsJob(a *app.App) {
	doDiagnostics(a)
	model.CreateRecurringTask("Diagnostics", func() {
		doDiagnostics(a)
	}, time.Hour*24)
}

func runTokenCleanupJob(a *app.App) {
	doTokenCleanup(a)
	model.CreateRecurringTask("Token Cleanup", func() {
		doTokenCleanup(a)
	}, time.Hour*1)
}

func runCommandWebhookCleanupJob(a *app.App) {
	doCommandWebhookCleanup(a)
	model.CreateRecurringTask("Command Hook Cleanup", func() {
		doCommandWebhookCleanup(a)
	}, time.Hour*1)
}

func runSessionCleanupJob(a *app.App) {
	doSessionCleanup(a)
	model.CreateRecurringTask("Session Cleanup", func() {
		doSessionCleanup(a)
	}, time.Hour*24)
}

func resetStatuses(a *app.App) {
	if result := <-a.Srv.Store.Status().ResetAll(); result.Err != nil {
		l4g.Error(utils.T("mattermost.reset_status.error"), result.Err.Error())
	}
}

func doSecurity(a *app.App) {
	a.DoSecurityUpdateCheck()
}

func doDiagnostics(a *app.App) {
	if *a.Config().LogSettings.EnableDiagnostics {
		a.SendDailyDiagnostics()
	}
}

func notifyReady() {
	// If the environment vars provide a systemd notification socket,
	// notify systemd that the server is ready.
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		l4g.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			l4g.Error(err.Error())
		}
	}
}

func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}

func doTokenCleanup(a *app.App) {
	a.Srv.Store.Token().Cleanup()
}

func doCommandWebhookCleanup(a *app.App) {
	a.Srv.Store.CommandWebhook().Cleanup()
}

func doSessionCleanup(a *app.App) {
	a.Srv.Store.Session().Cleanup(model.GetMillis(), SESSIONS_CLEANUP_BATCH_SIZE)
}
