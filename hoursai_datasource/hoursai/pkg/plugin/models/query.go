package models

import (
	"encoding/json"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/intervalv2"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	varInterval     = "$__interval"
	varIntervalMs   = "$__interval_ms"
	varRange        = "$__range"
	varRangeS       = "$__range_s"
	varRangeMs      = "$__range_ms"
	varRateInterval = "$__rate_interval"
)

const (
	varIntervalAlt     = "${__interval}"
	varIntervalMsAlt   = "${__interval_ms}"
	varRangeAlt        = "${__range}"
	varRangeSAlt       = "${__range_s}"
	varRangeMsAlt      = "${__range_ms}"
	varRateIntervalAlt = "${__rate_interval}"
)

type TimeSeriesQueryType string

var safeResolution = 11000

var defaultTimeInterval = 15 * time.Second

type QueryModel struct {
	Expr            string   `json:"expr"`
	LegendFormat    string   `json:"legendFormat"`
	Interval        string   `json:"interval"`
	IntervalMS      int64    `json:"intervalMS"`
	StepMode        string   `json:"stepMode"`
	RangeQuery      bool     `json:"range"`
	InstantQuery    bool     `json:"instant"`
	ExemplarQuery   bool     `json:"exemplar"`
	IntervalFactor  int64    `json:"intervalFactor"`
	UtcOffsetSec    int64    `json:"utcOffsetSec"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Params          string   `json:"params"`
	QueryType       string   `json:"queryType"`
	AlgorithmList   bool     `json:"algorithmList"`
	TaskInfo        []string `json:"A_Realtime_Save"`
	AlertEnable     bool     `json:"alertEnable"`
	AlertTemplateId int64    `json:"alertTemplateId"`
	PanelId         int64    `json:"panelId"`
	DashboardUID    string   `json:"dashboardUID"`
	Series          string   `json:"series"`
}

type TimeRange struct {
	Start time.Time
	End   time.Time
	Step  time.Duration
}

type Query struct {
	Expr            string
	Step            time.Duration
	LegendFormat    string
	Start           time.Time
	End             time.Time
	RefId           string
	InstantQuery    bool
	RangeQuery      bool
	ExemplarQuery   bool
	UtcOffsetSec    int64
	Name            string
	Version         string
	Params          string
	QueryType       string
	JsonData        json.RawMessage
	AlgorithmList   bool
	TaskInfo        []string
	AlertEnable     bool
	AlertTemplateId int64
	PanelId         int64
	DashboardUID    string
	Series          string
}

func (query *Query) TimeRange() TimeRange {
	return TimeRange{
		Step:  query.Step,
		Start: AlignTimeRange(query.Start, query.Step, query.UtcOffsetSec),
		End:   AlignTimeRange(query.End, query.Step, query.UtcOffsetSec),
	}
}

func AlignTimeRange(t time.Time, step time.Duration, offset int64) time.Time {
	offsetNano := float64(offset * 1e9)
	stepNano := float64(step.Nanoseconds())
	return time.Unix(0, int64(math.Floor((float64(t.UnixNano())+offsetNano)/stepNano)*stepNano-offsetNano)).UTC()
}

// Parse 计算interval、转换promql解析式、转换成QueryData结构体
func Parse(query backend.DataQuery, timeInterval string, intervalCalculator intervalv2.Calculator,
	jsonData json.RawMessage) (*Query, error) {
	model := &QueryModel{}
	if err := json.Unmarshal(query.JSON, model); err != nil {
		log.DefaultLogger.Error("Parse query to model error, error is: ", err)
		return nil, err
	}
	log.DefaultLogger.Info("Model is: ", model)

	// 计算出最后的interval值
	interval, err := calculatePrometheusInterval(model, timeInterval, query, intervalCalculator)
	log.DefaultLogger.Info("Final interval is: ", interval)

	if err != nil {
		return nil, err
	}

	timeRange := query.TimeRange.To.Sub(query.TimeRange.From)
	expr := interpolateVariables(model, interval, timeRange, timeInterval)
	rangeQuery := model.RangeQuery
	if !model.InstantQuery && !model.RangeQuery {
		rangeQuery = true
	}

	return &Query{
		Expr:            expr,
		Step:            interval,
		LegendFormat:    model.LegendFormat,
		Start:           query.TimeRange.From,
		End:             query.TimeRange.To,
		RefId:           query.RefID,
		InstantQuery:    model.InstantQuery,
		RangeQuery:      rangeQuery,
		ExemplarQuery:   model.ExemplarQuery,
		UtcOffsetSec:    model.UtcOffsetSec,
		Name:            model.Name,
		Version:         model.Version,
		Params:          model.Params,
		QueryType:       model.QueryType,
		JsonData:        jsonData,
		AlgorithmList:   model.AlgorithmList,
		TaskInfo:        model.TaskInfo,
		AlertEnable:     model.AlertEnable,
		AlertTemplateId: model.AlertTemplateId,
		PanelId:         model.PanelId,
		DashboardUID:    model.DashboardUID,
		Series:          model.Series,
	}, nil
}

// interpolateVariables 按照单位替换promql中的interval占位符
func interpolateVariables(model *QueryModel, interval time.Duration, timeRange time.Duration,
	timeInterval string) string {
	expr := model.Expr
	log.DefaultLogger.Info("Before expr process ", expr)
	rangeMs := timeRange.Milliseconds()
	rangeSRounded := int64(math.Round(float64(rangeMs) / 1000.0))

	var rateInterval time.Duration
	if model.Interval == varRateInterval || model.Interval == varRateIntervalAlt {
		rateInterval = interval
	} else {
		rateInterval = calculateRateInterval(interval, timeInterval)
	}

	expr = strings.ReplaceAll(expr, varIntervalMs, strconv.FormatInt(int64(interval/time.Millisecond), 10))
	expr = strings.ReplaceAll(expr, varInterval, intervalv2.FormatDuration(interval))
	expr = strings.ReplaceAll(expr, varRangeMs, strconv.FormatInt(rangeMs, 10))
	expr = strings.ReplaceAll(expr, varRangeS, strconv.FormatInt(rangeSRounded, 10))
	expr = strings.ReplaceAll(expr, varRange, strconv.FormatInt(rangeSRounded, 10)+"s")
	expr = strings.ReplaceAll(expr, varRateInterval, rateInterval.String())

	// Repetitive code, we should have functionality to unify these
	expr = strings.ReplaceAll(expr, varIntervalMsAlt, strconv.FormatInt(int64(interval/time.Millisecond), 10))
	expr = strings.ReplaceAll(expr, varIntervalAlt, intervalv2.FormatDuration(interval))
	expr = strings.ReplaceAll(expr, varRangeMsAlt, strconv.FormatInt(rangeMs, 10))
	expr = strings.ReplaceAll(expr, varRangeSAlt, strconv.FormatInt(rangeSRounded, 10))
	expr = strings.ReplaceAll(expr, varRangeAlt, strconv.FormatInt(rangeSRounded, 10)+"s")
	expr = strings.ReplaceAll(expr, varRateIntervalAlt, rateInterval.String())
	log.DefaultLogger.Info("After expr process ", expr)
	return expr
}

func calculatePrometheusInterval(model *QueryModel, timeInterval string, query backend.DataQuery,
	intervalCalculator intervalv2.Calculator) (time.Duration, error) {
	qInterval := model.Interval

	if isVariableInterval(qInterval) {
		qInterval = ""
	}

	// 拿到合理的查询interval
	queryInterval, err := intervalv2.GetIntervalFrom(timeInterval, qInterval, model.IntervalMS, defaultTimeInterval)
	if err != nil {
		return time.Duration(0), err
	}
	log.DefaultLogger.Info("Query interval: ", queryInterval)

	// 根据from、to、maxDataPoints、queryInterval计算合理的interval
	calculatedInterval := intervalCalculator.Calculate(query.TimeRange, queryInterval, query.MaxDataPoints)
	log.DefaultLogger.Info("Calculated interval: ", calculatedInterval)

	// 根据预先设置的最大点数求出最小interval，计算出的interval必须大于它，要不然panel可能会显示不全
	safeInterval := intervalCalculator.CalculateSafeInterval(query.TimeRange, int64(safeResolution))
	log.DefaultLogger.Info("Safe interval: ", safeInterval)
	adjustedInterval := safeInterval.Value
	if calculatedInterval.Value > adjustedInterval {
		adjustedInterval = calculatedInterval.Value
	}

	// 转换interval
	if model.Interval == varRateInterval || model.Interval == varRateIntervalAlt {
		return calculateRateInterval(adjustedInterval, timeInterval), nil
	} else {
		intervalFactor := model.IntervalFactor
		if intervalFactor == 0 {
			intervalFactor = 1
		}
		return time.Duration(int64(adjustedInterval) * intervalFactor), nil
	}
}

func calculateRateInterval(interval time.Duration, scrapeInterval string) time.Duration {
	scrape := scrapeInterval
	if scrape == "" {
		scrape = "15s"
	}

	scrapeIntervalDuration, err := intervalv2.ParseIntervalStringToTimeDuration(scrape)
	if err != nil {
		return time.Duration(0)
	}
	rateInterval := time.Duration(int64(math.Max(float64(interval+scrapeIntervalDuration),
		float64(4)*float64(scrapeIntervalDuration))))
	return rateInterval
}

func isVariableInterval(interval string) bool {
	if interval == varInterval || interval == varIntervalMs || interval == varRateInterval {
		return true
	}

	if interval == varIntervalAlt || interval == varIntervalMsAlt || interval == varRateIntervalAlt {
		return true
	}
	return false
}
