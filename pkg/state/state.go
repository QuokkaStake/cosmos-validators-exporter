package state

import (
	"fmt"
	"main/pkg/constants"
	"reflect"
	"sync"
)

type State struct {
	mutex sync.Mutex
	data  map[constants.FetcherName]interface{}
}

func NewState() *State {
	return &State{
		data: map[constants.FetcherName]interface{}{},
	}
}

func (s *State) GetData(fetcherNames []constants.FetcherName) []interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data := make([]interface{}, len(fetcherNames))

	for index, fetcherName := range fetcherNames {
		data[index] = s.data[fetcherName]
	}

	return data
}

func (s *State) Set(fetcherName constants.FetcherName, data interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[fetcherName] = data
}

func (s *State) Get(fetcherName constants.FetcherName) (interface{}, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, found := s.data[fetcherName]
	return data, found
}

func (s *State) Length() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.data)
}

func StateGet[T any](state *State, fetcherName constants.FetcherName) (T, bool) {
	var zero T

	dataRaw, found := state.Get(fetcherName)
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
