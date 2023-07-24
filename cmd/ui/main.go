// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	jconfig "github.com/uber/jaeger-client-go/config"
	"github.com/ultravioletrs/mainflux-ui/ui"
	"github.com/ultravioletrs/mainflux-ui/ui/api"
)

const (
	defLogLevel        = "info"
	defClientTLS       = "false"
	defCACerts         = ""
	defPort            = "9090"
	defRedirectURL     = "http://localhost:9090/"
	defJaegerURL       = ""
	defHTTPAdapterPort = "8008"
	defReaderPort      = ""
	defThingsPort      = "9000"
	defUsersPort       = "9002"
	defTLSVerification = "false"
	defBaseURL         = "http://localhost"
	defInstanceID      = ""

	envLogLevel        = "MF_GUI_LOG_LEVEL"
	envClientTLS       = "MF_GUI_CLIENT_TLS"
	envCACerts         = "MF_GUI_CA_CERTS"
	envPort            = "MF_GUI_PORT"
	envRedirectURL     = "MF_GUI_REDIRECT_URL"
	envJaegerURL       = "MF_JAEGER_URL"
	envHTTPAdapterPort = "MF_HTTP_ADAPTER_PORT"
	envReaderPort      = "MF_READER_PORT"
	envThingsPort      = "MF_THINGS_HTTP_PORT"
	envUsersPort       = "MF_USERS_HTTP_PORT"
	envTLSVerification = "MF_VERIFICATION_TLS"
	envBaseURL         = "MF_SDK_BASE_URL"
	envInstanceID      = "MF_UI_INSTANCE_ID"
)

type config struct {
	baseURL     string
	logLevel    string
	port        string
	redirectURL string
	clientTLS   bool
	caCerts     string
	jaegerURL   string
	instanceID  string
	sdkConfig   sdk.Config
}

func main() {
	cfg := loadConfig()

	instanceID := cfg.instanceID
	if instanceID == "" {
		if uuid, err := uuid.New().ID(); err != nil {
			log.Fatalf("Failed to generate instanceID: %s", err)
		} else {
			instanceID = uuid
		}
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tracer, closer := initJaeger("ui", cfg.jaegerURL, logger)
	defer closer.Close()

	sdk := sdk.NewSDK(cfg.sdkConfig)

	svc := ui.New(sdk)

	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "ui",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "ui",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.port)
		logger.Info(fmt.Sprintf("GUI service started on port %s", cfg.port))
		errs <- http.ListenAndServe(p, api.MakeHandler(svc, cfg.redirectURL, tracer, instanceID))
	}()

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("GUI service terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}
	mfTLS, err := strconv.ParseBool(mainflux.Env(envTLSVerification, defTLSVerification))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envTLSVerification)
	}
	baseURL := mainflux.Env(envBaseURL, defBaseURL)
	return config{
		baseURL:     baseURL,
		logLevel:    mainflux.Env(envLogLevel, defLogLevel),
		port:        mainflux.Env(envPort, defPort),
		redirectURL: mainflux.Env(envRedirectURL, defRedirectURL),
		clientTLS:   tls,
		caCerts:     mainflux.Env(envCACerts, defCACerts),
		jaegerURL:   mainflux.Env(envJaegerURL, defJaegerURL),
		instanceID:  mainflux.Env(envInstanceID, defInstanceID),
		sdkConfig: sdk.Config{
			HTTPAdapterURL:  fmt.Sprintf("%s:%s", baseURL, mainflux.Env(envHTTPAdapterPort, defHTTPAdapterPort)),
			ReaderURL:       fmt.Sprintf("%s:%s", baseURL, mainflux.Env(envReaderPort, defReaderPort)),
			ThingsURL:       fmt.Sprintf("%s:%s", baseURL, mainflux.Env(envThingsPort, defThingsPort)),
			UsersURL:        fmt.Sprintf("%s:%s", baseURL, mainflux.Env(envUsersPort, defUsersPort)),
			MsgContentType:  sdk.ContentType(string(sdk.CTJSONSenML)),
			TLSVerification: mfTLS,
		},
	}
}

func initJaeger(svcName, url string, logger logger.Logger) (opentracing.Tracer, io.Closer) {
	if url == "" {
		return opentracing.NoopTracer{}, ioutil.NopCloser(nil)
	}

	tracer, closer, err := jconfig.Configuration{
		ServiceName: svcName,
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jconfig.ReporterConfig{
			LocalAgentHostPort: url,
			LogSpans:           true,
		},
	}.NewTracer()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger client: %s", err))
		os.Exit(1)
	}

	return tracer, closer
}