package metrics

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"google.golang.org/grpc"
)

func TestGRPCMetricsInterceptor_Success(t *testing.T) {

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Success",
	}
	counter := GRPCRequestsTotal.WithLabelValues(info.FullMethod, "OK")
	before := testutil.ToFloat64(counter)

	resp, err := GRPCMetricsInterceptor()(
		context.Background(),
		nil,
		info,
		handler,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %q, got %v", "ok", resp)
	}

	after := testutil.ToFloat64(counter)
	if after-before != 1 {
		t.Fatalf("expected counter increment 1, got %v", after-before)
	}
}
