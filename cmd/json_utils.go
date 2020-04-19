package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hokaccha/go-prettyjson"
	"github.com/tidwall/sjson"
)

// PrettyPrintJSONObject prints the given object as pretty printed JSON
func PrettyPrintJSONObject(obj interface{}) error {
	s, err := prettyjson.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(s))
	return err
}

// PrettyPrintJSONBytes prints the given byte array as pretty printed JSON
func PrettyPrintJSONBytes(bytes []byte) error {
	s, err := prettyjson.Format(bytes)
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(s))
	return err
}

// PrettyPrintJSONString prints the given string as pretty printed JSON
func PrettyPrintJSONString(str string) error {
	return PrettyPrintJSONBytes([]byte(str))
}

// SetJSONKeyPathValuesFromStringOnBytes sets a JSON dotted path expression of the form (this.key=value) to a new value in a JSON byte array
func SetJSONKeyPathValuesFromStringOnBytes(jsonPathDescriptor string, bytes []byte) ([]byte, error) {
	components := strings.SplitN(jsonPathDescriptor, "=", 2)
	path, value := components[0], components[1:]
	return sjson.SetBytes(bytes, path, value[0])
}

// SetJSONKeyPathValuesFromStringsOnBytes sets an array of JSON dotted path expressions of the form (this.key=value) to a new value in a JSON byte array
func SetJSONKeyPathValuesFromStringsOnBytes(jsonPathDescriptors []string, bytes []byte) ([]byte, error) {
	var err error // declare err to avoid shadowing effects in the loop
	for _, exp := range jsonPathDescriptors {
		bytes, err = SetJSONKeyPathValuesFromStringOnBytes(exp, bytes)
		PrettyPrintJSONBytes(bytes)
		if err != nil {
			return bytes, err
		}
	}
	return bytes, nil
}

// PrettyPrintJSONResponse prints the given API response as pretty printed JSON
func PrettyPrintJSONResponse(resp *resty.Response) error {
	if resp.IsSuccess() {
		if r := resp.Result(); r != nil {
			return PrettyPrintJSONObject(r)
		}
	} else if resp.IsError() {
		if e := resp.Error(); e != nil {
			return PrettyPrintJSONObject(e)
		}
	}
	var result map[string]interface{}
	err := json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return err
	}
	return PrettyPrintJSONObject(result)
}
