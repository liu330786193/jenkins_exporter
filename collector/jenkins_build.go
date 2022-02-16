package collector

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"os"
)

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 10:03
 */

type lastBuildCollector struct {
}

const subsystem = "last_build"

func init() {
	registerCollector(subsystem, defaultEnabled, LastBuildCollector)
}

func LastBuildCollector(logger log.Logger) (Collector, error) {
	return &lastBuildCollector{}, nil
}

func GetBuildInfo() *BuildInfo {

	url := "http://192.168.2.100:8080/api/json?pretty=true&tree=jobs[fullName,url,lastBuild[fullName,number,timestamp,duration,actions[queuingDurationMillis,totalDurationMillis,skipCount,failCount,totalCount,passCount]],lastFailedBuild[fullName,number,timestamp,duration,actions[queuingDurationMillis,totalDurationMillis,skipCount,failCount,totalCount,passCount]],lastSuccessfulBuild[fullName,number,timestamp,duration,actions[queuingDurationMillis,totalDurationMillis,skipCount,failCount,totalCount,passCount]],lastUnstableBuild[fullName,number,timestamp,duration,actions[queuingDurationMillis,totalDurationMillis,skipCount,failCount,totalCount,passCount]]]"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}
	req.SetBasicAuth("admin", "123456")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	fmt.Println(resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Println("ok")
	}

	var bi BuildInfo
	json.Unmarshal(body, &bi)
	fmt.Println(bi)
	return &bi
}

func (c *lastBuildCollector) Update(ch chan<- prometheus.Metric) error {

	buildInfos := GetBuildInfo()

	for _, job := range buildInfos.Jobs {

		fullName := job.FullName

		metricsMap := map[string]float64{
			"duration":    job.LastBuild.Duration,
			"number":      job.LastBuild.Number,
			"timestamp":   job.LastBuild.Timestamp,
			"fail_count":  job.LastBuild.Actions[3].FailCount,
			"skip_count":  job.LastBuild.Actions[3].SkipCount,
			"total_count": job.LastBuild.Actions[3].TotalCount,
		}

		for key, value := range metricsMap {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, key),
					"jenkins",
					[]string{"project_name"}, nil,
				),
				prometheus.GaugeValue,
				value,
				fullName,
			)
		}

	}
	return nil
}
