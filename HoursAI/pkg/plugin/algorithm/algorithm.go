package algorithm

import (
	"encoding/json"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/client"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/intervalv2"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/models"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util/converter"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	algorithmListPath        = "/grafana/manager/generics/"
	syncPreviewPath          = "/grafana/manager/once/"
	realtimeResultSinglePath = "/grafana/manager/stream/query/single/"
	realtimeTaskSavePath     = "/grafana/manager/create/"
	realtimeCheckPath        = "/grafana/manager/stream/query/"
	realtimeTaskRemovePath   = "/grafana/manager/delete/"
	createAlertPath          = "/grafana/manager/createAlert/"
	removeAlertPath          = "/grafana/manager/removeAlert/"
	realtimeTaskListPath     = "/grafana/manager/list/"
)

const (
	syncPreviewType          = "syncPreview"
	realtimeCheckType        = "realtimeCheck"
	realtimeResultSingleType = "realtimeResultSingle"
	realtimeTaskSaveType     = "realtimeTaskSave"
	realtimeTaskRemoveType   = "realtimeTaskRemove"
	createAlertType          = "createAlert"
	removeAlertType          = "removeAlert"
	algorithmListType        = "algorithmList"
	realtimeTaskListType     = "realtimeTaskList"
	metricsType              = "metrics"
	labelNamesType           = "labelNames"
	seriesType               = "series"
)

type CallAlgorithmQuery []query

type query struct {
	Labels    string            `json:"labels"`
	Timestamp []int64           `json:"timestamp"`
	Value     []float64         `json:"value"`
	GenericId string            `json:"genericId"`
	Interval  int64             `json:"interval"`
	Tags      map[string]string `json:"tags"`
}

type RealtimeTaskRequest struct {
	GenericId       string `json:"genericId"`
	Query           string `json:"query"`
	Step            int64  `json:"step"`
	AlertEnable     bool   `json:"alertEnable"`
	AlertTemplateId int64  `json:"alertTemplateId"`
	PanelId         int64  `json:"panelId"`
	DashboardUID    string `json:"dashboardUID"`
	PromUri         string `json:"promUri"`
}

type RealtimeResultRequest struct {
	TaskId    string `json:"taskId"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

type RealtimeSingleResultRequest struct {
	TaskId    string `json:"taskId"`
	Series    string `json:"series"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

func newRealtimeTaskRequest(q *models.Query, promClient *client.Client) RealtimeTaskRequest {
	return RealtimeTaskRequest{
		GenericId:       q.GenericId,
		Query:           q.Expr,
		Step:            int64(q.Step / 1000000000),
		AlertEnable:     q.AlertEnable,
		AlertTemplateId: q.AlertTemplateId,
		PanelId:         q.PanelId,
		DashboardUID:    q.DashboardUID,
		PromUri:         promClient.GetClientUrl(),
	}
}

func newRealtimeResultRequest(q *models.Query) RealtimeResultRequest {
	return RealtimeResultRequest{
		TaskId:    q.TaskId,
		StartTime: q.TimeRange().Start.UnixMilli(),
		EndTime:   q.TimeRange().End.UnixMilli(),
	}
}

func newRealtimeSingleResultRequest(q *models.Query) RealtimeSingleResultRequest {
	return RealtimeSingleResultRequest{
		TaskId:    q.TaskId,
		Series:    q.Series,
		StartTime: q.TimeRange().Start.UnixMilli(),
		EndTime:   q.TimeRange().End.UnixMilli(),
	}
}

func newSyncPreview(response *backend.DataResponse, q *models.Query) CallAlgorithmQuery {
	var querys []query
	for _, frame := range response.Frames {
		log.DefaultLogger.Info("Frame: ", frame)

		// 获取timestamp和value
		var (
			timestamps []int64
			values     []float64
		)
		for i := 0; i <= frame.Fields[0].Len()-1; i++ {
			timestamps = append(timestamps, frame.Fields[0].At(i).(time.Time).Unix()*1000)
		}
		for i := 0; i <= frame.Fields[1].Len()-1; i++ {
			values = append(values, frame.Fields[1].At(i).(float64))
		}

		// 获取labels json字符串
		var labelString string
		for _, field := range frame.Fields {
			if field.Labels != nil {
				labels, _ := json.Marshal(field.Labels)
				labelString = string(labels)
			}
		}
		q := query{
			Labels:    labelString,
			Timestamp: timestamps,
			Value:     values,
			GenericId: q.GenericId,
			Interval:  int64(q.Step / 1000000000),
			Tags: map[string]string{
				"promql": q.Expr,
				"legend": q.LegendFormat,
			},
		}
		querys = append(querys, q)
	}
	return querys
}

// CallAlgorithm 调用相关算法接口
func CallAlgorithm(ctx context.Context, r *backend.DataResponse, q *models.Query) (*backend.DataResponse, error) {
	response := &backend.DataResponse{}
	log.DefaultLogger.Info("Datasource json data is: ", q.JsonData)
	// 从JsonData中过滤出需要的字段
	jsonMap := make(map[string]string)
	if err := json.Unmarshal(q.JsonData, &jsonMap); err != nil {
		log.DefaultLogger.Error("Query json data to map error, error is: ", err)
		return response, err
	}
	header := http.Header{
		"accessToken": []string{jsonMap["token"]},
	}

	var body []byte
	var err error
	c := client.NewClient(&http.Client{}, "", "")
	switch q.QueryType {
	case syncPreviewType:
		c.SetUrl(jsonMap["managerUrl"] + syncPreviewPath)
		c.SetMethod(http.MethodPost)
		syncPreviewRequest := newSyncPreview(r, q)
		body, err = json.Marshal(syncPreviewRequest)
	case realtimeCheckType:
		c.SetUrl(jsonMap["managerUrl"] + realtimeCheckPath)
		c.SetMethod(http.MethodPost)
		realTimeRequest := newRealtimeResultRequest(q)
		body, err = json.Marshal(realTimeRequest)
	case realtimeResultSingleType:
		c.SetUrl(jsonMap["managerUrl"] + realtimeResultSinglePath)
		c.SetMethod(http.MethodPost)
		realTimeSingleRequest := newRealtimeSingleResultRequest(q)
		body, err = json.Marshal(realTimeSingleRequest)
	}

	log.DefaultLogger.Info("Http body is: ", string(body))
	if err != nil {
		log.DefaultLogger.Error("Request to json error, error is: ", err)
		return response, err
	}

	resp, err := c.CallAlgorithm(ctx, body, header)
	if err != nil {
		log.DefaultLogger.Error("Http request to call algorithm error, error is: ", err)
		return response, err
	}

	switch q.QueryType {
	case syncPreviewType:
		response, err = converter.ParseAlgorithmResponse(resp, r)
	case realtimeCheckType:
		response, err = converter.ParseRealTimeResultResponse(resp, r)
	case realtimeResultSingleType:
		response, err = converter.ParseRealTimeSingleResultResponse(resp, r)
	}
	if err != nil {
		log.DefaultLogger.Error("Parse algorithm response error, error is: ", err)
		return response, err
	}
	return response, nil
}

type RemoveTaskRequest struct {
	TaskId string `json:"taskId"`
}

type TimeRange struct {
	From  int64  `json:"from"`
	To    int64  `json:"to"`
	Match string `json:"match[]"`
}

// CallCore 调用与查询无关的业务接口
func CallCore(ctx context.Context, body []byte, jsonMap map[string]string, operationType string, timeInterval string,
	intervalCalculator intervalv2.Calculator, jsonData json.RawMessage, promClient *client.Client) ([]byte, error) {
	header := http.Header{
		"accessToken": []string{jsonMap["token"]},
	}
	c := client.NewClient(&http.Client{}, "", "")
	switch operationType {
	case createAlertType:
		c.SetUrl(jsonMap["managerUrl"] + createAlertPath)
		c.SetMethod(http.MethodPost)
	case removeAlertType:
		c.SetUrl(jsonMap["managerUrl"] + removeAlertPath)
		c.SetMethod(http.MethodGet)
	case realtimeTaskSaveType:
		c.SetUrl(jsonMap["managerUrl"] + realtimeTaskSavePath)
		c.SetMethod(http.MethodPost)
		query := backend.DataQuery{JSON: body}
		queryObj, err := models.Parse(query, timeInterval, intervalCalculator, jsonData)
		if err != nil {
			log.DefaultLogger.Error("")
			return []byte(err.Error()), err
		}
		realTimeSaveTask := newRealtimeTaskRequest(queryObj, promClient)
		body, err = json.Marshal(realTimeSaveTask)
		if err != nil {
			return []byte(err.Error()), err
		}
	case realtimeTaskRemoveType:
		request := RemoveTaskRequest{}
		if err := json.Unmarshal(body, &request); err != nil {
			log.DefaultLogger.Error("Realtime task delete json data to map error, error is: ", err)
			return []byte(err.Error()), err
		}
		c.SetUrl(jsonMap["managerUrl"] + realtimeTaskRemovePath + request.TaskId)
		c.SetMethod(http.MethodGet)
	case algorithmListType:
		c.SetUrl(jsonMap["managerUrl"] + algorithmListPath)
		c.SetMethod(http.MethodGet)
	case realtimeTaskListType:
		c.SetUrl(jsonMap["managerUrl"] + realtimeTaskListPath)
		c.SetMethod(http.MethodGet)
	}

	log.DefaultLogger.Info("QueryByte is: ", string(body))
	resp, err := c.CallAlgorithm(ctx, body, header)
	if err != nil {
		log.DefaultLogger.Error("Http request to call algorithm error, error is: ", err)
		return []byte(err.Error()), err
	}

	var result []byte
	switch operationType {
	case realtimeTaskSaveType:
		result, err = converter.ParseRealTimeTaskResponse(resp)
	case algorithmListType:
		result, err = converter.ParseAlgorithmListResponse(resp)
	case realtimeTaskListType:
		result, err = converter.ParseRealTimeTaskListResponse(resp)
	default:
		result, err = converter.ParseNoDataResponse(resp)
	}
	if err != nil {
		log.DefaultLogger.Error("Parse algorithm response error, error is: ", err)
		return []byte(err.Error()), err
	}
	return result, nil
}

//CallPrometheusMetadata 查询prometheus相关metadata
func CallPrometheusMetadata(ctx context.Context, body []byte, operationType string,
	promClient *client.Client) ([]byte, error) {
	var result []byte
	tr := TimeRange{}
	if err := json.Unmarshal(body, &tr); err != nil {
		log.DefaultLogger.Error("Metric query body to struct error, error is: ", err)
		return []byte(err.Error()), err
	}
	var resp *http.Response
	var err error
	switch operationType {
	case metricsType:
		resp, err = promClient.QueryMetrics(ctx, tr.From, tr.To, nil)
	case labelNamesType:
		resp, err = promClient.QueryLabelNames(ctx, tr.From, tr.To, nil)
	case seriesType:
		resp, err = promClient.QuerySeries(ctx, tr.From, tr.To, tr.Match, nil)
	default:
		log.DefaultLogger.Info("")
	}
	if err != nil {
		log.DefaultLogger.Error("Http request to call prometheus metadata error, error is: ", err)
		return []byte(err.Error()), err
	}
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.DefaultLogger.Error("Metrics http response body to []byte error, error is: ", err)
		return []byte(err.Error()), err
	}
	log.DefaultLogger.Info("Call prometheus metadata result is: ", string(result))
	return result, nil
}
