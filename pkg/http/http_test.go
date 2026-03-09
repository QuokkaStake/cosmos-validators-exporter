package http

import (
	"context"
	"encoding/json"
	"main/assets"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestHttpClientErrorCreating(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(logger, "chain", tracer)
	queryInfo, _, err := client.Get("://test", nil, types.HTTPPredicateAlwaysPass(), nil)
	require.Error(t, err)
	require.False(t, queryInfo.Success)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientPredicateFail(t *testing.T) {
	httpmock.Activate()

	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")).HeaderAdd(http.Header{
			constants.HeaderBlockHeight: []string{"1"},
		}),
	)

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(logger, "chain", tracer)
	queryInfo, _, err := client.Get("https://example.com", nil, types.HTTPPredicateCheckHeightAfter(100), nil)
	require.Error(t, err)
	require.False(t, queryInfo.Success)
}

// TestTransportReuse verifies that the HTTP client reuses a single transport
// across requests, rather than creating a new one per request. A new transport
// per request leaks goroutines (read/write loops per connection pool) that
// linger for IdleConnTimeout (default 90s).
//
//nolint:paralleltest // goroutine count measurement is incompatible with parallel test execution
func TestTransportReuse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	logger := zerolog.Nop()
	tracer := noop.NewTracerProvider().Tracer("test")
	client := NewClient(&logger, "test-chain", tracer)

	predicate := func(res *http.Response) error { return nil }

	// Force GC and record baseline
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	baselineGoroutines := runtime.NumGoroutine()

	const numRequests = 200
	for i := range numRequests {
		var target map[string]string

		_, _, err := client.Get(server.URL, &target, predicate, context.Background())
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	goroutineGrowth := runtime.NumGoroutine() - baselineGoroutines

	// A reused transport adds very few goroutines regardless of request count.
	// A leaking implementation (new transport per request) would show 3x growth
	// (read loop + write loop + idle manager per connection pool).
	if goroutineGrowth > 20 {
		t.Errorf("possible transport leak: goroutine count grew by %d after %d requests "+
			"(expected <20 with connection reuse)", goroutineGrowth, numRequests)
	}
}

func BenchmarkHTTPClientGet(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	logger := zerolog.Nop()
	tracer := noop.NewTracerProvider().Tracer("test")
	client := NewClient(&logger, "test-chain", tracer)

	predicate := func(res *http.Response) error { return nil }

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		var target map[string]string

		_, _, _ = client.Get(server.URL, &target, predicate, context.Background())
	}
}
