package collector

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 15:09
 */
type BuildStruct struct {
	Class  string   `json:"_class"`
	Builds []Builds `json:"builds"`
}
type Builds struct {
	Class  string `json:"_class"`
	Number int    `json:"number"`
	URL    string `json:"url"`
}
