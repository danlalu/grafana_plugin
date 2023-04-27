package main

import (
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"os"
)

func main() {
	// Start listening to requests sent from Grafana. This call is blocking so
	// it won't finish until Grafana shuts down the process or the plugin choose
	// to exit by itself using os.Exit. Manage automatically manages life cycle
	// of datasource instances. It accepts datasource instance factory as first
	// argument. This factory will be automatically called on incoming request
	// from Grafana to create different instances of SampleDatasource (per datasource
	// ID). When datasource configuration changed Dispose method will be called and
	// new datasource instance created using NewSampleDatasource factory.
	if err := datasource.Manage("myorgid-simple-backend-datasource", plugin.NewSampleDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}

	////query请求
	//pluginContext := backend.PluginContext{
	//	OrgID:                      1,
	//	PluginID:                   "cloudwise-backend-plugin-HoursAI",
	//	User:                       nil,
	//	AppInstanceSettings:        nil,
	//	DataSourceInstanceSettings: nil,
	//}
	//
	//var queries []backend.DataQuery
	//metaInfo := map[string]interface{}{
	//	"interval": "60000", "labels": "{\"__name__\":\"up\",\"instance\":\"10.1.23.49:9090\",\"job\":\"prometheus\"}",
	//	"legend": "__auto", "promql": "up{}",
	//}
	//metaInfoByte, err := json.Marshal(metaInfo)
	//if err != nil {
	//	panic(err)
	//}
	//taskInfo := map[string]interface{}{
	//	"metaInfo": string(metaInfoByte),
	//	"name":     "Auto Value Detection",
	//	"params":   "[]",
	//	"taskId":   "a76cf0b7010d1b1fbb91981196491036",
	//	"version":  "2.0",
	//}
	//taskInfoByte, err := json.Marshal(taskInfo)
	//if err != nil {
	//	panic(err)
	//}
	//jsonMap := map[string]interface{}{"datasourceId": 3, "editorMode": "builder", "exemplar": false,
	//	"expr": "up{}", "interval": "",
	//	"intervalMs": 60000, "legendFormat": "__auto",
	//	"maxDataPoints": 1488, "queryType": "realtimeResult", "range": true, "refId": "A", "requestId": "2A",
	//	"utcOffsetSec": 2880, "algorithmList": false,
	//	"name":     "Auto Value Detection",
	//	"version":  "2.0",
	//	"params":   "[]",
	//	"series":   "anomaly",
	//	"taskInfo": []string{string(taskInfoByte)},
	//}
	//b, err := json.Marshal(jsonMap)
	//if err != nil {
	//	panic(err)
	//}
	//queries = append(queries, backend.DataQuery{
	//	RefID:         "A",
	//	QueryType:     "timeSeriesQuery",
	//	MaxDataPoints: 1488,
	//	TimeRange: backend.TimeRange{
	//		To:   time.Now(),
	//		From: time.Now().AddDate(0, 0, -2),
	//	},
	//	JSON: b,
	//})
	//model := &models.QueryModel{}
	//err = json.Unmarshal(queries[0].JSON, model)
	//if err != nil {
	//	log.DefaultLogger.Error("Error is :", err)
	//	syscall.Exit(1)
	//}
	//
	//queryDataRequest := &backend.QueryDataRequest{
	//	PluginContext: pluginContext,
	//	Headers:       nil,
	//	Queries:       queries,
	//}
	//
	//dsis := backend.DataSourceInstanceSettings{
	//	ID: 3,
	//	JSONData: []byte(`{"httpMethod":"POST","managerUrl":"http://10.1.20.103:8081",
	//"token":"35eefe37-e31c-4b67-b627-238ffb61c4d2"}`),
	//	URL: "http://10.1.23.49:9090",
	//}
	//
	//queryData, _ := querydata.New(&http.Client{}, dsis)
	//ctx, cancel := context.WithCancel(context.Background())
	//_, err = queryData.Execute(ctx, queryDataRequest)
	//if err != nil {
	//	fmt.Println("err: ", err)
	//}
	//cancel()

	//// resource handler请求
	//dsis := backend.DataSourceInstanceSettings{
	//	ID: 3,
	//	//JSONData: []byte(`{"httpMethod":"POST","backendUrl":"http://10.0.14.88:19888/grafana/manager/once"}`),
	//	//JSONData: []byte(`{"httpMethod":"GET","backendUrl":"http://10.0.14.88:19888/grafana/manager/generics"}`),
	//	JSONData: []byte(`{"httpMethod":"POST","managerUrl":"http://10.1.20.103:8082",
	//"token":"35eefe37-e31c-4b67-b627-238ffb61c4d2"}`),
	//	URL: "http://10.1.23.49:9090",
	//}
	//
	//queryData, _ := querydata.New(&http.Client{}, dsis)
	//ctx, cancel := context.WithCancel(context.Background())
	//jsonMap := map[string]string{"httpMethod": "POST", "managerUrl": "http://10.1.20.103:8081",
	//	"token": "35eefe37-e31c-4b67-b627-238ffb61c4d2"}
	//var body []byte
	////// metrics
	////body = []byte(`{
	////	"from":1677696180,
	////	"to": 1677717780
	////	}`)
	//////series
	////body = []byte(`{
	////	"from":1677726920,
	////	"to": 1677651209,
	////	"match[]": "up"
	////	}`)
	////body = []byte(``)
	//// generateToken
	//body = []byte(`{
	//		"url": "http://10.1.23.49:9090"
	//		}`)
	//// generateTaskId
	//jsonMapp := map[string]interface{}{
	//	"expr":         "up{}",
	//	"name":         "Auto Value Detection",
	//	"version":      "2.0",
	//	"params":       "[]",
	//	"start":        1681084520,
	//	"end":          1681091720,
	//	"interval":     60000,
	//	"legendFormat": "__auto",
	//}
	//body, err := json.Marshal(jsonMapp)
	//if err != nil {
	//	panic(err)
	//}
	//// algorithmList, metrics, generateToken, generateTaskId
	////labelNames, series
	//result, err := queryData.CallAlgorithmBackend(ctx, body, jsonMap, "generateTaskId")
	//if err != nil {
	//	fmt.Println("err: ", err)
	//}
	//log.DefaultLogger.Info("Result is: ", string(result))
	//cancel()
}
