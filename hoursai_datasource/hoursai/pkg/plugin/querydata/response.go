package querydata

import (
	"fmt"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/models"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util/converter"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"sort"
	"strings"
	"time"
)

// parseResponse 解析prometheus返回的response
func (s *QueryData) parseResponse(q *models.Query, res *http.Response) (*backend.DataResponse, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r := converter.ReadPrometheusStyleResult(iter, converter.Options{
		MatrixWideSeries: s.enableWideSeries,
		VectorWideSeries: s.enableWideSeries,
	})
	if r == nil {
		return nil, fmt.Errorf("received empty response from prometheus")
	}

	for _, frame := range r.Frames {
		if s.enableWideSeries {
			addMetadataToWideFrame(q, frame)
		} else {
			addMetadataToMultiFrame(q, frame)
		}
	}

	r = processExemplars(r)
	return r, nil
}

func processExemplars(dr *backend.DataResponse) *backend.DataResponse {
	sampler := newExemplarSampler()

	var frames []*data.Frame
	exemplarFrame := data.NewFrame("exemplar")

	for _, frame := range dr.Frames {
		if !isExemplarFrame(frame) {
			frames = append(frames, frame)
			continue
		}

		exemplarFrame.Meta = frame.Meta
		exemplarFrame.RefID = frame.RefID

		step := time.Duration(frame.Fields[0].Config.Interval) * time.Millisecond
		seriesLabels := getSeriesLabels(frame)
		for rowIdx := 0; rowIdx < frame.Fields[0].Len(); rowIdx++ {
			row := frame.RowCopy(rowIdx)
			ts := row[0].(time.Time)
			val := row[1].(float64)
			labels := getLabels(frame, row)
			sampler.update(step, ts, val, seriesLabels, labels)
		}
	}

	exemplars := sampler.getSampledExemplars()
	if len(exemplars) == 0 {
		return dr
	}

	timeField := data.NewField(data.TimeSeriesTimeFieldName, nil, make([]time.Time, 0, len(exemplars)))
	valueField := data.NewField(data.TimeSeriesValueFieldName, nil, make([]float64, 0, len(exemplars)))
	exemplarFrame.Fields = append(exemplarFrame.Fields, timeField, valueField)
	labelNames := sampler.getLabelNames()
	for _, labelName := range labelNames {
		exemplarFrame.Fields = append(exemplarFrame.Fields, data.NewField(labelName, nil, make([]string, 0,
			len(exemplars))))
	}

	for _, b := range exemplars {
		timeField.Append(b.ts)
		valueField.Append(b.val)
		for i, labelName := range labelNames {
			labelValue, ok := b.labels[labelName]
			if !ok {
				labelValue = b.seriesLabels[labelName]
			}
			colIdx := i + 2
			exemplarFrame.Fields[colIdx].Append(labelValue)
		}
	}
	frames = append(frames, exemplarFrame)

	return &backend.DataResponse{
		Frames: frames,
		Error:  dr.Error,
	}
}

func getLabels(frame *data.Frame, row []interface{}) map[string]string {
	labels := make(map[string]string)
	for i := 2; i < len(row); i++ {
		labels[frame.Fields[i].Name] = row[i].(string)
	}
	return labels
}

func getSeriesLabels(frame *data.Frame) data.Labels {
	return frame.Fields[1].Labels.Copy()
}

func isExemplarFrame(frame *data.Frame) bool {
	rt := models.ResultTypeFromFrame(frame)
	return rt == models.ResultTypeExemplar
}

func addMetadataToMultiFrame(q *models.Query, frame *data.Frame) {
	if frame.Meta == nil {
		frame.Meta = &data.FrameMeta{}
	}
	frame.Meta.ExecutedQueryString = executeQueryString(q)
	if len(frame.Fields) < 2 {
		return
	}
	frame.Name = getName(q, frame.Fields[1])
	frame.Fields[0].Config = &data.FieldConfig{Interval: float64(q.Step.Milliseconds())}
	if frame.Name != "" {
		frame.Fields[1].Config = &data.FieldConfig{DisplayNameFromDS: frame.Name}
	}
}

func getName(q *models.Query, field *data.Field) string {
	labels := field.Labels
	legend := metricNameFromLabels(field)

	if q.LegendFormat == legendFormatAuto {
		if len(labels) > 0 {
			legend = ""
		}
	} else if q.LegendFormat != "" {
		result := legendFormatRegexp.ReplaceAllFunc([]byte(q.LegendFormat), func(in []byte) []byte {
			labelName := strings.Replace(string(in), "{{", "", 1)
			labelName = strings.Replace(labelName, "}}", "", 1)
			labelName = strings.TrimSpace(labelName)
			if val, exists := labels[labelName]; exists {
				return []byte(val)
			}
			return []byte{}
		})
		legend = string(result)
	}

	if legend == "{}" {
		return q.Expr
	}

	return legend
}

func metricNameFromLabels(f *data.Field) string {
	labels := f.Labels
	metricName, hasName := labels["__name__"]
	numLabels := len(labels) - 1
	if !hasName {
		numLabels = len(labels)
	}
	labelStrings := make([]string, 0, numLabels)
	for label, value := range labels {
		if label != "__name__" {
			labelStrings = append(labelStrings, fmt.Sprintf("%s=%q", label, value))
		}
	}
	switch numLabels {
	case 0:
		if hasName {
			return metricName
		}
		return "{}"
	default:
		sort.Strings(labelStrings)
		return fmt.Sprintf("%s{%s}", metricName, strings.Join(labelStrings, ", "))

	}
}

func executeQueryString(q *models.Query) string {
	return "Expr: " + q.Expr + "\n" + "Step: " + q.Step.String()
}

func addMetadataToWideFrame(q *models.Query, frame *data.Frame) {
	if frame.Meta == nil {
		frame.Meta = &data.FrameMeta{}
	}
	frame.Meta.ExecutedQueryString = executeQueryString(q)
	if len(frame.Fields) < 2 {
		return
	}
	frame.Name = getName(q, frame.Fields[1])
	frame.Fields[0].Config = &data.FieldConfig{Interval: float64(q.Step.Milliseconds())}
	for _, f := range frame.Fields {
		if f.Type() == data.FieldTypeFloat64 || f.Type() == data.FieldTypeNullableFloat64 {
			f.Name = getName(q, f)
		}
	}
}
