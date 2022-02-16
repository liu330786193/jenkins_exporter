package collector

import (
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 10:03
 */

type buildHistoryCollector struct {
}

func init() {
	registerCollector("build_history", defaultEnabled, BuildBuildCollector)
}

func BuildBuildCollector(logger log.Logger) (Collector, error) {
	return &buildHistoryCollector{}, nil
}

func GetBuildHistoryInfo() []*BuildDetail {

	url := "http://192.168.2.100:8080/api/json?pretty=true&tree=jobs[fullName]"
	fullNameBytes := getInfo(url)

	var bi FullName
	json.Unmarshal(fullNameBytes, &bi)
	//fmt.Println(bi)

	var buildDetails []*BuildDetail
	for _, job := range bi.Jobs {
		//http://192.168.2.100:8080/job/test/15/api/json?pretty=true&tree=builds[number,url]
		url := "http://192.168.2.100:8080/job/" + job.FullName + "/api/json?pretty=true&tree=builds[number,url]"
		buildStructBytes := getInfo(url)
		var bs BuildStruct
		json.Unmarshal(buildStructBytes, &bs)
		var bsSlices []Builds
		if len(bs.Builds) >= 5 {
			bsSlices = bs.Builds[:5]
		} else {
			bsSlices = bs.Builds
		}
		for _, bsSlice := range bsSlices {
			//http://192.168.2.100:8080/job/test/15/api/json?pretty=true&tree=actions[causes[userName],lastBuiltRevision[branch[name]]]
			url := "http://192.168.2.100:8080/job/" + job.FullName + "/" + strconv.Itoa(bsSlice.Number) + "/api/json?pretty=true&tree=actions[causes[userName],lastBuiltRevision[branch[name]]],fullDisplayName,timestamp"
			buildDetailBytes := getInfo(url)
			var bd BuildDetail
			json.Unmarshal(buildDetailBytes, &bd)
			fmt.Println(bd)
			buildDetails = append(buildDetails, &bd)
		}
	}
	return buildDetails

}

func getInfo(url string) []byte {
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
	if resp.StatusCode == 200 {
		fmt.Println("ok")
	}
	return body
}

func (c *buildHistoryCollector) Update(ch chan<- prometheus.Metric) error {
	const subsystem = "build_history"
	buildDetails := GetBuildHistoryInfo()
	for _, buildDetail := range buildDetails {
		fullName := buildDetail.FullDisplayName
		name := "time"

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, name),
				"jenkins",
				[]string{"project_name", "branch", "user"}, nil,
			),
			prometheus.GaugeValue,
			buildDetail.Timestamp,
			fullName,
			buildDetail.Actions[0].Causes[0].UserName,
			buildDetail.Actions[1].LastBuiltRevision.Branch[0].Name,
		)

	}
	return nil
}
