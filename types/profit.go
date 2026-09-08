package types

// ProfitBasis preserves the price before the user's group multiplier. It is
// independent of the amount charged, including when the user is charged zero.
type ProfitBasis struct {
	BaseQuota     float64  `json:"base_quota"`
	GroupRatio    float64  `json:"group_ratio"`
	UpstreamRatio *float64 `json:"upstream_ratio"`
	RevenueQuota  *int     `json:"-"`
}
