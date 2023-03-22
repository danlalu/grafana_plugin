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

	//query请求
	//	pluginContext := backend.PluginContext{
	//		OrgID:                      1,
	//		PluginID:                   "cloudwise-backend-plugin-HoursAI",
	//		User:                       nil,
	//		AppInstanceSettings:        nil,
	//		DataSourceInstanceSettings: nil,
	//	}
	//
	//	var queries []backend.DataQuery
	//	queries = append(queries, backend.DataQuery{
	//		RefID:         "A",
	//		QueryType:     "timeSeriesQuery",
	//		MaxDataPoints: 1488,
	//		TimeRange: backend.TimeRange{
	//			To:   time.Now(),
	//			From: time.Now().AddDate(0, 0, -2),
	//		},
	//		JSON: []byte(`{"datasourceId":3,"editorMode":"builder","exemplar":false,
	//				"expr":"up{}","interval":"",
	//			"intervalMs":60000,"legendFormat":"__auto",
	//				"maxDataPoints":1488,"queryType":"realtimeResultSingle","range":true,"refId":"A","requestId":"2A",
	//		"utcOffsetSec":2880,
	//				"genericId":"559645131888657152","algorithmList":false, "A_Realtime_Save": "78", "taskId": "78",
	//"series": "upper"}`),p
	//	})
	//	model := &models.QueryModel{}
	//	err := json.Unmarshal(queries[0].JSON, model)
	//	log.DefaultLogger.Info("model:", model)
	//	if err != nil {
	//		log.DefaultLogger.Error("Error is :", err)
	//		syscall.Exit(1)
	//	}
	//
	//	queryDataRequest := &backend.QueryDataRequest{
	//		PluginContext: pluginContext,
	//		Headers:       nil,
	//		Queries:       queries,
	//	}
	//
	//	dsis := backend.DataSourceInstanceSettings{
	//		ID: 3,
	//		//JSONData: []byte(`{"httpMethod":"POST","backendUrl":"http://10.0.14.88:19888/grafana/manager/once"}`),
	//		//JSONData: []byte(`{"httpMethod":"GET","backendUrl":"http://10.0.14.88:19888/grafana/manager/generics"}`),
	//		JSONData: []byte(`{"httpMethod":"POST","managerUrl":"http://10.0.14.88:19888",
	//	"token":"026042be-07e6-4503-8ff9-0af3f7b8e9b9"}`),
	//		URL: "http://10.1.21.149:9090",
	//	}
	//
	//	queryData, _ := querydata.New(&http.Client{}, dsis)
	//	ctx, cancel := context.WithCancel(context.Background())
	//	result, err := queryData.Execute(ctx, queryDataRequest)
	//	if err != nil {
	//		fmt.Println("err: ", err)
	//	}
	//	log.DefaultLogger.Info("Result is: ", result)
	//	cancel()

	// resource handler请求
	//dsis := backend.DataSourceInstanceSettings{
	//	ID: 3,
	//	//JSONData: []byte(`{"httpMethod":"POST","backendUrl":"http://10.0.14.88:19888/grafana/manager/once"}`),
	//	//JSONData: []byte(`{"httpMethod":"GET","backendUrl":"http://10.0.14.88:19888/grafana/manager/generics"}`),
	//	JSONData: []byte(`{"httpMethod":"POST","managerUrl":"http://10.1.20.226:19888","token":"026042be-07e6-4503-8ff9-0af3f7b8e9b9"}`),
	//	URL:      "http://10.0.14.88:9090",
	//}

	//queryData, _ := querydata.New(&http.Client{}, dsis)
	//ctx, cancel := context.WithCancel(context.Background())
	//jsonMap := map[string]string{"httpMethod": "POST", "managerUrl": "http://10.0.14.88:19888",
	//	"token": "026042be-07e6-4503-8ff9-0af3f7b8e9b9"}
	// realtimeTaskSave
	//body := []byte(`{
	//		"expr": "up{}",
	//		"legendFormat": "",
	//		"interval": "",
	//		"intervalMs": 15000,
	//		"maxDataPoints": 1473,
	//		"genericId":"558912910244446976",
	//		"panelId": 1,
	//		"datasourceId": 1,
	//		"alertEnable": true,
	//		"alertTemplateId": 1
	//		}`)
	// RealtimeTaskRemove
	//body = []byte(`{"taskId": "43"}`)
	//// createAlert
	//body = []byte(`{
	//	"taskId": 43,
	//	"alertTemplateId": 1,
	//	"alertTitle": ""
	//	}`)
	//// removeAlert
	//body = []byte(`{"taskId": 38}`)
	//// createAlert
	//body = []byte(`{
	//		"taskId": 38,
	//		"alertTemplateId": 1,
	//		"alertTitle": "abc"
	//		}`)
	//// metrics
	//body = []byte(`{
	//	"from":1677696180,
	//	"to": 1677717780
	//	}`)
	////series
	//body = []byte(`{
	//	"from":1677726920,
	//	"to": 1677651209,
	//	"match[]": "up"
	//	}`)
	//body := []byte(``)
	//
	//// createAlert, removeAlert, realtimeTaskSave, realtimeTaskRemove, algorithmList, realtimeTaskList, metrics,
	////labelNames, series
	//result, err := queryData.CallAlgorithmBackend(ctx, body, jsonMap, "realtimeTaskList")
	//if err != nil {
	//	fmt.Println("err: ", err)
	//}
	//log.DefaultLogger.Info("Result is: ", string(result))
	//cancel()
}
