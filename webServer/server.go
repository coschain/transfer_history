package webServer

import (
	"context"
	"fmt"
	"github.com/transfer_history/config"
	"github.com/transfer_history/logs"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	getTransferHistoryUrl = "/api/getTransferHistory"
	getTransferHistoryInBlockUrl = "/api/getTransferHistoryByBlock"

	writeTimeOut = 3
	readTimeOut  = 3
)

var (
	syncLock sync.Mutex
	server   *http.Server
)


func StartServer() error {
	var g errgroup.Group
	serverMux := initHandlers()
	server = &http.Server{Handler: serverMux, ReadTimeout: readTimeOut * time.Minute, WriteTimeout: writeTimeOut * time.Minute}
	addr := ":" + config.GetHttpPort()
	listener,err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Fail to listen port, the error is %v \n", err)
		return err
	}
	g.Go(func() error {
		fmt.Println("start http server")
		err = server.Serve(listener)
		if err != nil {
			fmt.Printf("Fail to start the http server, the serever is %v \n", err)
		} else {
			fmt.Println("success start http server")
		}
		return err
	})
	err = g.Wait()
	return err
}


func initHandlers() *http.ServeMux {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc(getTransferHistoryUrl, func(writer http.ResponseWriter, request *http.Request) {
		getTransferHistory(writer, request)
	})
	serverMux.HandleFunc(getTransferHistoryInBlockUrl, func(writer http.ResponseWriter, request *http.Request) {
		getTransferHistoryOfBlock(writer, request)
	})
	return serverMux
}

func StopServer()  {
	if err := server.Shutdown(context.Background());err != nil {
		log := logs.GetLogger()
		log.Errorf("StopServer: fail to stop http server, the error is %v", err)
	}
}