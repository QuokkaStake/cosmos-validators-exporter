package state

import (
	"fmt"
	"main/pkg/constants"
	"reflect"
)

type State map[constants.FetcherName]interface{}

func (s State) GetData(fetcherNames []constants.FetcherName) []interface{} {
	data := make([]interface{}, len(fetcherNames))

	for index, fetcherName := range fetcherNames {
		data[index] = s[fetcherName]
	}

	return data
}

func (s State) Set(fetcherName constants.FetcherName, data interface{}) {
	s[fetcherName] = data
}

func StateGet[T any](state State, fetcherName constants.FetcherName) (T, bool) {
	var zero T

	dataRaw, found := state[fetcherName]
	if !found {
		return zero, false
	}

	return Convert[T](dataRaw)
}

func Convert[T any](input interface{}) (T, bool) {
	var zero T

	if input == nil {
		return zero, false
	}

	data, converted := input.(T)
	if !converted {
		panic(fmt.Sprintf(
			"Error converting data: expected %s, got %s",
			reflect.TypeOf(zero).String(),
			reflect.TypeOf(input).String(),
		))
	}

	if reflect.ValueOf(data).Kind() == reflect.Ptr && reflect.ValueOf(data).IsNil() {
		return zero, false
	}

	return data, true
}
