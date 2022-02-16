package collector

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 15:00
 */
type FullName struct {
	Class string `json:"_class"`
	Jobs  []struct {
		Class    string `json:"_class"`
		FullName string `json:"fullName"`
	} `json:"jobs"`
}
