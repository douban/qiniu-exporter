package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qiniu/go-sdk/v7/auth"
	"log"
	"net"
	"net/http"
	"os"
	"qiniu-exporter/exporter"
	"qiniu-exporter/retrieve"
	"strconv"
)


func main() {
	accessKey := flag.String("access_key", os.Getenv("QINIU_ACCESS_KEY"), "Qiniu cloud access key")
	secretKey := flag.String("secret_key", os.Getenv("QINIU_SECRET_KEY"), "Qiniu cloud secret key")
	host := flag.String("host", "0.0.0.0", "服务监听地址")
	port := flag.Int("port", 9270, "服务监听端口")
	delayTime := flag.Int64("delayTime", 300, "时间偏移量, 结束时间=now-delay_seconds")
	rangeTime := flag.Int64("rangeTime", 1800, "选取时间范围, 开始时间=now-range_seconds, 结束时间=now")
	granularity := flag.String("granularity", "5min", "粒度，可选项为 5min、1hour、1day")
	metricsPath := flag.String("metricsPath", "/metrics", "默认的metrics路径")
	// auth 认证
	credential := auth.New(*accessKey, *secretKey)
	domainList := retrieve.GetDomains(credential)

	cdn := exporter.NewCdnExporter(&domainList, credential, *rangeTime, *delayTime, *granularity)
	prometheus.MustRegister(cdn)
	listenAddress := net.JoinHostPort(*host, strconv.Itoa(*port))
	log.Println(listenAddress)
	log.Println("Running on", listenAddress)
	http.Handle(*metricsPath, promhttp.Handler()) //注册

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
           <head><title>Qiniu CDN Exporter</title></head>
           <body>
           <h1>Qiniu cdn exporter</h1>
           <p><a href='` + *metricsPath + `'>Metrics</a></p>
           </body>
           </html>`))
	})

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}