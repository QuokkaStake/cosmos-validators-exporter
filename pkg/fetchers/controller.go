package fetchers

import (
	"context"
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
	"reflect"
	"sync"

	"github.com/rs/zerolog"
)

type FetchersStatuses map[constants.FetcherName]bool

func (s FetchersStatuses) IsAllDone(fetcherNames []constants.FetcherName) bool {
	for _, fetcherName := range fetcherNames {
		if _, ok := s[fetcherName]; !ok {
			return false
		}
	}

	return true
}

type State map[constants.FetcherName]interface{}

func (s State) GetData(fetcherNames []constants.FetcherName) []interface{} {
	data := make([]interface{}, len(fetcherNames))

	for index, fetcherName := range fetcherNames {
		data[index] = s[fetcherName]
	}

	return data
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

type Controller struct {
	Fetchers Fetchers
	Logger   zerolog.Logger
}

func NewController(
	fetchers Fetchers,
	logger *zerolog.Logger,
) *Controller {
	return &Controller{
		Logger: logger.With().
			Str("component", "controller").
			Logger(),
		Fetchers: fetchers,
	}
}

func (c *Controller) Fetch(ctx context.Context) (
	State,
	[]*types.QueryInfo,
) {
	data := State{}
	queries := []*types.QueryInfo{}
	fetchersStatus := FetchersStatuses{}

	var mutex sync.Mutex
	var wg sync.WaitGroup

	processFetcher := func(fetcher Fetcher) {
		defer wg.Done()

		c.Logger.Trace().Str("name", string(fetcher.Name())).Msg("Processing fetcher...")

		mutex.Lock()
		fetcherDependenciesData := data.GetData(fetcher.Dependencies())
		mutex.Unlock()

		fetcherData, fetcherQueries := fetcher.Fetch(ctx, fetcherDependenciesData...)

		mutex.Lock()
		data[fetcher.Name()] = fetcherData
		queries = append(queries, fetcherQueries...)
		fetchersStatus[fetcher.Name()] = true
		mutex.Unlock()

		c.Logger.Trace().
			Str("name", string(fetcher.Name())).
			Msg("Processed fetcher")
	}

	for {
		c.Logger.Trace().Msg("Processing all pending fetchers...")

		if fetchersStatus.IsAllDone(c.Fetchers.GetNames()) {
			c.Logger.Trace().Msg("All fetchers are fetched.")
			break
		}

		fetchersToStart := Fetchers{}

		for _, fetcher := range c.Fetchers {
			if _, ok := fetchersStatus[fetcher.Name()]; ok {
				c.Logger.Trace().
					Str("name", string(fetcher.Name())).
					Msg("Fetcher is already being processed or is processed, skipping.")
				continue
			}

			if !fetchersStatus.IsAllDone(fetcher.Dependencies()) {
				c.Logger.Trace().
					Str("name", string(fetcher.Name())).
					Msg("Fetcher's dependencies are not yet processed, skipping for now.")
				continue
			}

			fetchersToStart = append(fetchersToStart, fetcher)
		}

		c.Logger.Trace().
			Strs("names", fetchersToStart.GetNamesAsString()).
			Msg("Starting the following fetchers")

		wg.Add(len(fetchersToStart))

		for _, fetcher := range fetchersToStart {
			go processFetcher(fetcher)
		}

		wg.Wait()
	}

	return data, queries
}
