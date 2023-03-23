package converter

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"time"
)

type Options struct {
	MatrixWideSeries bool
	VectorWideSeries bool
}

func ReadPrometheusStyleResult(iter *jsoniter.Iterator, opt Options) *backend.DataResponse {
	var rsp *backend.DataResponse
	status := "unknown"
	errorType := ""
	err := ""
	var warnings []data.Notice

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			rsp = readPrometheusData(iter, opt)
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		case "error":
			err = iter.ReadString()
			log.DefaultLogger.Info("Case error: ", "key", l1Field, "value", err)
		case "errorType":
			errorType = iter.ReadString()
			log.DefaultLogger.Info("Case errorType: ", "key", l1Field, "value", errorType)
		case "warnings":
			warnings = readWarnings(iter)
			log.DefaultLogger.Info("Case warnings: ", "key", l1Field, "value", warnings)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	if status == "error" {
		return &backend.DataResponse{
			Error: fmt.Errorf("%s: %s", errorType, err),
		}
	}

	if len(warnings) > 0 {
		for _, frame := range rsp.Frames {
			if frame.Meta == nil {
				frame.Meta = &data.FrameMeta{}
			}
			frame.Meta.Notices = warnings
		}
	}

	return rsp
}

func readPrometheusData(iter *jsoniter.Iterator, opt Options) *backend.DataResponse {
	t := iter.WhatIsNext()
	if t == jsoniter.ArrayValue {
		return readArrayData(iter)
	}

	if t != jsoniter.ObjectValue {
		return &backend.DataResponse{
			Error: fmt.Errorf("expected object type"),
		}
	}

	resultType := ""
	var rsp *backend.DataResponse
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "resultType":
			resultType = iter.ReadString()
			log.DefaultLogger.Info("Case resultType: ", "key", l1Field, "value", resultType)
		case "result":
			switch resultType {
			case "matrix":
				if opt.MatrixWideSeries {
					rsp = readMatrixOrVectorWide(iter, resultType)
				} else {
					rsp = readMatrixOrVectorMulti(iter, resultType)
				}
				log.DefaultLogger.Info("Case result: ", "key", l1Field, "value", rsp)
			case "vector":
				if opt.VectorWideSeries {
					rsp = readMatrixOrVectorWide(iter, resultType)
				} else {
					rsp = readMatrixOrVectorMulti(iter, resultType)
				}
				log.DefaultLogger.Info("Case vector: ", "key", l1Field, "value", rsp)
			case "streams":
				rsp = readStream(iter)
			case "string":
				rsp = readString(iter)
				log.DefaultLogger.Info("Case string: ", "key", l1Field, "value", rsp)
			case "scalar":
				rsp = readScalar(iter)
				log.DefaultLogger.Info("Case scalar: ", "key", l1Field, "value", rsp)
			default:
				iter.Skip()
				rsp = &backend.DataResponse{
					Error: fmt.Errorf("unknown result type: %s", resultType),
				}
				log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", rsp)
			}
		case "status":
			v := iter.Read()
			if len(rsp.Frames) > 0 {
				meta := rsp.Frames[0].Meta
				if meta == nil {
					meta = &data.FrameMeta{}
					rsp.Frames[0].Meta = meta
				}
				meta.Custom = map[string]interface{}{
					"status": v,
				}
			}
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", v)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	return rsp
}

func readWarnings(iter *jsoniter.Iterator) []data.Notice {
	var warnings []data.Notice
	if iter.WhatIsNext() != jsoniter.ArrayValue {
		return warnings
	}

	for iter.ReadArray() {
		if iter.WhatIsNext() == jsoniter.StringValue {
			notice := data.Notice{
				Severity: data.NoticeSeverityWarning,
				Text:     iter.ReadString(),
			}
			warnings = append(warnings, notice)
		}
	}
	return warnings
}

func readScalar(iter *jsoniter.Iterator) *backend.DataResponse {
	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = data.TimeSeriesTimeFieldName
	valueField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	valueField.Name = data.TimeSeriesValueFieldName
	valueField.Labels = data.Labels{}

	t, v, err := readTimeValuePair(iter)
	if err != nil {
		timeField.Append(t)
		valueField.Append(v)
	}

	frame := data.NewFrame("", timeField, valueField)
	frame.Meta = &data.FrameMeta{
		Type:   data.FrameTypeTimeSeriesMany,
		Custom: resultTypeToCustomMeta("scalar"),
	}

	return &backend.DataResponse{
		Frames: []*data.Frame{frame},
	}
}

func resultTypeToCustomMeta(resultType string) map[string]string {
	return map[string]string{"resultType": resultType}
}

func readTimeValuePair(iter *jsoniter.Iterator) (time.Time, float64, error) {
	iter.ReadArray()
	t := iter.ReadFloat64()
	iter.ReadArray()
	v := iter.ReadString()
	iter.ReadArray()

	tt := timeFromFloat(t)
	fv, err := strconv.ParseFloat(v, 64)
	return tt, fv, err
}

func timeFromFloat(fv float64) time.Time {
	return time.UnixMilli(int64(fv * 1000.0)).UTC()
}

func readString(iter *jsoniter.Iterator) *backend.DataResponse {
	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = data.TimeSeriesTimeFieldName
	valueField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	valueField.Name = data.TimeSeriesValueFieldName
	valueField.Labels = data.Labels{}

	iter.ReadArray()
	t := iter.ReadFloat64()
	iter.ReadArray()
	v := iter.ReadString()
	iter.ReadArray()

	tt := timeFromFloat(t)
	timeField.Append(tt)
	valueField.Append(v)

	frame := data.NewFrame("", timeField, valueField)
	frame.Meta = &data.FrameMeta{
		Type:   data.FrameTypeTimeSeriesMany,
		Custom: resultTypeToCustomMeta("string"),
	}

	return &backend.DataResponse{
		Frames: []*data.Frame{frame},
	}
}

func readStream(iter *jsoniter.Iterator) *backend.DataResponse {
	rsp := &backend.DataResponse{}

	labelsField := data.NewFieldFromFieldType(data.FieldTypeJSON, 0)
	labelsField.Name = "__labels"

	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = "Time"

	lineField := data.NewFieldFromFieldType(data.FieldTypeString, 0)
	lineField.Name = "line"

	tsField := data.NewFieldFromFieldType(data.FieldTypeString, 0)
	tsField.Name = "TS"

	labels := data.Labels{}
	labelJson, err := labelsToRawJson(labels)
	if err != nil {
		return &backend.DataResponse{Error: err}
	}

	for iter.ReadArray() {
		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "stream":
				labels := data.Labels{}
				iter.ReadVal(&labels)
				labelJson, err = labelsToRawJson(labels)
				if err != nil {
					return &backend.DataResponse{Error: err}
				}
			case "values":
				for iter.ReadArray() {
					iter.ReadArray()
					ts := iter.ReadString()
					iter.ReadArray()
					line := iter.ReadString()
					iter.ReadArray()

					t := timeFromLokiString(ts)

					labelsField.Append(labelJson)
					timeField.Append(t)
					lineField.Append(line)
					tsField.Append(ts)
				}
			}
		}
	}

	frame := data.NewFrame("", labelsField, timeField, lineField, tsField)
	frame.Meta = &data.FrameMeta{}
	rsp.Frames = append(rsp.Frames, frame)
	return rsp
}

func timeFromLokiString(str string) time.Time {
	s := len(str)
	if s < 19 || (s == 19 && str[0] == '1') {
		ns, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			return time.Unix(0, ns).UTC()
		}
	}

	ss, _ := strconv.ParseInt(str[0:10], 10, 64)
	ns, _ := strconv.ParseInt(str[10:], 10, 64)
	return time.Unix(ss, ns).UTC()
}

func labelsToRawJson(labels data.Labels) (json.RawMessage, error) {
	bytes, err := jsoniter.Marshal(labels)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

type histogramInfo struct {
	time    *data.Field
	yMin    *data.Field
	yMax    *data.Field
	count   *data.Field
	yLayout *data.Field
}

func newHistogramInfo() *histogramInfo {
	hist := &histogramInfo{
		time:    data.NewFieldFromFieldType(data.FieldTypeTime, 0),
		yMin:    data.NewFieldFromFieldType(data.FieldTypeFloat64, 0),
		yMax:    data.NewFieldFromFieldType(data.FieldTypeFloat64, 0),
		count:   data.NewFieldFromFieldType(data.FieldTypeFloat64, 0),
		yLayout: data.NewFieldFromFieldType(data.FieldTypeInt8, 0),
	}

	hist.time.Name = "xMax"
	hist.yMin.Name = "yMin"
	hist.yMax.Name = "yMax"
	hist.count.Name = "count"
	hist.yLayout.Name = "yLayout"

	return hist
}

func readMatrixOrVectorMulti(iter *jsoniter.Iterator, resultType string) *backend.DataResponse {
	rsp := &backend.DataResponse{}

	for iter.ReadArray() {
		timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
		timeField.Name = data.TimeSeriesTimeFieldName
		valueField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
		valueField.Name = data.TimeSeriesValueFieldName
		valueField.Labels = data.Labels{}

		var histogram *histogramInfo

		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "metric":
				iter.ReadVal(&valueField.Labels)
			case "value":
				t, v, err := readTimeValuePair(iter)
				if err == nil {
					timeField.Append(t)
					valueField.Append(v)
				}
			case "values":
				for iter.ReadArray() {
					t, v, err := readTimeValuePair(iter)
					if err == nil {
						timeField.Append(t)
						valueField.Append(v)
					}
				}
			case "histogram":
				if histogram == nil {
					histogram = newHistogramInfo()
				}
				err := readHistogram(iter, histogram)
				if err != nil {
					rsp.Error = err
				}
			case "histograms":
				if histogram == nil {
					histogram = newHistogramInfo()
				}
				for iter.ReadArray() {
					err := readHistogram(iter, histogram)
					if err != nil {
						rsp.Error = err
					}
				}
			default:
				iter.Skip()
				log.DefaultLogger.Info("readMatrixOrVector: %s", l1Field)
			}
		}

		if histogram != nil {
			histogram.yMin.Labels = valueField.Labels
			frame := data.NewFrame(valueField.Name, histogram.time, histogram.yMin, histogram.yMax, histogram.count,
				histogram.yLayout)
			frame.Meta = &data.FrameMeta{
				Type: "heatmap-cells",
			}
			if frame.Name == data.TimeSeriesValueFieldName {
				frame.Name = ""
			}
			rsp.Frames = append(rsp.Frames, frame)
		} else {
			frame := data.NewFrame("", timeField, valueField)
			frame.Meta = &data.FrameMeta{
				Type:   data.FrameTypeTimeSeriesMany,
				Custom: resultTypeToCustomMeta(resultType),
			}
			rsp.Frames = append(rsp.Frames, frame)
		}
	}

	if len(rsp.Frames) == 0 {
		rsp.Status = backend.StatusOK
		rsp.Error = fmt.Errorf("empty query result")
	}
	return rsp
}

func readHistogram(iter *jsoniter.Iterator, hist *histogramInfo) error {
	iter.ReadArray()
	t := timeFromFloat(iter.ReadFloat64())

	var err error

	iter.ReadArray()
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "count":
			iter.Skip()
		case "sum":
			iter.Skip()
		case "buckets":
			for iter.ReadArray() {
				hist.time.Append(t)

				iter.ReadArray()
				hist.yLayout.Append(iter.ReadInt8())

				iter.ReadArray()
				err = appendValueFromString(iter, hist.yMin)
				if err != nil {
					return err
				}

				iter.ReadArray()
				err = appendValueFromString(iter, hist.yMax)
				if err != nil {
					return nil
				}

				iter.ReadArray()
				err = appendValueFromString(iter, hist.count)
				if err != nil {
					return err
				}

				if iter.ReadArray() {
					return fmt.Errorf("expected close array")
				}
			}
		default:
			iter.Skip()
			log.DefaultLogger.Info("[SKIP] readHistogram: %s", l1Field)
		}
	}
	if iter.ReadArray() {
		return fmt.Errorf("expected to be done")
	}
	return nil
}

func appendValueFromString(iter *jsoniter.Iterator, field *data.Field) error {
	v, err := strconv.ParseFloat(iter.ReadString(), 64)
	if err != nil {
		return err
	}
	field.Append(v)
	return nil
}

func readMatrixOrVectorWide(iter *jsoniter.Iterator, resultType string) *backend.DataResponse {
	rowIdx := 0
	timeMap := map[int64]int{}
	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = data.TimeSeriesTimeFieldName
	frame := data.NewFrame("", timeField)
	frame.Meta = &data.FrameMeta{
		Type:   data.FrameTypeTimeSeriesWide,
		Custom: resultTypeToCustomMeta(resultType),
	}
	rsp := &backend.DataResponse{
		Frames: []*data.Frame{},
	}

	for iter.ReadArray() {
		valueField := data.NewFieldFromFieldType(data.FieldTypeNullableFloat64, frame.Rows())
		valueField.Name = data.TimeSeriesValueFieldName
		valueField.Labels = data.Labels{}
		frame.Fields = append(frame.Fields, valueField)

		var histogram *histogramInfo

		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "metric":
				iter.ReadVal(&valueField.Labels)
			case "value":
				timeMap, rowIdx = addValuePairToFrame(frame, timeMap, rowIdx, iter)
			case "values":
				for iter.ReadArray() {
					timeMap, rowIdx = addValuePairToFrame(frame, timeMap, rowIdx, iter)
				}
			case "histogram":
				if histogram == nil {
					histogram = newHistogramInfo()
				}
				err := readHistogram(iter, histogram)
				if err != nil {
					rsp.Error = err
				}
			case "histograms":
				if histogram == nil {
					histogram = newHistogramInfo()
				}
				for iter.ReadArray() {
					err := readHistogram(iter, histogram)
					if err != nil {
						rsp.Error = err
					}
				}
			default:
				iter.Skip()
				log.DefaultLogger.Info("readMatrixOrVector: %s", l1Field)
			}
		}

		if histogram != nil {
			histogram.yMin.Labels = valueField.Labels
			frame := data.NewFrame(valueField.Name, histogram.time, histogram.yMin, histogram.yMax, histogram.count,
				histogram.yLayout)
			frame.Meta = &data.FrameMeta{
				Type: "heatmap-cells",
			}
			if frame.Name == data.TimeSeriesTimeFieldName {
				frame.Name = ""
			}
			rsp.Frames = append(rsp.Frames, frame)
		}
	}

	if len(rsp.Frames) == 0 {
		rsp.Status = backend.StatusOK
		rsp.Error = fmt.Errorf("empty query result")
	}
	return rsp
}

func addValuePairToFrame(frame *data.Frame, timeMap map[int64]int, rowIdx int,
	iter *jsoniter.Iterator) (map[int64]int, int) {
	timeField := frame.Fields[0]
	valueField := frame.Fields[len(frame.Fields)-1]

	t, v, err := readTimeValuePair(iter)
	if err != nil {
		return timeMap, rowIdx
	}

	ns := t.UnixNano()
	i, ok := timeMap[ns]
	if !ok {
		timeMap[ns] = rowIdx
		i = rowIdx
		expandFrame(frame, i)
		rowIdx++
	}

	timeField.Set(i, t)
	valueField.Set(i, &v)
	return timeMap, rowIdx
}

func expandFrame(frame *data.Frame, idx int) {
	for _, f := range frame.Fields {
		if idx+1 > f.Len() {
			f.Extend(idx + 1 - f.Len())
		}
	}
}

func readArrayData(iter *jsoniter.Iterator) *backend.DataResponse {
	lookup := make(map[string]*data.Field)

	var labelFrame *data.Frame
	rsp := &backend.DataResponse{}
	stringField := data.NewFieldFromFieldType(data.FieldTypeString, 0)
	stringField.Name = "Value"
	for iter.ReadArray() {
		switch iter.WhatIsNext() {
		case jsoniter.StringValue:
			stringField.Append(iter.ReadString())
		case jsoniter.ObjectValue:
			exemplar, labelpairs := readLabelOrExemplars(iter)
			if exemplar != nil {
				rsp.Frames = append(rsp.Frames, exemplar)
			} else if labelpairs != nil {
				max := 0
				for _, pair := range labelpairs {
					k := pair[0]
					v := pair[1]
					f, ok := lookup[k]
					if !ok {
						f = data.NewFieldFromFieldType(data.FieldTypeString, 0)
						f.Name = k
						lookup[k] = f

						if labelFrame == nil {
							labelFrame = data.NewFrame("")
							rsp.Frames = append(rsp.Frames, labelFrame)
						}
						labelFrame.Fields = append(labelFrame.Fields, f)
					}
					f.Append(fmt.Sprintf("%v", v))
					if f.Len() > max {
						max = f.Len()
					}
				}
				for _, f := range lookup {
					diff := max - f.Len()
					if diff > 0 {
						f.Extend(diff)
					}
				}
			}
		default:
			ext := iter.ReadAny()
			v := fmt.Sprintf("%v", ext)
			stringField.Append(v)
		}
	}

	if stringField.Len() > 0 {
		rsp.Frames = append(rsp.Frames, data.NewFrame("", stringField))
	}

	return rsp
}

func readLabelOrExemplars(iter *jsoniter.Iterator) (*data.Frame, [][2]string) {
	pairs := make([][2]string, 0, 10)
	labels := data.Labels{}
	var frame *data.Frame

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "seriesLabels":
			iter.ReadVal(&labels)
		case "exemplars":
			lookup := make(map[string]*data.Field)
			timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
			timeField.Name = data.TimeSeriesTimeFieldName
			valueField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
			valueField.Name = data.TimeSeriesValueFieldName
			valueField.Labels = labels
			frame = data.NewFrame("", timeField, valueField)
			frame.Meta = &data.FrameMeta{
				Custom: resultTypeToCustomMeta("exemplar"),
			}
			for iter.ReadArray() {
				for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
					switch l2Field {
					case "value":
						v, _ := strconv.ParseFloat(iter.ReadString(), 64)
						valueField.Append(v)
					case "timestamp":
						ts := timeFromFloat(iter.ReadFloat64())
						timeField.Append(ts)
					case "labels":
						max := 0
						for _, pair := range readLabelsAsPairs(iter) {
							k := pair[0]
							v := pair[1]
							f, ok := lookup[k]
							if !ok {
								f = data.NewFieldFromFieldType(data.FieldTypeString, 0)
								f.Name = k
								lookup[k] = f
								frame.Fields = append(frame.Fields, f)
							}
							f.Append(v)
							if f.Len() > max {
								max = f.Len()
							}
						}
						for _, f := range lookup {
							diff := max - f.Len()
							if diff > 0 {
								f.Extend(diff)
							}
						}
					default:
						iter.Skip()
						frame.AppendNotices(data.Notice{
							Severity: data.NoticeSeverityError,
							Text:     fmt.Sprintf("unable to parse key: %s in response body", l2Field),
						})
					}
				}
			}
		default:
			v := fmt.Sprintf("%v", iter.Read())
			pairs = append(pairs, [2]string{l1Field, v})
		}
	}
	return frame, pairs
}

func readLabelsAsPairs(iter *jsoniter.Iterator) [][2]string {
	pairs := make([][2]string, 0, 10)
	for k := iter.ReadObject(); k != ""; k = iter.ReadObject() {
		pairs = append(pairs, [2]string{k, iter.ReadString()})
	}
	return pairs
}
