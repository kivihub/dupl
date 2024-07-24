package utils

import (
	"encoding/json"
	"strings"
)

func MarshalPretty(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		panic(err)
	}
	return strings.ReplaceAll(string(bytes), "\\u0026", "&")
}
