package v1

type ResourcesSpec struct {
	RequestsSpec RequestsSpec `json:"requests,omitempty"`
	LimitsSpec   LimitsSpec   `json:"limits,omitempty"`
}

type RequestsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type LimitsSpec struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}
