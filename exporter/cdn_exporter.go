package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qiniu/go-sdk/v7/auth"
	"qiniu-exporter/retrieve"
)

const cdnNameSpace = "qiniu"


type cdnExporter struct {
	domainList               *[]string
	credential               *auth.Credentials
	rangeTime                int64
	delayTime                int64
	granularity              string
	cdnHitRate               *prometheus.Desc
	cdnFluxHitRate           *prometheus.Desc
	cdnBandWidth             *prometheus.Desc
	cdnStatusRate            *prometheus.Desc
}


func NewCdnExporter(domainList *[]string, credential *auth.Credentials, rangeTime int64, delayTime int64, granularity string) *cdnExporter {
	return &cdnExporter{
		domainList: domainList,
		credential:      credential,
		rangeTime:  rangeTime,
		delayTime:  delayTime,
		granularity: granularity,

		cdnHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "hit_rate"),
			"cdn请求命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnFluxHitRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "flux_hit_rate"),
			"cdn字节命中率(%)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnBandWidth: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "bandwidth"),
			"cdn总带宽(Mbps)",
			[]string{
				"instanceId",
			},
			nil,
		),
		cdnStatusRate: prometheus.NewDesc(
			prometheus.BuildFQName(cdnNameSpace, "cdn", "status_rate"),
			"cdn状态码概率(%)",
			[]string{
				"instanceId",
				"status",
			},
			nil,
		),
	}
}


func (e *cdnExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cdnHitRate
	ch <- e.cdnFluxHitRate
	ch <- e.cdnBandWidth
	ch <- e.cdnStatusRate
}

func (e *cdnExporter) Collect(ch chan<- prometheus.Metric) {
	for _, domain := range *e.domainList {

		bandWidth := retrieve.GetBandWidth(e.credential, domain, e.rangeTime, e.delayTime, e.granularity)
		hitRate, fluxHitRate := retrieve.GetHitMiss(e.credential, domain, e.rangeTime, e.delayTime, e.granularity)
		statusProportion := retrieve.GetStatusCode(e.credential, domain, e.rangeTime, e.delayTime, e.granularity)

		ch <- prometheus.MustNewConstMetric(
			e.cdnBandWidth,
			prometheus.GaugeValue,
			bandWidth / 1024 / 1024,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnHitRate,
			prometheus.GaugeValue,
			hitRate * 100,
			domain,
		)
		ch <- prometheus.MustNewConstMetric(
			e.cdnFluxHitRate,
			prometheus.GaugeValue,
			fluxHitRate * 100,
			domain,
		)
		for status, proportion := range statusProportion {
			ch <- prometheus.MustNewConstMetric(
				e.cdnStatusRate,
				prometheus.GaugeValue,
				proportion * 100,
				domain,
				status,
			)
		}
	}
}