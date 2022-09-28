package retrieve

import (
	"encoding/json"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/sms/bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	FusionHost = "http://fusion.qiniuapi.com"
	QiniuHost = "http://api.qiniu.com"
)

type domainResp struct {
	Domains []struct{
		Name string `json:"name"`
		Type string `json:"type"`
		Cname string `json:"cname"`
		GeoCover string `json:"geo_cover"`
		Platform string `json:"platform"`
		Protocol string `json:"protocol"`
	} `json:"domains"`
}

type bandWidthOfZone struct {
	China    []int64 `json:"china"`
	Oversea  []int64 `json:"oversea"`
}

type bandWidthResp struct {
	Code  int64   `json:"code"`
	Error string  `json:"error"`
	Data  map[string]bandWidthOfZone `json:"data"`
}

type hitMissDetail struct {
	Hit  []int64  `json:"hit"`
	Miss []int64  `json:"miss"`
	TrafficHit  []int64  `json:"trafficHit"`
	TrafficMiss []int64  `json:"trafficMiss"`
}

type hitMissResp struct {
	Code  int64   `json:"code"`
	Error string  `json:"error"`
	Data  hitMissDetail `json:"data"`
}

type codes struct {
	Codes map[string][]int64 `json:"codes"`
}

type statusResp struct {
	Code  int64   `json:"code"`
	Error string  `json:"error"`
	Data  codes `json:"data"`
}

func getRequest(mac *auth.Credentials, path string, query interface{}) (resData []byte, err error) {
	urlStr := fmt.Sprintf("%s%s", QiniuHost, path)
	reqData, _ := json.Marshal(query)
	req, reqErr := http.NewRequest("GET", urlStr, bytes.NewReader(reqData))
	if reqErr != nil {
		log.Fatal("请求参数异常", reqErr)
	}

	accessToken, signErr := mac.SignRequest(req)
	if signErr != nil {
		log.Fatal("获取认证失败", signErr)
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Fatal("请求失败", respErr)
	}
	defer resp.Body.Close()

	resData, ioErr := io.ReadAll(resp.Body)
	if ioErr != nil {
		err = ioErr
		return
	}
	return
}


func postRequest(mac *auth.Credentials, path string, body interface{}) (resData []byte, err error) {
	urlStr := fmt.Sprintf("%s%s", FusionHost, path)
	reqData, _ := json.Marshal(body)
	req, reqErr := http.NewRequest("POST", urlStr, bytes.NewReader(reqData))
	if reqErr != nil {
		log.Fatal("请求参数异常", reqErr)
	}

	accessToken, signErr := mac.SignRequest(req)
	if signErr != nil {
		log.Fatal("获取认证失败", signErr)
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		log.Fatal("请求失败", respErr)
	}
	defer resp.Body.Close()

	resData, ioErr := io.ReadAll(resp.Body)
	if ioErr != nil {
		err = ioErr
		return
	}
	return
}


func GetDomains(credential *auth.Credentials) []string {
	queryBody := map[string]interface{}{
		"Limit": 50,
		"SourceTypes": []string{"domain"},
	}
	resp, err := getRequest(credential, "/domain", queryBody)
	if err != nil {
		log.Fatal("请求失败", err)
	}
	var (
		domains    domainResp
		domainList []string
	)
	errJson := json.Unmarshal(resp, &domains)
	if errJson != nil {
		log.Fatal("解析失败", errJson)
	}
	for _, domain := range  domains.Domains{
		domainList = append(domainList, domain.Name)
	}
	return domainList
}


func GetBandWidth(credential *auth.Credentials, domain string, rangeTime int64, delayTime int64, granularity string) float64 {
	startDate :=time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02 15:04:05")
	endDate := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02 15:04:05")
	reqBody := map[string]interface{}{
		"startDate":   startDate,
		"endDate":     endDate,
		"granularity": granularity,
		"domains":     domain,
		"type":        "bandwidth",
	}

	resData, reqErr := postRequest(credential, "/v2/tune/monitoring/bandwidth", reqBody)
	if reqErr != nil {
		log.Fatal("请求失败", reqErr)
	}
	var (
		response bandWidthResp
		bandWithTotal int64
		count int
		bandWithAverage float64
	)
	errJson := json.Unmarshal(resData, &response)
	if errJson != nil {
		log.Fatal("解析失败", errJson)
	}
	for _, point := range response.Data {
		for _, bandWidth := range point.China{
			bandWithTotal += bandWidth
		}
		count = len(point.China)
	}
	bandWithAverage = float64(bandWithTotal) / float64(count)
	return bandWithAverage
}


func GetHitMiss(credential *auth.Credentials, domain string, rangeTime int64, delayTime int64, granularity string) (hitRateAverage float64, fluxHitRateAverage float64){
	startDate :=time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02")
	endDate := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02")
	reqBody := map[string]interface{}{
		"startDate":   startDate,
		"endDate":     endDate,
		"freq": granularity,
		"domains":     []string{domain},
	}

	resData, reqErr := postRequest(credential, "/v2/tune/loganalyze/hitmiss", reqBody)
	if reqErr != nil {
		log.Fatal("请求失败", reqErr)
	}
	var (
		response hitMissResp
		hitTotal int64
		mistTotal int64
		fluxHitTotal int64
		fluxMissTotal int64
	)
	errJson := json.Unmarshal(resData, &response)
	if errJson != nil {
		log.Fatal("解析失败", errJson)
	}
	for _, hitPoint := range response.Data.Hit {
		hitTotal += hitPoint
	}
	for _, missPoint :=range response.Data.Miss {
		mistTotal += missPoint
	}
	for _, fluxHitPoint := range response.Data.TrafficHit {
		fluxHitTotal += fluxHitPoint
	}
	for _, fluxMissPoint :=range response.Data.TrafficMiss {
		fluxMissTotal += fluxMissPoint
	}
	hitRateAverage = float64(hitTotal) / float64(mistTotal + hitTotal)
	fluxHitRateAverage = float64(fluxHitTotal) / float64(fluxHitTotal + fluxMissTotal)
	return
}


func GetStatusCode(credential *auth.Credentials, domain string, rangeTime int64, delayTime int64, granularity string) map[string]float64 {
	startDate :=time.Now().Add(-time.Second * time.Duration(rangeTime)).Format("2006-01-02")
	endDate := time.Now().Add(-time.Second * time.Duration(delayTime)).Format("2006-01-02")

	reqBody := map[string]interface{}{
		"startDate":   startDate,
		"endDate":     endDate,
		"freq": granularity,
		"domains":     []string{domain},
	}

	resData, reqErr := postRequest(credential, "/v2/tune/loganalyze/statuscode", reqBody)
	if reqErr != nil {
		log.Fatal("请求失败", reqErr)
	}
	var (
		response statusResp
		allStatusTotal int64
	)
	errJson := json.Unmarshal(resData, &response)
	if errJson != nil {
		log.Fatal("解析失败", errJson)
	}
	statusTotal := make(map[string]int64)
	statusProportion := make(map[string]float64)
	for key, point := range response.Data.Codes {
		var total int64
		for _, value := range point {
			total += value
			allStatusTotal += value
		}
		statusTotal[key] = total
	}
	for status, total := range statusTotal {
		if !strings.HasSuffix(status, "x") {
			proportion := float64(total) / float64(allStatusTotal)
			statusProportion[status] = proportion

			if strings.HasPrefix(status, "2") {
				statusProportion["2xx"] += proportion
			} else if strings.HasPrefix(status, "3") {
				statusProportion["3xx"] += proportion
			} else if strings.HasPrefix(status, "4") {
				statusProportion["4xx"] += proportion
			} else if strings.HasPrefix(status, "5") {
				statusProportion["5xx"] += proportion
			}
		}
	}
	return statusProportion
}