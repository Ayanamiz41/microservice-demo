// frontend — the HTTP storefront for the Online Boutique replica.
//
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo
// src/frontend: a Go net/http service that renders the shop UI from
// templates/ + static/ and aggregates every backend service over gRPC
// (product catalog, cart, currency, shipping, checkout, recommendations,
// ads, shopping assistant). It is a pure client — it defines no server
// contract of its own.
//
// The HTTP server listens on :8080 and additionally serves the standard
// grpc.health.v1.Health service on the same port (HTTP/2 cleartext via
// h2c), so the running process can be verified with:
//
//	grpcurl -plaintext -d '{"service":"frontend"}' \
//	  localhost:8080 grpc.health.v1.Health/Check
package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	port            = "8080"
	defaultCurrency = "USD"
	cookieMaxAge    = 60 * 60 * 48

	cookiePrefix    = "shop_"
	cookieSessionID = cookiePrefix + "session-id"
	cookieCurrency  = cookiePrefix + "currency"
)

// whitelistedCurrencies is the set of currencies the UI offers (same as
// upstream). The real rates still come from the currencyservice; when that
// service is not reachable the frontend falls back to this whitelist so the
// storefront keeps rendering (see rpc.go).
var whitelistedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"CAD": true,
	"JPY": true,
	"GBP": true,
	"TRY": true,
}

// healthServiceName is the canonical gRPC health service name for this
// process. frontend has no server contract (it is a pure client, so the
// src/README.md matrix marks its health service as "—"); this name is what
// its own Health service is registered under so QA can probe it.
const healthServiceName = "frontend"

// ctxKeySessionID is the context key for the per-browser session id.
type ctxKeySessionID struct{}

// frontendServer holds the lazy gRPC client connections to every backend
// service. Connections are created with grpc.NewClient (which does not dial
// at startup), so the frontend boots even when downstream services are not
// reachable; RPCs against them fail later with Unavailable.
type frontendServer struct {
	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn

	currencySvcAddr string
	currencySvcConn *grpc.ClientConn

	cartSvcAddr string
	cartSvcConn *grpc.ClientConn

	recommendationSvcAddr string
	recommendationSvcConn *grpc.ClientConn

	checkoutSvcAddr string
	checkoutSvcConn *grpc.ClientConn

	shippingSvcAddr string
	shippingSvcConn *grpc.ClientConn

	adSvcAddr string
	adSvcConn *grpc.ClientConn

	shoppingAssistantSvcAddr string
	shoppingAssistantSvcConn *grpc.ClientConn
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	srvPort := envOr("PORT", port)
	listenAddr := envOr("LISTEN_ADDR", "")

	svc := &frontendServer{
		productCatalogSvcAddr:    envOr("PRODUCT_CATALOG_SERVICE_ADDR", "localhost:3550"),
		currencySvcAddr:          envOr("CURRENCY_SERVICE_ADDR", "localhost:7000"),
		cartSvcAddr:              envOr("CART_SERVICE_ADDR", "localhost:7070"),
		recommendationSvcAddr:    envOr("RECOMMENDATION_SERVICE_ADDR", "localhost:8080"),
		checkoutSvcAddr:          envOr("CHECKOUT_SERVICE_ADDR", "localhost:5050"),
		shippingSvcAddr:          envOr("SHIPPING_SERVICE_ADDR", "localhost:50051"),
		adSvcAddr:                envOr("AD_SERVICE_ADDR", "localhost:9555"),
		shoppingAssistantSvcAddr: envOr("SHOPPING_ASSISTANT_SERVICE_ADDR", "localhost:8082"),
	}

	svc.productCatalogSvcConn = mustConnGRPC(svc.productCatalogSvcAddr)
	svc.currencySvcConn = mustConnGRPC(svc.currencySvcAddr)
	svc.cartSvcConn = mustConnGRPC(svc.cartSvcAddr)
	svc.recommendationSvcConn = mustConnGRPC(svc.recommendationSvcAddr)
	svc.checkoutSvcConn = mustConnGRPC(svc.checkoutSvcAddr)
	svc.shippingSvcConn = mustConnGRPC(svc.shippingSvcAddr)
	svc.adSvcConn = mustConnGRPC(svc.adSvcAddr)
	svc.shoppingAssistantSvcConn = mustConnGRPC(svc.shoppingAssistantSvcAddr)
	log.Printf("service config: %+v", svc)

	log.Printf("frontend starting to listen on %s:%s", listenAddr, srvPort)
	if err := http.ListenAndServe(listenAddr+":"+srvPort, h2c.NewHandler(newHandler(svc), &http2.Server{})); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// newHandler builds the full HTTP handler chain: routes, the gRPC health
// service (served on the same port via h2c), request logging and session-id
// cookie management. Shared by main() and the test suite.
func newHandler(fe *frontendServer) http.Handler {
	mux := http.NewServeMux()
	fe.routes(mux)

	// Standard grpc.health.v1.Health served over h2c on the same port, so
	// the process passes a plain grpcurl health check without a second
	// listener. gRPC always uses POST, which also keeps the pattern from
	// conflicting with the "GET /" catch-all route.
	grpcServer := grpc.NewServer()
	healthcheck := health.NewServer()
	healthcheck.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthcheck.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthcheck)
	mux.Handle("POST /grpc.health.v1.Health/", grpcServer)

	handler := http.Handler(mux)
	handler = &logHandler{next: handler}
	handler = ensureSessionID(handler)
	return handler
}

// envOr returns the value of the environment variable key, or def when unset.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// mustConnGRPC returns a lazily-connecting client connection to addr.
// grpc.NewClient does not dial at startup, so the service boots even when
// downstream services are not (yet) reachable; RPCs against them fail later
// with Unavailable.
func mustConnGRPC(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc: failed to create client for %s: %v", addr, err)
	}
	return conn
}
