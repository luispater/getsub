package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//noinspection GoUnusedExportedFunction
func HttpGet(url string) ([]byte, error) {
	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return nil, errReadAll
		}
		return bodyBytes, nil
	}
	return nil, err
}

//noinspection GoUnusedExportedFunction
func HttpPostJson(url string, param interface{}) ([]byte, error) {
	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}
	byteJson, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(byteJson))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return nil, errReadAll
		}
		return bodyBytes, nil
	}
	return nil, err
}

//noinspection GoUnusedExportedFunction
func HttpPost(url string, params map[string]interface{}) ([]byte, error) {
	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	strParams := HttpBuildQuery(params)
	resp, err := client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(strParams))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, errReadAll := ioutil.ReadAll(resp.Body)
		if errReadAll != nil {
			return nil, errReadAll
		}
		return bodyBytes, nil
	}
	return nil, err
}

func HttpBuildQuery(queryData map[string]interface{}) string {
	arrayParams := make([]string, 0)
	for strKey := range queryData {
		arrayParams = append(arrayParams, strKey+"="+ToStr(queryData[strKey]))
	}
	return strings.Join(arrayParams, "&")
}

func ToStr(obj interface{}) string {
	switch obj.(type) {
	case string:
		return fmt.Sprintf("%s", obj)
	case uint:
		return fmt.Sprintf("%d", obj)
	case int:
		return fmt.Sprintf("%d", obj)
	case int64:
		return fmt.Sprintf("%d", obj)
	case float32:
		return fmt.Sprintf("%f", obj)
	case float64:
		return fmt.Sprintf("%f", obj)
	default:
		return fmt.Sprint(obj)
	}
}
