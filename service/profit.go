package service

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
)

func ProfitBasis(info *relaycommon.RelayInfo, baseQuota float64) *types.ProfitBasis {
	if info.ChannelMeta == nil || info.UpstreamRatio == nil {
		return nil
	}
	return &types.ProfitBasis{BaseQuota: baseQuota, GroupRatio: info.PriceData.GroupRatioInfo.GroupRatio, UpstreamRatio: info.UpstreamRatio, RevenueQuota: info.ProfitRevenueQuota}
}
