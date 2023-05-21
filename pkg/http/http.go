package http

import (
	"encoding/json"
	"main/pkg/types"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Client struct {
	logger zerolog.Logger
	chain  string
}

func NewClient(logger *zerolog.Logger, chain string) *Client {
	return &Client{
		logger: logger.With().
			Str("component", "http").
			Str("chain", chain).
			Logger(),
		chain: chain,
	}
}

func (c *Client) Get(url string, target interface{}) (types.QueryInfo, error) {
	client := &http.Client{Timeout: 10 * 1000000000}
	start := time.Now()

	queryInfo := types.QueryInfo{
		Success: false,
		Chain:   c.chain,
		URL:     url,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return queryInfo, err
	}

	req.Header.Set("User-Agent", "cosmos-validators-exporter")

	c.logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	queryInfo.Duration = time.Since(start)
	if err != nil {
		c.logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return queryInfo, err
	}
	defer res.Body.Close()

	c.logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	err = json.NewDecoder(res.Body).Decode(target)
	queryInfo.Success = err == nil

	return queryInfo, err
}
