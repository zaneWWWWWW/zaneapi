package dto

type UpstreamDTO struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name" binding:"required"`
	BaseURL  string `json:"base_url" binding:"required"`
	Endpoint string `json:"endpoint"`
}

type UpstreamRequest struct {
	ChannelIDs []int64       `json:"channel_ids"`
	Upstreams  []UpstreamDTO `json:"upstreams"`
	Timeout    int           `json:"timeout"`
}

// TestResult 上游测试连通性结果
type TestResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// DifferenceItem 差异项
// Current 为本地值，可能为 nil
// Upstreams 为各渠道的上游值，具体数值 / "same" / nil

type DifferenceItem struct {
	Current    interface{}            `json:"current"`
	Upstreams  map[string]interface{} `json:"upstreams"`
	Confidence map[string]bool        `json:"confidence"`
}

type SyncableChannel struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Status  int    `json:"status"`
	Type    int    `json:"type"`
}

type PriceAppleCatalogItem struct {
	ModelName string `json:"model_name"`
	Provider  string `json:"provider,omitempty"`
	Enabled   bool   `json:"enabled"`
}

type CatalogDifference struct {
	ModelName string `json:"model_name"`
	Provider  string `json:"provider,omitempty"`
	Enabled   bool   `json:"enabled"`
	Current   *struct {
		Exists   bool   `json:"exists"`
		Provider string `json:"provider,omitempty"`
		Enabled  bool   `json:"enabled"`
	} `json:"current,omitempty"`
}

type ToolPriceDifference struct {
	Key      string   `json:"key"`
	Current  *float64 `json:"current"`
	Upstream float64  `json:"upstream"`
}

type ApplyPriceAppleExtrasRequest struct {
	Catalog    []PriceAppleCatalogItem `json:"catalog"`
	ToolPrices map[string]float64      `json:"tool_prices"`
}
