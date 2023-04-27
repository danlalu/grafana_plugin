package util

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"net/http"
)

func GetJsonData(settings backend.DataSourceInstanceSettings) (map[string]interface{}, error) {
	var jsonData map[string]interface{}
	err := json.Unmarshal(settings.JSONData, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSONData: %w", err)
	}
	return jsonData, nil
}

func GetStringOptional(obj map[string]interface{}, key string) (string, error) {
	if untypedValue, ok := obj[key]; ok {
		if value, ok := untypedValue.(string); ok {
			return value, nil
		} else {
			err := fmt.Errorf("the field '%s' should be a string", key)
			return "", err
		}
	} else {
		return "", nil
	}
}

func SdkHeaderToHttpHeader(headers map[string]string) http.Header {
	httpHeader := make(http.Header)
	for key, val := range headers {
		httpHeader[key] = []string{val}
	}
	return httpHeader
}
