package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	client "github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/client"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/querydata"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Datasource struct {
	settings        backend.DataSourceInstanceSettings
	httpClient      *http.Client
	resourceHandler backend.CallResourceHandler
}

func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
	d.httpClient.CloseIdleConnections()
}

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	opts, err := settings.HTTPClientOptions()
	if err != nil {
		return nil, fmt.Errorf("http client options: %w", err)
	}
	cl, err := httpclient.New(opts)
	if err != nil {
		return nil, fmt.Errorf("httpclient new: %w", err)
	}
	return &Datasource{
		settings:   settings,
		httpClient: cl,
	}, nil
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	if len(req.Queries) == 0 {
		return &backend.QueryDataResponse{}, fmt.Errorf("query contains no queries")
	}
	instance, err := querydata.New(&http.Client{}, d.settings)
	log.DefaultLogger.Info("Instance: ", instance)
	if err != nil {
		log.DefaultLogger.Error("Create query data instance error, error is: ", err)
		return nil, err
	}
	result, err := instance.Execute(ctx, req)

	return result, err
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult,
	error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status backend.HealthStatus
	var message string

	client2 := client.NewClient(d.httpClient, http.MethodGet, d.settings.URL+"/-/healthy")
	result, err := client2.CheckHealthy(ctx, nil)
	if err != nil {
		status = backend.HealthStatusError
		message = err.Error()
		log.DefaultLogger.Error("CheckHealth error, error is ", err)
	}

	respJson, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.DefaultLogger.Error("Error is: ", err)
	}

	log.DefaultLogger.Info("CheckHealth result is: ", string(respJson))
	if strings.Contains(string(respJson), "Healthy.") {
		status = backend.HealthStatusOk
		message = "Data source is working."
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *Datasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.
	SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	if req.Path == "stream" {
		// Allow subscribing only on expected path.
		status = backend.SubscribeStreamStatusOK
	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest,
	sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField("values", nil, make([]int64, 1)),
	)

	counter := 0

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(time.Second):
			// Send new data periodically.
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, int64(10*(counter%2+1)))

			counter++

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *Datasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.
	PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest,
	sender backend.CallResourceResponseSender) error {
	// 获取后端数据源插件设置详情
	log.DefaultLogger.Info("Request body is", string(req.Body))
	instance, err := querydata.New(&http.Client{}, d.settings)
	log.DefaultLogger.Info("Instance: ", instance)
	if err != nil {
		log.DefaultLogger.Error("Create query data instance error, error is: ", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   []byte(err.Error()),
		})
	}

	// 解析出managerUrl字段
	jsonMap := make(map[string]string)
	if err := json.Unmarshal(instance.JsonData, &jsonMap); err != nil {
		log.DefaultLogger.Error("Query json data to map error, error is: ", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   []byte(err.Error()),
		})
	}
	// 处理json数据
	var bodyMap map[string]interface{}
	if err = json.Unmarshal(req.Body, &bodyMap); err != nil {
		log.DefaultLogger.Error("Body to map error, error is: ", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	if manageUrl, ok := bodyMap["hoursAIUrl"]; ok {
		jsonMap["managerUrl"] = manageUrl.(string)
	}

	log.DefaultLogger.Info("Json map is: ", jsonMap)
	var response []byte
	switch req.Path {
	case util.AlgorithmListType:
		response, err = instance.CallAlgorithmBackend(ctx, req.Body, jsonMap, util.AlgorithmListType)
	case util.GenerateTokenType:
		response, err = instance.CallAlgorithmBackend(ctx, req.Body, jsonMap, util.GenerateTokenType)
	case util.RealtimeInitType:
		response, err = instance.CallAlgorithmBackend(ctx, req.Body, jsonMap, util.RealtimeInitType)

	case util.MetricsType:
		response, err = instance.CallPrometheus(ctx, req.Body, util.MetricsType)
	case util.LabelNamesType:
		response, err = instance.CallPrometheus(ctx, req.Body, util.LabelNamesType)
	case util.SeriesType:
		response, err = instance.CallPrometheus(ctx, req.Body, util.SeriesType)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   []byte(req.Path),
		})
	}
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(err.Error()),
		})
	}
	log.DefaultLogger.Info("Metric response is: ", string(response))
	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   response,
	})
}
