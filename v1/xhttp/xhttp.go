package xhttp

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/stevenkitter/api/v1/common"
)

func Post(url string, jsonStr []byte) ([]byte, error) {

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	// req.Header.Set("X-Custom-Header", "myvalue")
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func PostStruct(url string, obj interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(obj) //json

	if err != nil {
		return nil, err
	}

	body, err := Post(url, jsonStr)

	return body, err
}

func PostModel(url string, obj interface{}, response interface{}) error {
	body, err := PostStruct(url, obj)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, response) //json 转 对象
	if err != nil {
		return err
	}
	return nil
}
func PostModelToData(url string, obj interface{}, response interface{}) (map[string]interface{}, error) {
	err := PostModel(url, obj, response)
	if err != nil {
		return nil, err
	}
	data := common.StructToMap(response)
	return data, nil
}
