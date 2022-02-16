package main

import (
	"flag"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
	"jenkins_exporter/collector"
	stdlog "log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

type handler struct {
	unfilteredHandler       http.Handler
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	maxRequests             int
	logger                  log.Logger
}

func newHandler(includeExporterMetrics bool, maxRequests int, logger log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		maxRequests:             maxRequests,
		logger:                  logger,
	}

	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}

	if innerHandler, err := h.innerHandler(); err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	level.Debug(h.logger).Log("msg", "collect query:", "filters", filters)

	if len(filters) == 0 {
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}

	filteredHandler, err := h.innerHandler(filters...)
	if err != nil {
		level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	filteredHandler.ServeHTTP(w, r)

}

func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	nc, err := collector.NewNodeCollector(h.logger, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	if len(filters) == 0 {
		level.Info(h.logger).Log("msg", "Enabled collectors")
		collectors := []string{}
		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, c := range collectors {
			level.Info(h.logger).Log("collector", c)
		}
	}

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("jenkins_exporter"))
	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn,t register node collectorL %s", err)
	}

	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: h.maxRequests,
			Registry:            h.exporterMetricsRegistry,
		},
	)
	if h.includeExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil

}

func getEnvStr(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if envVal, ok := os.LookupEnv(key); ok {
		if val, err := strconv.ParseBool(envVal); err != nil {
			return val
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if envVal, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(envVal); err != nil {
			return val
		}
	}
	return defaultVal
}

func main() {

	collector.GetBuildHistoryInfo()

	var (
		jenkinsAddr = flag.String("jenkins.addr", getEnvStr("JENKINS_ADDR", ":58080"), "jenkins addr")
		//jenkinsUser = flag.String("jenkins.user", getEnvStr("JENKINS_USER", "admin"), "jenkins user")
		//jenkinsPassword = flag.String("jenkins.password", getEnvStr("JENKINS_PASSWORD", "123456"), "jenkins password")
		//jenkinsPort = flag.String("jenkins.port", getEnvStr("JENKINS_PORT", "58080"), "jenkins addr")
		metricsPath       = flag.String("metrics.path", getEnvStr("METRICS_PATH", "/metrics"), "metrics path")
		maxRequests       = flag.Int("max.requests", getEnvInt("MAX_REQUESTS", 40), "metrics requests")
		configFile        = flag.String("web.config", getEnvStr("WEB.CONFIG", ""), "[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.")
		disableCollectors = flag.Bool("disable.collectors", getEnvBool("DISABLE.COLLECTORS", false), "disable collectors")
	)
	flag.Parse()

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	if *disableCollectors {
		collector.DisableDefaultCollectors()
	}

	http.Handle(*metricsPath, newHandler(!*disableCollectors, *maxRequests, logger))

	server := &http.Server{Addr: *jenkinsAddr}
	if err := web.ListenAndServe(server, *configFile, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

}
