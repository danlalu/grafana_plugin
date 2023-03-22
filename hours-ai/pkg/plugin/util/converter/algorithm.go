package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Algorithm struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	AlgorithmName       string `json:"algorithmName"`
	AlgorithmVersion    string `json:"algorithmVersion"`
	BuiltinDisplayNames string `json:"builtinDisplayNames"`
	BuiltinDescriptions string `json:"builtinDescriptions"`
	Parameters          string `json:"parameters"`
}

type AlgorithmList []Algorithm

func ParseAlgorithmResponse(res *http.Response, result *backend.DataResponse) (*backend.DataResponse, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()
	r := &backend.DataResponse{}
	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r = ReadAlgorithmStyleResult(iter, result)
	if r == nil {
		return r, fmt.Errorf("received empty algorithm result from response")
	}
	return r, nil
}

func ReadAlgorithmStyleResult(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	var rsp *backend.DataResponse
	var code backend.Status
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			rsp = readAlgorithmData(iter, result)
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		case "msg":
			msg = iter.ReadString()
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code = backend.Status(iter.ReadInt())
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}

	if status == "error" {
		rsp.Error = errors.New(msg)
		rsp.Status = code
	}
	return rsp
}

func readAlgorithmData(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	var pairs [][2]string

	for iter.ReadArray() {
		var labels data.Labels
		var interval float64
		var timeField *data.Field
		var valueField *data.Field
		var upperField *data.Field
		var lowerField *data.Field
		var baselineField *data.Field
		var anomalyField *data.Field
		var significanceField *data.Field

		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "labels":
				labelsJson := iter.ReadString()
				if err := json.Unmarshal([]byte(labelsJson), &labels); err != nil {
					log.DefaultLogger.Error("Json to map error, error is: ", err)
					return nil
				}
				log.DefaultLogger.Info("Algorithm result case labels", "key", l1Field, "value", labels)
			case "interval":
				interval = iter.ReadFloat64()
				log.DefaultLogger.Info("Algorithm result case interval", "key", l1Field, "value", interval)
			case "timestamp":
				timeField = data.NewFieldFromFieldType(data.FieldTypeTime, 0)
				timeField.Name = data.TimeSeriesTimeFieldName
				timeField.Config = &data.FieldConfig{Interval: interval * 1000}

				for iter.ReadArray() {
					timeField.Append(time.UnixMilli(iter.ReadInt64()))
				}
				log.DefaultLogger.Info("Algorithm result case timestamp", "key", l1Field, "value", timeField)
			case "value":
				valueField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				valueField.Name = data.TimeSeriesValueFieldName
				valueField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					valueField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case value", "key", l1Field, "value", valueField)
			case "upper":
				upperField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				upperField.Name = data.TimeSeriesValueFieldName
				upperField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					upperField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case upper", "key", l1Field, "value", upperField)
			case "lower":
				lowerField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				lowerField.Name = data.TimeSeriesValueFieldName
				lowerField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					lowerField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case lower", "key", l1Field, "value", lowerField)
			case "baseline":
				baselineField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				baselineField.Name = data.TimeSeriesValueFieldName
				baselineField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					baselineField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case baseline", "key", l1Field, "value", baselineField)
			case "anomaly":
				anomalyField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				anomalyField.Name = data.TimeSeriesValueFieldName
				anomalyField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					anomalyField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case anomaly", "key", l1Field, "value", anomalyField)
			case "significance":
				significanceField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
				significanceField.Name = data.TimeSeriesValueFieldName
				significanceField.Labels = labels

				for v := 0; iter.ReadArray(); v++ {
					significanceField.Append(iter.ReadFloat64())
				}
				log.DefaultLogger.Info("Algorithm result case significance", "key", l1Field, "value", significanceField)

			default:
				v := fmt.Sprintf("%v", iter.Read())
				pairs = append(pairs, [2]string{l1Field, v})
			}
		}
		upperFrame := data.NewFrame("upper", timeField, upperField)
		result.Frames = append(result.Frames, upperFrame)
		lowerFrame := data.NewFrame("lower", timeField, lowerField)
		result.Frames = append(result.Frames, lowerFrame)
		baselineFrame := data.NewFrame("baseline", timeField, baselineField)
		result.Frames = append(result.Frames, baselineFrame)
		anomalyFrame := data.NewFrame("anomaly", timeField, anomalyField)
		result.Frames = append(result.Frames, anomalyFrame)
		significanceFrame := data.NewFrame("significance", timeField, significanceField)
		result.Frames = append(result.Frames, significanceFrame)
	}
	return result
}

type AlgorithmListResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
	Msg    string   `json:"msg"`
	Code   int      `json:"code"`
}

func ParseAlgorithmListResponse(res *http.Response) ([]byte, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r := ReadAlgorithmListStyleResult(iter)
	result, err := json.Marshal(r)
	log.DefaultLogger.Info(string(result))
	if err != nil {
		return []byte(`{"msg": "algorithm list result to json error.", "status": "error"}`), err
	}
	return result, nil
}

func ReadAlgorithmListStyleResult(iter *jsoniter.Iterator) AlgorithmListResponse {
	result := AlgorithmListResponse{}
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			result.Status = status
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			rsp := readAlgorithmListData(iter)
			result.Data = rsp
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		case "msg":
			msg = iter.ReadString()
			result.Msg = msg
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code := iter.ReadInt()
			result.Code = code
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	return result
}

func readAlgorithmListData(iter *jsoniter.Iterator) []string {
	algorithmList := make([]string, 0)
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "content":
			for iter.ReadArray() {
				var algorithm Algorithm
				for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
					switch l2Field {
					case "id":
						algorithm.Id = iter.ReadString()
					case "name":
						algorithm.Name = iter.ReadString()
					case "algorithmName":
						algorithm.AlgorithmName = iter.ReadString()
					case "algorithmVersion":
						algorithm.AlgorithmVersion = iter.ReadString()
					case "builtinDisplayNames":
						algorithm.BuiltinDisplayNames = iter.ReadString()
					case "builtinDescriptions":
						algorithm.BuiltinDescriptions = iter.ReadString()
					case "parameters":
						paramArray := make([]map[string]string, 0)
						for iter.ReadArray() {
							param := make(map[string]string)
							var name string
							var value string
							for l3Field := iter.ReadObject(); l3Field != ""; l3Field = iter.ReadObject() {
								switch l3Field {
								case "name":
									name = iter.ReadString()
								case "value":
									value = iter.ReadString()
								}
							}
							param["name"] = name
							param["value"] = value
							paramArray = append(paramArray, param)
						}
						paramJson, err := json.Marshal(paramArray)
						if err != nil {
							log.DefaultLogger.Error("Transfer param to string error, error is: ", err)
						}
						algorithm.Parameters = string(paramJson)
					default:
						v := iter.Read()
						log.DefaultLogger.Info("Case default: ", "key", l2Field, "value", v)
					}
				}
				algorithmString, _ := json.Marshal(algorithm)
				algorithmList = append(algorithmList, string(algorithmString))
			}
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)

		}
	}
	return algorithmList
}

type CoreResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
	Code   int         `json:"code"`
}

func ParseRealTimeTaskListResponse(res *http.Response) ([]byte, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r := ReadRealtimeTaskListStyleResult(iter)
	result, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"msg": "realtime task list result to json error.", "status": "error"}`), err
	}
	return result, nil
}

func ReadRealtimeTaskListStyleResult(iter *jsoniter.Iterator) CoreResponse {
	result := CoreResponse{}
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status := iter.ReadString()
			result.Status = status
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "msg":
			msg = iter.ReadString()
			result.Msg = msg
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code := iter.ReadInt()
			result.Code = code
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		case "data":
			rsp := readRealtimeTaskListData(iter)
			result.Data = rsp
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	log.DefaultLogger.Info("Realtime task save result is: ", result)
	return result
}

func readRealtimeTaskListData(iter *jsoniter.Iterator) []map[string]string {
	result := make([]map[string]string, 0)
	for iter.ReadArray() {
		currentResult := make(map[string]string)
		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "taskId":
				currentResult["taskId"] = iter.ReadString()
				log.DefaultLogger.Info("case taskId: ", "key", l1Field, "value", currentResult["taskId"])
			case "query":
				currentResult["query"] = iter.ReadString()
				log.DefaultLogger.Info("case query: ", "key", l1Field, "value", currentResult["query"])
			default:
				v := iter.Read()
				log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
			}
		}
		result = append(result, currentResult)
	}
	return result
}

func ParseRealTimeTaskResponse(res *http.Response) ([]byte, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r := ReadRealtimeTaskStyleResult(iter)
	result, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"msg": "realtime save result to json error.", "status": "error"}`), err
	}
	return result, nil
}

func ReadRealtimeTaskStyleResult(iter *jsoniter.Iterator) CoreResponse {
	result := CoreResponse{}
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			result.Status = status
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "msg":
			msg = iter.ReadString()
			result.Msg = msg
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code := iter.ReadInt()
			result.Code = code
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		case "data":
			rsp := readRealtimeTaskData(iter)
			result.Data = rsp
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	log.DefaultLogger.Info("Realtime task save result is: ", result)
	return result
}

func readRealtimeTaskData(iter *jsoniter.Iterator) map[string]string {
	result := make(map[string]string)
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "taskId":
			taskId := iter.ReadString()
			result["taskId"] = taskId
			log.DefaultLogger.Info("case taskId: ", "key", l1Field, "value", taskId)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)

		}
	}
	return result
}

func ParseRealTimeResultResponse(res *http.Response, result *backend.DataResponse) (*backend.DataResponse, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()
	r := &backend.DataResponse{}
	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r = ReadRealtimeResultStyleResult(iter, result)
	if r == nil {
		return r, fmt.Errorf("received empty realtime result from response")
	}
	return r, nil
}

func ReadRealtimeResultStyleResult(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	var rsp *backend.DataResponse
	var code backend.Status
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			rsp = readRealtimeResultData(iter, result)
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		case "msg":
			msg = iter.ReadString()
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code = backend.Status(iter.ReadInt())
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}

	if status == "error" {
		rsp.Error = errors.New(msg)
		rsp.Status = code
	}
	return rsp
}

func readRealtimeResultData(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "data":
			for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
				switch l2Field {
				case "result":
					for iter.ReadArray() {
						labels := make(data.Labels)
						var timeField *data.Field
						var valueField *data.Field
						for l3Field := iter.ReadObject(); l3Field != ""; l3Field = iter.ReadObject() {
							switch l3Field {
							case "metric":
								for l4Field := iter.ReadObject(); l4Field != ""; l4Field = iter.ReadObject() {
									labels[l4Field] = iter.ReadString()
								}
								log.DefaultLogger.Info("Algorithm result case labels", "key", l1Field, "value", labels)
							case "values":
								timeField = data.NewFieldFromFieldType(data.FieldTypeTime, 0)
								timeField.Name = data.TimeSeriesTimeFieldName
								valueField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
								valueField.Name = data.TimeSeriesValueFieldName
								valueField.Labels = labels
								for iter.ReadArray() {
									for iter.ReadArray() {
										timeField.Append(time.UnixMilli(iter.ReadInt64() * 1000))
										iter.ReadArray()
										v, _ := strconv.ParseFloat(iter.ReadString(), 64)
										valueField.Append(v)
									}
								}
							}
						}
						switch {
						case strings.HasPrefix(labels["__name__"], "upper"):
							upperFrame := data.NewFrame(labels["__name__"], timeField, valueField)
							result.Frames = append(result.Frames, upperFrame)
						case strings.HasPrefix(labels["__name__"], "lower"):
							lowerFrame := data.NewFrame(labels["__name__"], timeField, valueField)
							result.Frames = append(result.Frames, lowerFrame)
						case strings.HasPrefix(labels["__name__"], "baseline"):
							baselineFrame := data.NewFrame(labels["__name__"], timeField, valueField)
							result.Frames = append(result.Frames, baselineFrame)
						case strings.HasPrefix(labels["__name__"], "anomaly"):
							anomalyFrame := data.NewFrame(labels["__name__"], timeField, valueField)
							result.Frames = append(result.Frames, anomalyFrame)
						case strings.HasPrefix(labels["__name__"], "significance"):
							significanceFrame := data.NewFrame(labels["__name__"], timeField, valueField)
							result.Frames = append(result.Frames, significanceFrame)
						}

					}
				default:
					v := iter.Read()
					log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
				}
			}
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)

		}
	}
	return result
}

func ParseRealTimeSingleResultResponse(res *http.Response, result *backend.DataResponse) (*backend.DataResponse,
	error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()
	r := &backend.DataResponse{}
	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r = ReadRealtimeSingleResultStyleResult(iter, result)
	if r == nil {
		return r, fmt.Errorf("received empty realtime result from response")
	}
	return r, nil
}

func ReadRealtimeSingleResultStyleResult(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	var rsp *backend.DataResponse
	var code backend.Status
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			rsp = readRealtimeSingleResultData(iter, result)
			log.DefaultLogger.Info("Case data: ", "key", l1Field, "value", rsp)
		case "msg":
			msg = iter.ReadString()
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code = backend.Status(iter.ReadInt())
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}

	if status == "error" {
		rsp.Error = errors.New(msg)
		rsp.Status = code
	}
	return rsp
}

func readRealtimeSingleResultData(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "data":
			for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
				switch l2Field {
				case "result":
					for iter.ReadArray() {
						labels := make(data.Labels)
						var timeField *data.Field
						var valueField *data.Field
						for l3Field := iter.ReadObject(); l3Field != ""; l3Field = iter.ReadObject() {
							switch l3Field {
							case "metric":
								for l4Field := iter.ReadObject(); l4Field != ""; l4Field = iter.ReadObject() {
									labels[l4Field] = iter.ReadString()
								}
								log.DefaultLogger.Info("Algorithm result case labels", "key", l1Field, "value", labels)
							case "values":
								timeField = data.NewFieldFromFieldType(data.FieldTypeTime, 0)
								timeField.Name = data.TimeSeriesTimeFieldName
								valueField = data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
								valueField.Name = data.TimeSeriesValueFieldName
								valueField.Labels = labels
								for iter.ReadArray() {
									for iter.ReadArray() {
										timeField.Append(time.UnixMilli(iter.ReadInt64() * 1000))
										iter.ReadArray()
										v, _ := strconv.ParseFloat(iter.ReadString(), 64)
										valueField.Append(v)
									}
								}
							}
						}
						frame := data.NewFrame(labels["__name__"], timeField, valueField)
						result.Frames = append(result.Frames, frame)
					}
				default:
					v := iter.Read()
					log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
				}
			}
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)

		}
	}
	return result
}

func ParseNoDataResponse(res *http.Response) ([]byte, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r := ReadNoDataStyleResult(iter)
	result, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"msg": "realtime delete result to json error.", "status": "error"}`), err
	}
	return result, nil
}

func ReadNoDataStyleResult(iter *jsoniter.Iterator) map[string]string {
	result := make(map[string]string)
	status := "unknown"
	msg := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			result["status"] = status
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "msg":
			msg = iter.ReadString()
			result["msg"] = msg
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", msg)
		case "code":
			code := iter.ReadInt()
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}
	return result
}
