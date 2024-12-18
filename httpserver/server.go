package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	pb "github.com/zalgonoise/tendigitprimes/pb/primes/v1"
)

type Server struct {
	server http.Server
	mux    *runtime.ServeMux
}

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 210 * time.Second // consider wide-range queries
)

func NewServer(addr string) (*Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			md := metadata.MD{}

			otelgrpc.Inject(ctx, &md)

			return md
		}),
	)

	err := mux.HandlePath(http.MethodGet, "/ready", ready)
	if err != nil {
		return nil, err
	}

	tracingMiddleware := otelhttp.NewMiddleware("grpc-gateway",
		otelhttp.WithFilter(func(request *http.Request) bool {
			return request.URL.Path != "/ready" && request.URL.Path != "/metrics"
		}),
	)

	return &Server{
		server: http.Server{
			Handler:      tracingMiddleware(urlAttributesMiddleware(mux)),
			Addr:         addr,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
		},
		mux: mux,
	}, nil
}

func (s *Server) RegisterGRPC(
	ctx context.Context,
	primes pb.PrimesClient,
) error {
	return pb.RegisterPrimesHandlerClient(ctx, s.mux, primes)
}

func (s *Server) RegisterHTTP(method, path string, handler http.Handler) error {
	return s.mux.HandlePath(method, path, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func ready(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
	w.WriteHeader(http.StatusOK)
}

func urlAttributesMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		span := trace.SpanFromContext(request.Context())

		span.SetAttributes(semconv.URLPath(request.URL.Path))

		if request.URL.RawQuery != "" {
			span.SetAttributes(semconv.URLQuery(request.URL.RawQuery))
		}

		h.ServeHTTP(writer, request)
	})
}
