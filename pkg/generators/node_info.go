package generators

import (
	"github.com/prometheus/client_golang/prometheus"
	"main/pkg/constants"
	fetchersPkg "main/pkg/fetchers"
)

type NodeInfoGenerator struct {
}

func NewNodeInfoGenerator() *NodeInfoGenerator {
	return &NodeInfoGenerator{}
}

func (g *NodeInfoGenerator) Generate(state fetchersPkg.State) []prometheus.Collector {
	nodeInfos, ok := fetchersPkg.StateGet[fetchersPkg.NodeInfoData](state, constants.FetcherNameNodeInfo)
	if !ok {
		return []prometheus.Collector{}
	}

	networkInfoGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsPrefix + "chain_info",
			Help: "Chain info, always 1.",
		},
		[]string{
			"chain",
			"chain_id",
			"cosmos_sdk_version",
			"tendermint_version",
			"app_version",
			"name",
			"app_name",
		},
	)

	for chain, nodeInfo := range nodeInfos.NodeInfos {
		networkInfoGauge.With(prometheus.Labels{
			"chain":              chain,
			"chain_id":           nodeInfo.DefaultNodeInfo.Network,
			"cosmos_sdk_version": nodeInfo.ApplicationVersion.CosmosSDKVersion,
			"tendermint_version": nodeInfo.DefaultNodeInfo.Version,
			"app_version":        nodeInfo.ApplicationVersion.Version,
			"name":               nodeInfo.ApplicationVersion.Name,
			"app_name":           nodeInfo.ApplicationVersion.AppName,
		}).Set(1)
	}

	return []prometheus.Collector{networkInfoGauge}
}
