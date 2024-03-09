package dex_screener

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"main/pkg/utils"

	"github.com/rs/zerolog"
)

type DexScreener struct {
	Logger zerolog.Logger
}

func NewDexScreener(logger *zerolog.Logger) *DexScreener {
	return &DexScreener{
		Logger: logger.With().Str("component", "dex_screener").Logger(),
	}
}

type DexScreenerPair struct {
	PriceUSD string `json:"priceUsd"`
}

type DexScreenerResponse struct {
	Pairs []DexScreenerPair `json:"pairs"`
}

func (d *DexScreener) GetCurrency(chainID string, pair string) (float64, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/pairs/%s/%s", chainID, pair)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		d.Logger.Error().Err(err).Msg("Error initializing request")
		return 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		d.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		d.Logger.Warn().
			Str("url", url).
			Err(err).
			Int("status", res.StatusCode).
			Msg("Query returned bad HTTP code")
		return 0, fmt.Errorf("bad HTTP code: %d", res.StatusCode)
	}

	var response DexScreenerResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if len(response.Pairs) == 0 {
		d.Logger.Warn().
			Str("url", url).
			Err(err).
			Int("status", res.StatusCode).
			Msg("Got no pairs in response")
		return 0, errors.New("malformed response")
	}

	return utils.StrToFloat64(response.Pairs[0].PriceUSD), err
}
