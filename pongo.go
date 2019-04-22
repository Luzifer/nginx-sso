package main

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/flosch/pongo2"
)

func init() {
	pongo2.RegisterFilter("to_json", filterToJSON)
}

func filterToJSON(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	var buf = new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(in.Interface())
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "to_json",
			OrigError: err,
		}
	}

	result := strings.TrimSpace(buf.String())
	return pongo2.AsValue(result), nil
}
