package commands

import (
	"fmt"
	"github.com/coschain/cobra"
	"github.com/prometheus/common/log"
	"os"
	"os/signal"
	"syscall"
	"transfer_history/config"
	"transfer_history/db"
	"transfer_history/logs"
	"transfer_history/webServer"
)

var svEnv string

var StartCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "start",
		Short:     "start exchange transfer history net service",
		Long:      "start exchange transfer history net service,if has arg 'env',will use it for service env",
		ValidArgs: []string{"env"},
		Run:       startNetService,
	}
	cmd.Flags().StringVarP(&svEnv, "env", "e", "pro", "service env (default is pro)")

	return cmd
}

func startNetService(cmd *cobra.Command, args []string)  {
	fmt.Println("start transfer history net service")

	err := config.SetConfigEnv(svEnv)
	if err != nil {
		fmt.Printf("StartService:fail to set env")
		os.Exit(1)
	}
	//load config json file
	err = config.LoadExchangeTransferHistoryConfig("../../transfer_history.json")
	if err != nil {
		fmt.Println("TransferHistoryNetService:fail to load config file ")
		os.Exit(1)
	}

	logger,err := logs.StartLogService()
	if err != nil {
		log.Error("TransferHistoryNetService:fail to start log service")
		os.Exit(1)
	}

	//start db service
	err = db.StartDbService()
	if err != nil {
		logger.Error("StartDbService:fail to start db service")
		os.Exit(1)
	}
	defer db.CloseDbService()
	//start http service
	err = webServer.StartServer()
	if err != nil {
		os.Exit(1)
	}
	defer webServer.StopServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}