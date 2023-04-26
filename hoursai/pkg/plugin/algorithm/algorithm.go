package algorithm

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/client"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/models"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Point struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type Series []Point

type SyncPreviewQuery struct {
	Series   Series `json:"series"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Params   string `json:"params"`
	Interval int64  `json:"interval"`
	MetaInfo string `json:"metaInfo"`
}

type PrometheusSeriesTimeRange struct {
	From  int64  `json:"from"`
	To    int64  `json:"to"`
	Match string `json:"match[]"`
}

type RealtimeInitRequest struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Params   string `json:"params"`
	MetaInfo string `json:"metaInfo"`
	Interval int64  `json:"interval"`
}

type RealtimeRunRequest struct {
	TaskId   string `json:"taskId"`
	Series   Series `json:"series"`
	MetaInfo string `json:"metaInfo"`
}

type RealtimeResultRequest struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Params    string `json:"params"`
	MetaInfo  string `json:"metaInfo"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	TaskId    string `json:"taskId"`
	Type      string `json:"type"`
	Interval  int64  `json:"interval"`
}

func newRealtimeResultRequest(response *backend.DataResponse, q *models.Query) []RealtimeResultRequest {
	result := make([]RealtimeResultRequest, 0)
	for _, frame := range response.Frames {
		var r RealtimeResultRequest
		_, labelString, algorithm := getSeriesFromResponse(frame, q)
		taskId, metaInfo := getTaskIdFromTaskInfo(q.TaskInfo, labelString, algorithm, q)
		if taskId != "" {
			r = RealtimeResultRequest{
				StartTime: q.TimeRange().Start.Unix(),
				EndTime:   q.TimeRange().End.Unix(),
				Type:      q.Series,
				TaskId:    taskId,
				//Interval:  int64(q.Step / 1000000000),
			}
		} else {
			var (
				metaInfoByte []byte
				err          error
			)
			if metaInfoByte, err = json.Marshal(metaInfo); err != nil {
				log.DefaultLogger.Error("Create sync preview request error,", err)
			}
			r = RealtimeResultRequest{
				StartTime: q.TimeRange().Start.Unix(),
				EndTime:   q.TimeRange().End.Unix(),
				Type:      q.Series,
				Name:      algorithm["name"],
				Version:   algorithm["version"],
				Params:    algorithm["params"],
				MetaInfo:  string(metaInfoByte),
				//Interval:  int64(q.Step / 1000000000),
			}
		}
		result = append(result, r)
	}
	return result
}

func newRealtimeRunRequest(response *backend.DataResponse, q *models.Query) ([]RealtimeRunRequest, []map[string]string) {
	result := make([]RealtimeRunRequest, 0)
	metaInfos := make([]map[string]string, 0)
	for _, frame := range response.Frames {
		s, labelString, algorithm := getSeriesFromResponse(frame, q)
		taskId, metaInfo := getTaskIdFromTaskInfo(q.TaskInfo, labelString, algorithm, q)
		metaInfoByte, err := json.Marshal(metaInfo)
		if err != nil {
			log.DefaultLogger.Error("Create sync preview request error,", err)
		}
		metaInfos = append(metaInfos, metaInfo)
		result = append(result, RealtimeRunRequest{
			TaskId:   taskId,
			Series:   s,
			MetaInfo: string(metaInfoByte),
		})
	}
	return result, metaInfos
}

// getTaskIdFromTaskInfo 从taskInfo中根据序列信息获取taskId
func getTaskIdFromTaskInfo(info []string, labelString string, algorithm map[string]string,
	q *models.Query) (string, map[string]string) {
	var (
		taskMap     map[string]string
		metaInfoMap map[string]string
	)
	for _, task := range info {
		if err := json.Unmarshal([]byte(task), &taskMap); err != nil {
			log.DefaultLogger.Error("Unmarshal task info error,", "err", err)
		}
		if err := json.Unmarshal([]byte(taskMap["metaInfo"]), &metaInfoMap); err != nil {
			log.DefaultLogger.Error("Unmarshal task info error,", "err", err)
		}
		if metaInfoMap["promql"] == q.Expr &&
			metaInfoMap["labels"] == labelString &&
			//metaInfoMap["legend"] == q.LegendFormat &&
			taskMap["name"] == algorithm["name"] &&
			taskMap["params"] == algorithm["params"] &&
			taskMap["version"] == algorithm["version"] {

			return taskMap["taskId"], metaInfoMap
		}
	}
	return "", metaInfoMap
}

// getSeriesFromResponse 从response中获取series
func getSeriesFromResponse(frame *data.Frame, q *models.Query) (Series, string, map[string]string) {
	var (
		s           Series
		labelString string
		algorithm   map[string]string
	)

	// 获取series
	for i := 0; i <= frame.Fields[0].Len()-1; i++ {
		s = append(s, Point{
			Timestamp: frame.Fields[0].At(i).(time.Time).Unix() * 1000,
			Value:     frame.Fields[1].At(i).(float64),
		})
	}

	// 获取labels json字符串
	for _, field := range frame.Fields {
		if field.Labels != nil {
			labels, _ := json.Marshal(field.Labels)
			labelString = string(labels)
		}
	}

	//获取算法信息
	algorithm = map[string]string{
		"name":    q.Name,
		"version": q.Version,
		"params":  q.Params,
	}
	return s, labelString, algorithm
}

func newSyncPreviewRequest(response *backend.DataResponse, q *models.Query) ([]SyncPreviewQuery, []map[string]string) {
	var (
		querys    []SyncPreviewQuery
		metaInfos = make([]map[string]string, 0)
	)
	for _, frame := range response.Frames {
		s, labelString, algorithm := getSeriesFromResponse(frame, q)
		var (
			metaInfoByte []byte
			err          error
			metaInfo     = map[string]string{
				"promql": q.Expr,
				//"legend": q.LegendFormat,
				"labels":   labelString,
				"interval": strconv.FormatInt(int64(q.Step), 10),
			}
		)
		if metaInfoByte, err = json.Marshal(metaInfo); err != nil {
			log.DefaultLogger.Error("Create sync preview request error,", err)
		}

		q := SyncPreviewQuery{
			Name:     algorithm["name"],
			Version:  algorithm["version"],
			Params:   algorithm["params"],
			Series:   s,
			Interval: int64(q.Step / 1000000000),
			MetaInfo: string(metaInfoByte),
		}
		metaInfos = append(metaInfos, metaInfo)
		querys = append(querys, q)
	}
	return querys, metaInfos
}

// CallAlgorithm 调用相关算法接
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
		"Authorization": []string{jsonMap["token"]},
		"sourceType":    []string{util.ProjectType},
	}

	var (
		body      []byte
		err       error
		metaInfos []map[string]string
	)
	c := client.NewClient(&http.Client{Timeout: 60 * time.Second}, "", "")
	switch q.QueryType {
	case util.SyncPreviewType:
		c.SetUrl(jsonMap["managerUrl"] + util.SyncPreviewPath)
		c.SetMethod(http.MethodPost)
		var syncPreviewRequest []SyncPreviewQuery
		syncPreviewRequest, metaInfos = newSyncPreviewRequest(r, q)
		body, err = json.Marshal(syncPreviewRequest)
	case util.RealtimeRunType:
		c.SetUrl(jsonMap["managerUrl"] + util.RealtimeRunPath)
		c.SetMethod(http.MethodPost)
		var realtimeRunRequest []RealtimeRunRequest
		realtimeRunRequest, metaInfos = newRealtimeRunRequest(r, q)
		body, err = json.Marshal(realtimeRunRequest)
	case util.RealtimeResultType:
		c.SetUrl(jsonMap["managerUrl"] + util.RealtimeResultPath)
		c.SetMethod(http.MethodPost)
		realtimeResultRequest := newRealtimeResultRequest(r, q)
		body, err = json.Marshal(realtimeResultRequest)
	}

	if err != nil {
		log.DefaultLogger.Error("Request to json error, error is: ", err)
		return response, err
	}

	resp, err := c.CallAlgorithm(ctx, body, header)
	if err != nil {
		log.DefaultLogger.Error("Http request to call algorithm error, error is: ", err)
		return response, err
	}
	//var result []byte
	//if result, err = io.ReadAll(resp.Body); err != nil {
	//	log.DefaultLogger.Error("Metrics http response body to []byte error, error is: ", err)
	//}
	//log.DefaultLogger.Info("Call prometheus metadata result is: ", string(result))

	response, err = ParseAlgorithmResponse(resp, r, q.QueryType, metaInfos, q.Series)
	if err != nil {
		log.DefaultLogger.Error("Parse algorithm response error, error is: ", err)
		return response, err
	}
	return response, nil
}

// CallCore 调用与查询无关的业务接口
func CallCore(ctx context.Context, body []byte, jsonMap map[string]string, operationType string,
	promClient *client.Client) ([]byte, error) {
	header := http.Header{
		"Authorization": []string{jsonMap["token"]},
		"sourceType":    []string{util.ProjectType},
	}
	c := client.NewClient(&http.Client{}, "", "")
	var err error
	switch operationType {
	case util.AlgorithmListType:
		c.SetUrl(jsonMap["managerUrl"] + util.AlgorithmListPath)
		c.SetMethod(http.MethodGet)
	case util.RealtimeInitType:
		body, err = GenerateRealtimeInitBody(ctx, body, promClient)
		if err != nil {
			log.DefaultLogger.Error("Generate task id error, error is: ", err)
			return []byte(err.Error()), err
		}
		c.SetUrl(jsonMap["managerUrl"] + util.RealtimeInitPath)
		c.SetMethod(http.MethodPost)
	case util.GenerateTokenType:
		c.SetUrl(jsonMap["managerUrl"] + util.GenerateTokenPath)
		c.SetMethod(http.MethodPost)
	}

	resp, err := c.CallAlgorithm(ctx, body, header)
	if err != nil {
		log.DefaultLogger.Error("Http request to call algorithm error, error is: ", err)
		return []byte(err.Error()), err
	}

	result, err := ParseCoreResponse(resp, operationType)
	if err != nil {
		log.DefaultLogger.Error("Parse algorithm response error, error is: ", err)
		return []byte(err.Error()), err
	}
	return result, nil
}

func GenerateRealtimeInitBody(ctx context.Context, body []byte, promClient *client.Client) ([]byte, error) {
	// 将函数体内的变量声明提到最小作用域
	var (
		bodyMap       map[string]interface{}
		timeRange     PrometheusSeriesTimeRange
		timeRangeByte []byte
		pResult       []byte
		promResult    map[string]interface{}
		seriesList    []interface{}
		result        []RealtimeInitRequest
		resultByte    []byte
		err           error
		ok            bool
		metaInfoByte  []byte
	)
	// 处理json数据
	if err = json.Unmarshal(body, &bodyMap); err != nil {
		log.DefaultLogger.Error("Generate task id body to map error, error is: ", err)
		return []byte(err.Error()), err
	}
	// 构建时间范围
	timeRange = PrometheusSeriesTimeRange{
		From:  int64(bodyMap["start"].(float64)),
		To:    int64(bodyMap["end"].(float64)),
		Match: bodyMap["expr"].(string),
	}
	if timeRangeByte, err = json.Marshal(timeRange); err != nil {
		log.DefaultLogger.Error("Generate task id timerange to byte error, error is: ", err)
		return []byte(err.Error()), err
	}
	// 调用prometheus接口获取数据
	if pResult, err = CallPrometheusMetadata(ctx, timeRangeByte, util.SeriesType, promClient, true); err != nil {
		log.DefaultLogger.Error("Call Prometheus Metadata error, error is: ", err)
		return []byte(err.Error()), err
	}
	// 处理prometheus返回数据
	if err = json.Unmarshal(pResult, &promResult); err != nil {
		log.DefaultLogger.Error("Generate task id prometheus data to map error, error is: ", err)
		return []byte(err.Error()), err
	}
	if seriesList, ok = promResult["data"].([]interface{}); !ok || len(seriesList) == 0 {
		log.DefaultLogger.Error("No series list found in prometheus response.")
		return []byte("No series list found in prometheus response."), fmt.Errorf("no series list found in prometheus response")
	}
	// 处理返回结果
	for _, series := range seriesList {
		var seriesByte []byte
		if seriesByte, err = json.Marshal(series); err != nil {
			log.DefaultLogger.Error("Generate task id series byte to string error, error is: ", err)
			return []byte(err.Error()), err
		}
		if metaInfoByte, err = json.Marshal(map[string]string{
			"promql": bodyMap["expr"].(string),
			//"legend": bodyMap["legendFormat"].(string),
			//"interval": strconv.FormatFloat(bodyMap["interval"].(float64), 'f', -1, 64),
			"labels": string(seriesByte),
		}); err != nil {
			log.DefaultLogger.Error("Generate task id meta info to byte error, error is: ", err)
			return []byte(err.Error()), err
		}
		result = append(result, RealtimeInitRequest{
			Name:     bodyMap["name"].(string),
			Params:   bodyMap["params"].(string),
			Version:  bodyMap["version"].(string),
			MetaInfo: string(metaInfoByte),
			Interval: 10000,
		})
	}
	if resultByte, err = json.Marshal(result); err != nil {
		log.DefaultLogger.Error("Generate task id request to json error, error is: ", err)
		return []byte(err.Error()), err
	}
	return resultByte, nil
}

// CallPrometheusMetadata 查询prometheus相关metadata
func CallPrometheusMetadata(ctx context.Context, body []byte, operationType string,
	promClient *client.Client, needTime bool) ([]byte, error) {
	var (
		result []byte
		resp   *http.Response
		err    error
		tr     PrometheusSeriesTimeRange
	)
	if err := json.Unmarshal(body, &tr); err != nil {
		log.DefaultLogger.Error("Metric query body to struct error, error is: ", err)
		return []byte(err.Error()), err
	}

	switch operationType {
	case util.MetricsType:
		resp, err = promClient.QueryMetrics(ctx, tr.From, tr.To, nil)
	case util.LabelNamesType:
		resp, err = promClient.QueryLabelNames(ctx, tr.From, tr.To, nil)
	case util.SeriesType:
		resp, err = promClient.QuerySeries(ctx, tr.From, tr.To, tr.Match, nil, needTime)
	default:
		log.DefaultLogger.Info("")
	}
	if err != nil {
		log.DefaultLogger.Error("Http request to call prometheus metadata error, error is: ", err)
		return []byte(err.Error()), err
	}

	if result, err = io.ReadAll(resp.Body); err != nil {
		log.DefaultLogger.Error("Metrics http response body to []byte error, error is: ", err)
		return []byte(err.Error()), err
	}
	log.DefaultLogger.Info("Call prometheus metadata result is: ", string(result))
	return result, nil
}
