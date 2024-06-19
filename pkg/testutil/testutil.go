package testutil

import (
	"encoding/json"
	"reflect"
)

// JSONEqual compares two json strings. true means they are equal
func JSONEqual(a, b string) (bool, error) {
	var ao interface{}
	var bo interface{}

	var err error
	err = json.Unmarshal([]byte(a), &ao)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(b), &bo)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(ao, bo), nil
}
