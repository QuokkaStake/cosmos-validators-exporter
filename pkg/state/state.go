package state

import (
	"main/pkg/constants"
	"sync"
)

type State struct {
	state map[constants.FetcherName]interface{}
	mutex sync.Mutex
}

func NewState() *State {
	return &State{
		state: map[constants.FetcherName]interface{}{},
	}
}

func (s *State) Set(key constants.FetcherName, value interface{}) {
	s.mutex.Lock()
	s.state[key] = value
	s.mutex.Unlock()
}

func (s *State) Get(key constants.FetcherName) (interface{}, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, found := s.state[key]
	return value, found
}
