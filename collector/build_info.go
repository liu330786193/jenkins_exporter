package collector

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 13:31
 */
type BuildInfo struct {
	Class string `json:"_class"`
	Jobs  []struct {
		Class     string `json:"_class"`
		FullName  string `json:"fullName"`
		LastBuild struct {
			Class   string `json:"_class"`
			Actions []struct {
				Class      string  `json:"_class,omitempty"`
				FailCount  float64 `json:"failCount,omitempty"`
				SkipCount  float64 `json:"skipCount,omitempty"`
				TotalCount float64 `json:"totalCount,omitempty"`
			} `json:"actions"`
			Duration  float64 `json:"duration"`
			Number    float64 `json:"number"`
			Timestamp float64 `json:"timestamp"`
		} `json:"lastBuild"`
		LastFailedBuild struct {
			Class   string `json:"_class"`
			Actions []struct {
				Class      string  `json:"_class,omitempty"`
				FailCount  float64 `json:"failCount,omitempty"`
				SkipCount  float64 `json:"skipCount,omitempty"`
				TotalCount float64 `json:"totalCount,omitempty"`
			} `json:"actions"`
			Duration  float64 `json:"duration"`
			Number    float64 `json:"number"`
			Timestamp float64 `json:"timestamp"`
		} `json:"lastFailedBuild"`
		LastSuccessfulBuild struct {
			Class   string `json:"_class"`
			Actions []struct {
				Class      string  `json:"_class,omitempty"`
				FailCount  float64 `json:"failCount,omitempty"`
				SkipCount  float64 `json:"skipCount,omitempty"`
				TotalCount float64 `json:"totalCount,omitempty"`
			} `json:"actions"`
			Duration  float64 `json:"duration"`
			Number    float64 `json:"number"`
			Timestamp float64 `json:"timestamp"`
		} `json:"lastSuccessfulBuild"`
		URL string `json:"url"`
	} `json:"jobs"`
}
