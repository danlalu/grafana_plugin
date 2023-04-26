package converter

import (
	"encoding/json"
	"errors"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	data "github.com/grafana/grafana-plugin-sdk-go/data"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"time"
)

type CoreResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"msg"`
	Code    int         `json:"code"`
}

func ReadAlgorithmStyleResult(iter *jsoniter.Iterator, result *backend.DataResponse, responseType string,
	metaInfos []map[string]string, series string) *backend.DataResponse {
	var (
		rsp       *backend.DataResponse
		code      backend.Status
		status    = "unknown"
		message   = ""
		messageCn = ""
	)
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			switch responseType {
			case util.RealtimeResultType:
				rsp = readRealtimeResultData(iter, result)
			default:
				rsp = readAlgorithmData(iter, result, metaInfos, series)
			}
			log.DefaultLogger.Debug("Case data: ", "key", l1Field, "value", rsp)
		case "message":
			message = iter.ReadString()
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", message)
		case "messageCn":
			messageCn = iter.ReadString()
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", messageCn)
		case "code":
			code = backend.Status(iter.ReadInt())
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			v := iter.Read()
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", v)
		}
	}

	if status == "error" {
		rsp.Error = errors.New(message)
		rsp.Status = code
	}
	return rsp
}

func readAlgorithmData(iter *jsoniter.Iterator, result *backend.DataResponse, metaInfos []map[string]string, series string) *backend.DataResponse {
	labels := data.Labels{}
	for i := 0; iter.ReadArray(); i++ {
		metaInfo := metaInfos[i]
		if err := json.Unmarshal([]byte(metaInfo["labels"]), &labels); err != nil {
			log.DefaultLogger.Error("Label string to map error, ", err)
		}
		interval, err := strconv.ParseFloat(metaInfo["interval"], 64)
		if err != nil {
			log.DefaultLogger.Error("Interval to float error, ", err)
		}
		var (
			code              int
			status            = "unknown"
			message           = ""
			messageCn         = ""
			labels            data.Labels
			timeField         *data.Field
			upperField        *data.Field
			lowerField        *data.Field
			baselineField     *data.Field
			anomalyField      *data.Field
			significanceField *data.Field
		)
		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "status":
				code, status, message, messageCn = readStatus(iter)
				log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
				log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
				log.DefaultLogger.Info("Case message: ", "key", l1Field, "value", message)
				log.DefaultLogger.Info("Case messageCn: ", "key", l1Field, "value", messageCn)
			case "data":
				timeField, _, upperField, lowerField, baselineField, anomalyField,
					significanceField = readData(iter, labels, interval)
			default:
				log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", iter.Read())
			}
		}
		switch series {
		case "anomaly":
			anomalyFrame := data.NewFrame("anomaly", timeField, anomalyField)
			anomalyFrame.Meta = result.Frames[0].Meta
			frames := data.Frames{anomalyFrame}
			result = &backend.DataResponse{
				Frames: frames,
			}
		default:
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

	}
	return result
}

func readData(iter *jsoniter.Iterator, labels data.Labels, interval float64) (*data.Field, *data.Field, *data.Field,
	*data.Field, *data.Field, *data.Field, *data.Field) {
	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = data.TimeSeriesTimeFieldName
	timeField.Config = &data.FieldConfig{Interval: interval * 1000}

	valueField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	valueField.Name = data.TimeSeriesValueFieldName
	valueField.Labels = labels

	upperField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	upperField.Name = data.TimeSeriesValueFieldName
	upperField.Labels = labels

	lowerField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	lowerField.Name = data.TimeSeriesValueFieldName
	lowerField.Labels = labels

	baselineField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	baselineField.Name = data.TimeSeriesValueFieldName
	baselineField.Labels = labels

	anomalyField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	anomalyField.Name = data.TimeSeriesValueFieldName
	anomalyField.Labels = labels

	significanceField := data.NewFieldFromFieldType(data.FieldTypeFloat64, 0)
	significanceField.Name = data.TimeSeriesValueFieldName
	significanceField.Labels = labels
	for iter.ReadArray() {
		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "timestamp":
				timeField.Append(time.UnixMilli(iter.ReadInt64()))
			case "value":
				valueField.Append(iter.ReadFloat64())
			case "upper":
				upperField.Append(iter.ReadFloat64())
			case "lower":
				lowerField.Append(iter.ReadFloat64())
			case "baseline":
				baselineField.Append(iter.ReadFloat64())
			case "anomaly":
				anomalyField.Append(iter.ReadFloat64())
			case "significance":
				significanceField.Append(iter.ReadFloat64())
			default:
				log.DefaultLogger.Info("Algorithm result case default", "key", l1Field, "value", iter.Read())
			}
		}
	}
	return timeField, valueField, upperField, lowerField, baselineField, anomalyField, significanceField
}

func readStatus(iter *jsoniter.Iterator) (int, string, string, string) {
	var (
		code      int
		status    string
		message   string
		messageCn string
	)
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "code":
			code = iter.ReadInt()
			log.DefaultLogger.Info("Algorithm result case code", "key", l1Field, "value", code)
		case "status":
			status = iter.ReadString()
			log.DefaultLogger.Info("Algorithm result case status", "key", l1Field, "value", status)
		case "message":
			message = iter.ReadString()
			log.DefaultLogger.Info("Algorithm result case message", "key", l1Field, "value", message)
		case "messageCn":
			messageCn = iter.ReadString()
			log.DefaultLogger.Info("Algorithm result case messageCn", "key", l1Field, "value", messageCn)
		default:
			log.DefaultLogger.Info("Algorithm result case default", "key", l1Field, "value", iter.ReadString())
		}
	}
	return code, status, message, messageCn
}

func readRealtimeResultData(iter *jsoniter.Iterator, result *backend.DataResponse) *backend.DataResponse {
	for iter.ReadArray() {
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
				log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", iter.Read())
			}
		}
	}
	return result
}

func ReadCoreStyleResult(iter *jsoniter.Iterator, responseType string) CoreResponse {
	result := CoreResponse{}
	status := "unknown"
	message := ""

	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "status":
			status = iter.ReadString()
			result.Status = status
			log.DefaultLogger.Info("Case status: ", "key", l1Field, "value", status)
		case "data":
			var rsp interface{}
			switch responseType {
			case util.AlgorithmListType:
				rsp = readAlgorithmListData(iter)
			case util.RealtimeInitType:
				rsp = readRealtimeInitData(iter)
			case util.GenerateTokenType:
				rsp = readGenerateToken(iter)
			default:
				log.DefaultLogger.Error("Incorrect response type ", l1Field)
			}
			result.Data = rsp
			log.DefaultLogger.Debug("Case data: ", "key", l1Field, "value", rsp)
		case "message":
			message = iter.ReadString()
			result.Message = message
			log.DefaultLogger.Info("Case msg: ", "key", l1Field, "value", message)
		case "code":
			code := iter.ReadInt()
			result.Code = code
			log.DefaultLogger.Info("Case code: ", "key", l1Field, "value", code)
		default:
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", iter.Read())
		}
	}
	return result
}

func readGenerateToken(iter *jsoniter.Iterator) string {
	var token string
	for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
		switch l1Field {
		case "Token":
			token = iter.ReadString()
		default:
			log.DefaultLogger.Info("Case default: ", "key", l1Field, "value", iter.Read())
		}
	}
	return token
}

func readRealtimeInitData(iter *jsoniter.Iterator) []string {
	taskInfos := make([]string, 0)
	for iter.ReadArray() {
		var (
			taskInfoByte []byte
			taskInfo     map[string]string
			err          error
		)
		taskInfo = make(map[string]string)
		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "status":
				readStatus(iter)
			case "data":
				for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
					switch l2Field {
					case "name":
						name := iter.ReadString()
						taskInfo[l2Field] = name
						log.DefaultLogger.Info("Case algorithm name: ", "key", l1Field, "value", name)
					case "version":
						version := iter.ReadString()
						taskInfo[l2Field] = version
						log.DefaultLogger.Info("Case algorithm version : ", "key", l1Field, "value", version)
					case "params":
						params := iter.ReadString()
						taskInfo[l2Field] = params
						log.DefaultLogger.Info("Case algorithm params : ", "key", l1Field, "value", params)
					case "taskId":
						taskId := iter.ReadString()
						taskInfo[l2Field] = taskId
						log.DefaultLogger.Info("Case taskId : ", "key", l1Field, "value", taskId)
					case "metaInfo":
						metaInfo := iter.ReadString()
						taskInfo[l2Field] = metaInfo
						log.DefaultLogger.Info("Case metaInfo : ", "key", l1Field, "value", metaInfo)
					default:
						log.DefaultLogger.Info("Case default: ", "key", l2Field, "value", iter.Read())
					}
				}
			}
		}
		if taskInfoByte, err = json.Marshal(taskInfo); err != nil {
			log.DefaultLogger.Error("Algorithm to json error,", err)
			return taskInfos
		}
		taskInfos = append(taskInfos, string(taskInfoByte))
	}
	return taskInfos
}

func readAlgorithmListData(iter *jsoniter.Iterator) []string {
	algorithmList := make([]string, 0)
	for iter.ReadArray() {
		var (
			algorithmByte []byte
			err           error
			name          string
		)
		algorithm := make(map[string]string)

		for l1Field := iter.ReadObject(); l1Field != ""; l1Field = iter.ReadObject() {
			switch l1Field {
			case "name":
				name = iter.ReadString()
				log.DefaultLogger.Info("Case scene name: ", "key", l1Field, "value", name)
				if name != "timeseries_anomaly_detection" {
					continue
				}
			case "algorithms":
				if name != "timeseries_anomaly_detection" {
					continue
				}
				for iter.ReadArray() {
					for l2Field := iter.ReadObject(); l2Field != ""; l2Field = iter.ReadObject() {
						switch l2Field {
						case "name":
							name := iter.ReadString()
							algorithm[l2Field] = name
							log.DefaultLogger.Info("Case algorithm name: ", "key", l1Field, "value", name)
						case "version":
							version := iter.ReadString()
							algorithm[l2Field] = version
							log.DefaultLogger.Info("Case algorithm version : ", "key", l1Field, "value", version)
						case "params":
							params := iter.ReadString()
							algorithm[l2Field] = params
							log.DefaultLogger.Info("Case algorithm params : ", "key", l1Field, "value", params)
						default:
							log.DefaultLogger.Info("Case default: ", "key", l2Field, "value", iter.Read())
						}
					}
					if algorithmByte, err = json.Marshal(algorithm); err != nil {
						log.DefaultLogger.Error("Algorithm to json error,", err)
						return algorithmList
					}
					algorithmList = append(algorithmList, string(algorithmByte))
				}
			}
		}
	}
	return algorithmList
}
