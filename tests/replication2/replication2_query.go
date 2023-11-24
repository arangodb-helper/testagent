package replication2

type QueryRequest struct {
	Query     string `json:"query"`
	BatchSize int    `json:"batchSize,omitempty"`
	Count     bool   `json:"count,omitempty"`
}

type CursorResponse struct {
	HasMore bool          `json:"hasMore,omitempty"`
	ID      string        `json:"id,omitempty"`
	Result  []interface{} `json:"result,omitempty"`
}
