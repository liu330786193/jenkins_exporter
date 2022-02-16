package collector

/**
 * @description: TODO
 * @author lyl
 * @date 2022/2/16 15:35
 */
type BuildDetail struct {
	Class           string    `json:"_class"`
	Actions         []Actions `json:"actions"`
	FullDisplayName string    `json:"fullDisplayName"`
	Timestamp       float64   `json:"timestamp"`
}
type Causes struct {
	Class    string `json:"_class"`
	UserName string `json:"userName"`
}
type Branch struct {
	Name string `json:"name"`
}
type LastBuiltRevision struct {
	Branch []Branch `json:"branch"`
}
type Actions struct {
	Class             string            `json:"_class,omitempty"`
	Causes            []Causes          `json:"causes,omitempty"`
	LastBuiltRevision LastBuiltRevision `json:"lastBuiltRevision,omitempty"`
}
