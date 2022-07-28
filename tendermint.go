package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type RPC struct {
	URL    string
	Logger zerolog.Logger
}

func NewRPC(url string, logger zerolog.Logger) *RPC {
	return &RPC{
		URL:    url,
		Logger: logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetValidator(address string) (*ValidatorResponse, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s",
		rpc.URL,
		address,
	)

	var response *ValidatorResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) GetDelegationsCount(address string) (*PaginationResponse, QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.count_total=true&pagination.limit=1",
		rpc.URL,
		address,
	)

	var response *PaginationResponse
	info, err := rpc.Get(url, &response)
	if err != nil {
		return nil, info, err
	}

	return response, info, nil
}

func (rpc *RPC) Get(url string, target interface{}) (QueryInfo, error) {
	client := &http.Client{Timeout: 10 * 1000000000}
	start := time.Now()

	info := QueryInfo{
		URL:     url,
		Success: false,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return info, err
	}

	rpc.Logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		info.Duration = time.Since(start)
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return info, err
	}
	defer res.Body.Close()

	info.Duration = time.Since(start)

	rpc.Logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	err = json.NewDecoder(res.Body).Decode(target)
	info.Success = (err == nil)

	return info, err
}
