package http

import (
	"context"
	"encoding/json"
	"main/pkg/types"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/rs/zerolog"
)

type Client struct {
	logger zerolog.Logger
	chain  string
	tracer trace.Tracer
}

func NewClient(logger *zerolog.Logger, chain string, tracer trace.Tracer) *Client {
	return &Client{
		logger: logger.With().
			Str("component", "http").
			Str("chain", chain).
			Logger(),
		chain:  chain,
		tracer: tracer,
	}
}

func (c *Client) Get(
	url string,
	target interface{},
	predicate types.HTTPPredicate,
	ctx context.Context,
) (types.QueryInfo, http.Header, error) {
	childCtx, span := c.tracer.Start(ctx, "HTTP request")
	defer span.End()

	client := &http.Client{
		Timeout:   10 * 1000000000,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	start := time.Now()

	queryInfo := types.QueryInfo{
		Success: false,
		Chain:   c.chain,
		URL:     url,
	}

	req, err := http.NewRequestWithContext(childCtx, http.MethodGet, url, nil)
	if err != nil {
		span.RecordError(err)
		return queryInfo, nil, err
	}

	req.Header.Set("User-Agent", "cosmos-validators-exporter")

	c.logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	queryInfo.Duration = time.Since(start)
	if err != nil {
		c.logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return queryInfo, nil, err
	}
	defer res.Body.Close()

	c.logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	if predicateErr := predicate(res); predicateErr != nil {
		return queryInfo, res.Header, predicateErr
	}

	err = json.NewDecoder(res.Body).Decode(target)
	queryInfo.Success = err == nil

	return queryInfo, res.Header, err
}
