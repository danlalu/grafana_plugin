package querydata

import (
	"context"
	"encoding/json"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/algorithm"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/client"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/intervalv2"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/models"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"net/http"
	"regexp"
)

const legendFormatAuto = "__auto"

var legendFormatRegexp = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

type QueryData struct {
	intervalCalculator intervalv2.Calculator
	client             *client.Client
	ID                 int64
	URL                string
	TimeInterval       string
	enableWideSeries   bool
	JsonData           json.RawMessage
}

func New(httpClient *http.Client, settings backend.DataSourceInstanceSettings) (*QueryData, error) {
	jsonData, err := util.GetJsonData(settings)
	if err != nil {
		return nil, err
	}
	httpMethod, _ := util.GetStringOptional(jsonData, "httpMethod")

	timeInterval, err := util.GetStringOptional(jsonData, "timeInterval")
	if err != nil {
		return nil, err
	}

	promClient := client.NewClient(httpClient, httpMethod, settings.URL)
	log.DefaultLogger.Info("Query data info is", "url:", settings.URL,
		"TimeInterval:", timeInterval, "ID: ", settings.ID)
	return &QueryData{
		intervalCalculator: intervalv2.NewCalculator(),
		client:             promClient,
		TimeInterval:       timeInterval,
		ID:                 settings.ID,
		URL:                settings.URL,
		enableWideSeries:   false,
		JsonData:           settings.JSONData,
	}, nil
}

func (s *QueryData) Execute(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	result := backend.QueryDataResponse{Responses: backend.Responses{}}
	log.DefaultLogger.Info("The request contains ", "queries length:", len(req.Queries))
	for _, query := range req.Queries {
		log.DefaultLogger.Info("The current query is", query)
		// 把query的json解析成QueryData结构体
		query, err := models.Parse(query, s.TimeInterval, s.intervalCalculator, s.JsonData)
		if err != nil {
			return &result, err
		}
		r := &backend.DataResponse{}
		r, err = s.fetch(ctx, s.client, query, req.Headers)
		if err != nil {
			log.DefaultLogger.Error("Fetch data from prometheus error, error is: ", err)
			return &result, err
		}
		if len(r.Frames) == 0 {
			log.DefaultLogger.Error("Received nil response from runQuery", "query", query.Expr)
			log.DefaultLogger.Debug("Final result is: ", r)
			result.Responses[query.RefId] = *r
			continue
		}
		// 调用算法接口
		r, err = algorithm.CallAlgorithm(ctx, r, query)
		if err != nil {
			r = &backend.DataResponse{
				Status: backend.StatusInternal,
				Error:  err,
			}
			result.Responses[query.RefId] = *r
			log.DefaultLogger.Error("Call algorithm error, err is: ", err)
			return &result, err
		}
		result.Responses[query.RefId] = *r
	}
	log.DefaultLogger.Info("Final result is: ", result)
	return &result, nil
}

func (s *QueryData) fetch(ctx context.Context, client *client.Client, q *models.Query,
	headers map[string]string) (*backend.DataResponse, error) {
	log.DefaultLogger.Info("Sending query",
		"start", q.Start, "end", q.End, "step", q.Step, "query", q.Expr)

	response := &backend.DataResponse{
		Frames: data.Frames{},
		Error:  nil,
	}

	if q.InstantQuery {
		log.DefaultLogger.Info("This query is instant query.")
		res, err := s.instantQuery(ctx, client, q, headers)
		if err != nil {
			return nil, err
		}
		response.Error = res.Error
		response.Frames = res.Frames
	}

	if q.RangeQuery {
		log.DefaultLogger.Info("This query is range query.")
		res, err := s.rangeQuery(ctx, client, q, headers)
		if err != nil {
			return nil, err
		}
		if res.Error != nil {
			response.Error = res.Error
		}
		response.Frames = append(response.Frames, res.Frames...)
	}

	if q.ExemplarQuery {
		log.DefaultLogger.Info("This query is exemplar query.")
		res, err := s.exemplarQuery(ctx, client, q, headers)
		if err != nil {
			log.DefaultLogger.Error("Exemplar query failed", "query", q.Expr, "err", err)
		}
		if res != nil {
			response.Frames = append(response.Frames, res.Frames...)
		}
	}
	return response, nil
}

func (s *QueryData) rangeQuery(ctx context.Context, c *client.Client, q *models.Query,
	headers map[string]string) (*backend.DataResponse, error) {
	res, err := c.QueryRange(ctx, q, util.SdkHeaderToHttpHeader(headers))
	if err != nil {
		return nil, err
	}
	return s.parseResponse(q, res)
}

func (s *QueryData) instantQuery(ctx context.Context, c *client.Client, q *models.Query,
	headers map[string]string) (*backend.DataResponse, error) {
	res, err := c.QueryInstant(ctx, q, util.SdkHeaderToHttpHeader(headers))
	if err != nil {
		return nil, err
	}
	return s.parseResponse(q, res)
}

func (s *QueryData) exemplarQuery(ctx context.Context, c *client.Client, q *models.Query,
	headers map[string]string) (*backend.DataResponse, error) {
	res, err := c.QueryExemplars(ctx, q, util.SdkHeaderToHttpHeader(headers))
	if err != nil {
		return nil, err
	}
	return s.parseResponse(q, res)
}

func (s *QueryData) CallAlgorithmBackend(ctx context.Context, body []byte, jsonMap map[string]string,
	operationType string) ([]byte, error) {
	return algorithm.CallCore(ctx, body, jsonMap, operationType, s.client)
}

func (s *QueryData) CallPrometheus(ctx context.Context, body []byte, operationType string) ([]byte, error) {
	return algorithm.CallPrometheusMetadata(ctx, body, operationType, s.client, false)
}
