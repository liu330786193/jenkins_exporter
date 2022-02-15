package collector

import (
	"errors"
	"fmt"
	"github.com/go-kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/15 14:30
 */

var (
	factories                   = make(map[string]func(logger log.Logger) (Collector, error))
	initiatedCollectorsMtx      = sync.Mutex{}
	initiatedCollectors         = make(map[string]Collector)
	collectorState              = make(map[string]*bool)
	forcedCollectors            = map[string]bool{}
)

const namespace = "jenkins"

const (
	defaultEnabled  = true
	defaultDisabled = false
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
			"jenkins_exporter collector_duration_seconds",
			[]string{"collector"},
			nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"jenkins_exporter collector_success",
		[]string{"collector"},
		nil,
	)
)

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type typedDesc struct {
	desc        *prometheus.Desc
	valueType   prometheus.ValueType
}
func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}


type NodeCollector struct {
	Collectors map[string]Collector
	logger     log.Logger
}

func DisableDefaultCollectors() {
	for c := range collectorState {
		if _, ok := forcedCollectors[c]; !ok {
			*collectorState[c] = false
		}
	}
}

func (n NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (n NodeCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors{
		go func(name string, c Collector) {
			execute(name, c, ch, n.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func registerCollector(collector string, isDefaultEnabled bool, factory func(logger log.Logger) (Collector, error)) {

	//var helpDefaultState string
	//if isDefaultEnabled {
	//	helpDefaultState = "enabled"
	//} else {
	//	helpDefaultState = "disabled"
	//}

	//flagName := fmt.Sprintf("collector.%s", collector)
	//flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	//defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	//flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(CollectorFlagAction(collector)).Bool()
	collectorState[collector] = &isDefaultEnabled
	factories[collector] = factory
}

func CollectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger log.Logger)  {

	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			level.Debug(logger).Log("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		level.Debug(logger).Log("msg", "collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

func IsNoDataError(err error) bool {
	return err == errors.New("collector returned no data")
}

func NewNodeCollector(logger log.Logger, filters ...string) (*NodeCollector, error) {
	f := make(map[string]bool)
	for _, filter := range filters{
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, fmt.Errorf("missing collector: %s", filter)
		}
		if !*enabled {
			return nil, fmt.Errorf("disabled collector: %s", filter)
		}
		f[filter] = true
	}

	collectors := make(map[string]Collector)
	initiatedCollectorsMtx.Lock()
	defer initiatedCollectorsMtx.Unlock()

	for key, enabled := range collectorState{
		if !*enabled || (len(f) > 0 && !f[key]) {
			continue
		}
		if collector, ok := initiatedCollectors[key]; ok {
			collectors[key] = collector
		} else {
			collector, err := factories[key](log.With(logger, "collector", key))
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
			initiatedCollectors[key] = collector
		}
	}
	return &NodeCollector{Collectors: collectors, logger: logger}, nil
}
















