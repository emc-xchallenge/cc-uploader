package main

import (
	"crypto/tls"
	"flag"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cloudfoundry-incubator/cc-uploader/ccclient"
	"github.com/cloudfoundry-incubator/cc-uploader/handlers"
	"github.com/cloudfoundry-incubator/cf-debug-server"
	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry/dropsonde"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var serverAddress = flag.String(
	"address",
	"0.0.0.0:9090",
	"Specifies the address to bind to",
)

var skipCertVerify = flag.Bool(
	"skipCertVerify",
	false,
	"Skip SSL certificate verification",
)

var ccJobPollingInterval = flag.Duration(
	"ccJobPollingInterval",
	1*time.Second,
	"the interval between job polling requests",
)

var communicationTimeout = flag.Duration(
	"communicationTimeout",
	30*time.Second,
	"Timeout applied to all HTTP requests.",
)

const (
	ccUploadDialTimeout         = 10 * time.Second
	ccUploadKeepAlive           = 30 * time.Second
	ccUploadTLSHandshakeTimeout = 10 * time.Second
	dropsondeDestination        = "localhost:3457"
	dropsondeOrigin             = "cc_uploader"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cf_debug_server.AddFlags(flag.CommandLine)
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	cf_http.Initialize(*communicationTimeout)

	logger, reconfigurableSink := cf_lager.New("cc-uploader")

	initializeDropsonde(logger)

	members := grouper.Members{
		{"cc-uploader", initializeServer(logger)},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))
	logger.Info("ready")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func initializeServer(logger lager.Logger) ifrit.Runner {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   ccUploadDialTimeout,
			KeepAlive: ccUploadKeepAlive,
		}).Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *skipCertVerify,
		},
		TLSHandshakeTimeout: ccUploadTLSHandshakeTimeout,
	}

	pollerHttpClient := cf_http.NewClient()
	pollerHttpClient.Transport = transport

	uploader := ccclient.NewUploader(logger, &http.Client{Transport: transport})
	poller := ccclient.NewPoller(logger, pollerHttpClient, *ccJobPollingInterval)

	ccUploaderHandler, err := handlers.New(uploader, poller, logger)
	if err != nil {
		logger.Error("router-building-failed", err)
		os.Exit(1)
	}

	return http_server.New(*serverAddress, ccUploaderHandler)
}
