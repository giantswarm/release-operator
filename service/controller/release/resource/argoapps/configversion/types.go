package configversion

type catalogIndex struct {
	Entries map[string][]catalogIndexEntry `json:"entries"`
}

type catalogIndexEntry struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Version     string            `json:"version"`
}
