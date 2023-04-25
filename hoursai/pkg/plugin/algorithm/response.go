package algorithm

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-datasource-backend-cloudwise/pkg/plugin/util/converter"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

func ParseAlgorithmResponse(res *http.Response, result *backend.DataResponse, responseType string,
	metaInfos []map[string]string, series string) (*backend.DataResponse, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	var r *backend.DataResponse
	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r = converter.ReadAlgorithmStyleResult(iter, result, responseType, metaInfos, series)

	if r == nil {
		return r, fmt.Errorf("received empty from response")
	}
	return r, nil
}

func ParseCoreResponse(res *http.Response, responseType string) ([]byte, error) {
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.DefaultLogger.Error("Failed to close response body", "err", err)
		}
	}()

	var r interface{}
	iter := jsoniter.Parse(jsoniter.ConfigDefault, res.Body, 1024)
	r = converter.ReadCoreStyleResult(iter, responseType)

	result, err := json.Marshal(r)
	log.DefaultLogger.Info(string(result))
	if err != nil {
		return []byte(`{"msg": "algorithm list result to json error.", "status": "error"}`), err
	}
	return result, nil
}
