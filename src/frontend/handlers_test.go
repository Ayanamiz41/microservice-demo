package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// --- in-memory fakes for every downstream service ---------------------------

type fakeProductCatalogService struct {
	pb.UnimplementedProductCatalogServiceServer
	products []*pb.Product
}

func (f *fakeProductCatalogService) ListProducts(_ context.Context, _ *pb.Empty) (*pb.ListProductsResponse, error) {
	return &pb.ListProductsResponse{Products: f.products}, nil
}

func (f *fakeProductCatalogService) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	for _, p := range f.products {
		if p.GetId() == req.GetId() {
			return p, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "product %q not found", req.GetId())
}

type fakeCartService struct {
	pb.UnimplementedCartServiceServer
	mu      sync.Mutex
	userID  string
	items   []*pb.CartItem
	emptied bool
}

func (f *fakeCartService) AddItem(_ context.Context, req *pb.AddItemRequest) (*pb.Empty, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.userID = req.GetUserId()
	f.items = append(f.items, req.GetItem())
	return &pb.Empty{}, nil
}

func (f *fakeCartService) GetCart(_ context.Context, req *pb.GetCartRequest) (*pb.Cart, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return &pb.Cart{UserId: req.GetUserId(), Items: f.items}, nil
}

func (f *fakeCartService) EmptyCart(_ context.Context, _ *pb.EmptyCartRequest) (*pb.Empty, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.emptied = true
	f.items = nil
	return &pb.Empty{}, nil
}

// fakeCurrencyService returns a fixed currency list and identity conversion.
type fakeCurrencyService struct {
	pb.UnimplementedCurrencyServiceServer
}

func (f *fakeCurrencyService) GetSupportedCurrencies(_ context.Context, _ *pb.Empty) (*pb.GetSupportedCurrenciesResponse, error) {
	return &pb.GetSupportedCurrenciesResponse{CurrencyCodes: []string{"USD", "EUR", "CAD", "JPY", "GBP", "TRY", "XYZ"}}, nil
}

func (f *fakeCurrencyService) Convert(_ context.Context, req *pb.CurrencyConversionRequest) (*pb.Money, error) {
	return &pb.Money{
		CurrencyCode: req.GetToCode(),
		Units:        req.GetFrom().GetUnits(),
		Nanos:        req.GetFrom().GetNanos(),
	}, nil
}

type fakeRecommendationService struct {
	pb.UnimplementedRecommendationServiceServer
	ids []string
}

func (f *fakeRecommendationService) ListRecommendations(_ context.Context, _ *pb.ListRecommendationsRequest) (*pb.ListRecommendationsResponse, error) {
	return &pb.ListRecommendationsResponse{ProductIds: f.ids}, nil
}

type fakeShippingService struct {
	pb.UnimplementedShippingServiceServer
	quoteUSD pb.Money
}

func (f *fakeShippingService) GetQuote(_ context.Context, _ *pb.GetQuoteRequest) (*pb.GetQuoteResponse, error) {
	return &pb.GetQuoteResponse{CostUsd: &f.quoteUSD}, nil
}

func (f *fakeShippingService) ShipOrder(_ context.Context, _ *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	return &pb.ShipOrderResponse{TrackingId: "TRACK-42"}, nil
}

type fakeCheckoutService struct {
	pb.UnimplementedCheckoutServiceServer
	mu     sync.Mutex
	orders []*pb.PlaceOrderRequest
}

func (f *fakeCheckoutService) PlaceOrder(_ context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.orders = append(f.orders, req)
	return &pb.PlaceOrderResponse{Order: &pb.OrderResult{
		OrderId:            "ORD-1",
		ShippingTrackingId: "TRACK-1",
		ShippingCost:       &pb.Money{CurrencyCode: req.GetUserCurrency(), Units: 10, Nanos: 0},
		ShippingAddress:    req.GetAddress(),
		Items: []*pb.OrderItem{
			{Item: &pb.CartItem{ProductId: "OLJCESPC7Z", Quantity: 2}, Cost: &pb.Money{CurrencyCode: req.GetUserCurrency(), Units: 19, Nanos: 990000000}},
		},
	}}, nil
}

type fakeAdService struct {
	pb.UnimplementedAdServiceServer
	ads []*pb.Ad
}

func (f *fakeAdService) GetAds(_ context.Context, _ *pb.AdRequest) (*pb.AdResponse, error) {
	return &pb.AdResponse{Ads: f.ads}, nil
}

type fakeAssistantService struct {
	pb.UnimplementedShoppingAssistantServiceServer
	reply string
}

func (f *fakeAssistantService) GetCompletion(_ context.Context, req *pb.ShoppingAssistantRequest) (*pb.ShoppingAssistantResponse, error) {
	return &pb.ShoppingAssistantResponse{
		Message:        f.reply + " " + req.GetMessage(),
		ConversationId: "conv-1",
	}, nil
}

// --- test harness ------------------------------------------------------------

// testCatalog is a small built-in catalog shared by the tests.
func testCatalog() []*pb.Product {
	return []*pb.Product{
		{Id: "OLJCESPC7Z", Name: "Sunglasses", Description: "Sleek aviator sunglasses", Picture: "/static/img/products/sunglasses.jpg",
			PriceUsd: &pb.Money{CurrencyCode: "USD", Units: 19, Nanos: 990000000}, Categories: []string{"accessories"}},
		{Id: "66VCHSJNUP", Name: "Tank Top", Description: "Cropped cotton tank", Picture: "/static/img/products/tank-top.jpg",
			PriceUsd: &pb.Money{CurrencyCode: "USD", Units: 18, Nanos: 990000000}, Categories: []string{"clothing"}},
	}
}

func startFakeServer(t *testing.T, register func(s *grpc.Server)) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	register(srv)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func dial(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// newTestFrontend wires a frontendServer to the given fake services and
// returns the fully configured service. Services passed as nil are pointed
// at a dead address so RPCs against them fail fast with Unavailable (the
// same behavior as when a downstream service is not running locally).
func newTestFrontend(t *testing.T, pc *fakeProductCatalogService, cart *fakeCartService, currency *fakeCurrencyService,
	rec *fakeRecommendationService, shipping *fakeShippingService, checkout *fakeCheckoutService, ad *fakeAdService) *frontendServer {
	t.Helper()
	fe := &frontendServer{}
	fe.productCatalogSvcConn = connOrDead(t, pc != nil, func(s *grpc.Server) { pb.RegisterProductCatalogServiceServer(s, pc) })
	fe.cartSvcConn = connOrDead(t, cart != nil, func(s *grpc.Server) { pb.RegisterCartServiceServer(s, cart) })
	fe.currencySvcConn = connOrDead(t, currency != nil, func(s *grpc.Server) { pb.RegisterCurrencyServiceServer(s, currency) })
	fe.recommendationSvcConn = connOrDead(t, rec != nil, func(s *grpc.Server) { pb.RegisterRecommendationServiceServer(s, rec) })
	fe.shippingSvcConn = connOrDead(t, shipping != nil, func(s *grpc.Server) { pb.RegisterShippingServiceServer(s, shipping) })
	fe.checkoutSvcConn = connOrDead(t, checkout != nil, func(s *grpc.Server) { pb.RegisterCheckoutServiceServer(s, checkout) })
	fe.adSvcConn = connOrDead(t, ad != nil, func(s *grpc.Server) { pb.RegisterAdServiceServer(s, ad) })
	return fe
}

// connOrDead returns a connection to a fresh fake gRPC server when wantFake
// is true; otherwise a connection to a dead address (RPCs fail fast with
// Unavailable).
func connOrDead(t *testing.T, wantFake bool, register func(s *grpc.Server)) *grpc.ClientConn {
	if wantFake {
		return dial(t, startFakeServer(t, register))
	}
	return dial(t, "127.0.0.1:1")
}

// startTestServer runs the real handler chain (routes + h2c + gRPC health +
// logging + session) on a random localhost port, exactly like main().
func startTestServer(t *testing.T, fe *frontendServer) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := &http.Server{Handler: h2c.NewHandler(newHandler(fe), &http2.Server{})}
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { _ = srv.Close() })
	return lis.Addr().String()
}

// newTestClient returns an HTTP client that keeps cookies (session/currency)
// but does not follow redirects (tests assert on the redirect target).
func newTestClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func get(t *testing.T, client *http.Client, url string) *http.Response {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	return resp
}

func postForm(t *testing.T, client *http.Client, url string, form map[string]string) *http.Response {
	t.Helper()
	var b strings.Builder
	for k, v := range form {
		if b.Len() > 0 {
			b.WriteString("&")
		}
		b.WriteString(k + "=" + v)
	}
	resp, err := client.Post(url, "application/x-www-form-urlencoded", strings.NewReader(b.String()))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	return string(body)
}

// --- tests --------------------------------------------------------------------

func TestHealthEndpoints(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	// HTTP health endpoint.
	resp := get(t, newTestClient(t), "http://"+addr+"/_healthz")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/_healthz status = %d, want 200", resp.StatusCode)
	}
	if body := readBody(t, resp); body != "ok" {
		t.Errorf("/_healthz body = %q, want ok", body)
	}

	// gRPC health over h2c on the same port (same transport grpcurl uses).
	conn := dial(t, addr)
	hc := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, svc := range []string{"", healthServiceName} {
		hr, err := hc.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: svc})
		if err != nil {
			t.Fatalf("health Check(%q) failed: %v", svc, err)
		}
		if hr.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Errorf("health Check(%q) = %v, want SERVING", svc, hr.GetStatus())
		}
	}
}

func TestHomePage(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	currency := &fakeCurrencyService{}
	ad := &fakeAdService{ads: []*pb.Ad{{Text: "Advertisement", RedirectUrl: "https://example.com"}}}
	fe := newTestFrontend(t, pc, cart, currency, nil, nil, nil, ad)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	for _, want := range []string{"Sunglasses", "Tank Top", "$19.99", "$18.99", "Online Boutique"} {
		if !strings.Contains(body, want) {
			t.Errorf("home page body missing %q", want)
		}
	}
}

func TestHomePageCurrencyFallback(t *testing.T) {
	// currencyservice not wired -> getCurrencies must fall back to the
	// whitelist and the default USD flow must not need a conversion RPC.
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	for _, want := range []string{"Sunglasses", "$19.99", "USD"} {
		if !strings.Contains(body, want) {
			t.Errorf("home page body missing %q", want)
		}
	}
}

func TestProductPage(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	rec := &fakeRecommendationService{ids: []string{"66VCHSJNUP"}}
	fe := newTestFrontend(t, pc, cart, nil, rec, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/product/OLJCESPC7Z")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /product/OLJCESPC7Z status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	for _, want := range []string{"Sunglasses", "Sleek aviator sunglasses", "$19.99", "Add To Cart"} {
		if !strings.Contains(body, want) {
			t.Errorf("product page body missing %q", want)
		}
	}
}

func TestProductPageNotFound(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/product/NOPE")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("GET /product/NOPE status = %d, want 500", resp.StatusCode)
	}
}

func TestAddToCart(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	resp := postForm(t, client, "http://"+addr+"/cart", map[string]string{"product_id": "OLJCESPC7Z", "quantity": "2"})
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("POST /cart status = %d, want 302; body=%s", resp.StatusCode, readBody(t, resp))
	}
	if loc := resp.Header.Get("location"); loc != "/cart" {
		t.Errorf("POST /cart location = %q, want /cart", loc)
	}

	cart.mu.Lock()
	defer cart.mu.Unlock()
	if len(cart.items) != 1 {
		t.Fatalf("cart items = %d, want 1", len(cart.items))
	}
	if cart.items[0].GetProductId() != "OLJCESPC7Z" || cart.items[0].GetQuantity() != 2 {
		t.Errorf("cart item = %v, want OLJCESPC7Z x2", cart.items[0])
	}
	if cart.userID == "" {
		t.Error("cart AddItem user_id is empty, want the session id")
	}
}

func TestAddToCartInvalid(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	for name, form := range map[string]map[string]string{
		"zero quantity":    {"product_id": "OLJCESPC7Z", "quantity": "0"},
		"over 10 quantity": {"product_id": "OLJCESPC7Z", "quantity": "11"},
		"missing product":  {"product_id": "", "quantity": "1"},
	} {
		t.Run(name, func(t *testing.T) {
			resp := postForm(t, client, "http://"+addr+"/cart", form)
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("POST /cart status = %d, want 422", resp.StatusCode)
			}
		})
	}
}

func TestEmptyCart(t *testing.T) {
	cart := &fakeCartService{items: []*pb.CartItem{{ProductId: "OLJCESPC7Z", Quantity: 1}}}
	fe := newTestFrontend(t, nil, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := postForm(t, newTestClient(t), "http://"+addr+"/cart/empty", nil)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("POST /cart/empty status = %d, want 302; body=%s", resp.StatusCode, readBody(t, resp))
	}
	if loc := resp.Header.Get("location"); loc != "/" {
		t.Errorf("POST /cart/empty location = %q, want /", loc)
	}
	cart.mu.Lock()
	defer cart.mu.Unlock()
	if !cart.emptied {
		t.Error("cart was not emptied")
	}
}

func TestViewCart(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{items: []*pb.CartItem{{ProductId: "OLJCESPC7Z", Quantity: 2}}}
	shipping := &fakeShippingService{quoteUSD: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 0}}
	fe := newTestFrontend(t, pc, cart, nil, nil, shipping, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/cart")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /cart status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	for _, want := range []string{"Sunglasses", "Quantity: 2", "$10.00", "$49.98"} {
		if !strings.Contains(body, want) {
			t.Errorf("cart page body missing %q", want)
		}
	}
}

func TestSetCurrency(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	currency := &fakeCurrencyService{}
	fe := newTestFrontend(t, pc, cart, currency, nil, nil, nil, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	req, _ := http.NewRequest("POST", "http://"+addr+"/setCurrency", strings.NewReader("currency_code=EUR"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/product/OLJCESPC7Z")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /setCurrency failed: %v", err)
	}
	readBody(t, resp)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("POST /setCurrency status = %d, want 302", resp.StatusCode)
	}
	if loc := resp.Header.Get("location"); loc != "/product/OLJCESPC7Z" {
		t.Errorf("POST /setCurrency location = %q, want referer", loc)
	}

	// The cookie must be persisted; the home page then renders EUR prices
	// via the (identity) currency service.
	home := get(t, client, "http://"+addr+"/")
	homeBody := readBody(t, home)
	if !strings.Contains(homeBody, "€19.99") {
		t.Errorf("home page after switching to EUR missing €19.99; body=%s", homeBody)
	}
}

func TestSetCurrencyInvalid(t *testing.T) {
	fe := newTestFrontend(t, nil, nil, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := postForm(t, newTestClient(t), "http://"+addr+"/setCurrency", map[string]string{"currency_code": "xx"})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("POST /setCurrency status = %d, want 422", resp.StatusCode)
	}
}

func TestPlaceOrder(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	checkout := &fakeCheckoutService{}
	currency := &fakeCurrencyService{}
	fe := newTestFrontend(t, pc, cart, currency, nil, nil, checkout, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	form := map[string]string{
		"email":                        "buyer@example.com",
		"street_address":               "1600 Amphitheatre Parkway",
		"zip_code":                     "94043",
		"city":                         "Mountain View",
		"state":                        "CA",
		"country":                      "United States",
		"credit_card_number":           "4432801561520454",
		"credit_card_expiration_month": "1",
		"credit_card_expiration_year":  "2030",
		"credit_card_cvv":              "672",
	}
	resp := postForm(t, client, "http://"+addr+"/cart/checkout", form)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /cart/checkout status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	for _, want := range []string{"Your order is complete!", "ORD-1", "TRACK-1", "$49.98"} {
		if !strings.Contains(body, want) {
			t.Errorf("order page body missing %q", want)
		}
	}

	checkout.mu.Lock()
	defer checkout.mu.Unlock()
	if len(checkout.orders) != 1 {
		t.Fatalf("PlaceOrder called %d times, want 1", len(checkout.orders))
	}
	order := checkout.orders[0]
	if order.GetEmail() != "buyer@example.com" {
		t.Errorf("order email = %q, want buyer@example.com", order.GetEmail())
	}
	if order.GetUserCurrency() != "USD" {
		t.Errorf("order user_currency = %q, want USD", order.GetUserCurrency())
	}
	if order.GetAddress().GetCity() != "Mountain View" || order.GetAddress().GetZipCode() != 94043 {
		t.Errorf("order address = %+v, want Mountain View/94043", order.GetAddress())
	}
	if order.GetCreditCard().GetCreditCardNumber() != "4432801561520454" {
		t.Errorf("order cc number = %q, want 4432801561520454", order.GetCreditCard().GetCreditCardNumber())
	}
	if order.GetUserId() == "" {
		t.Error("order user_id is empty, want the session id")
	}
}

func TestPlaceOrderValidation(t *testing.T) {
	fe := newTestFrontend(t, nil, nil, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	bad := map[string]string{
		"email":                        "not-an-email",
		"street_address":               "",
		"zip_code":                     "",
		"city":                         "",
		"state":                        "",
		"country":                      "",
		"credit_card_number":           "",
		"credit_card_expiration_month": "0",
		"credit_card_expiration_year":  "0",
		"credit_card_cvv":              "0",
	}
	resp := postForm(t, client, "http://"+addr+"/cart/checkout", bad)
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("POST /cart/checkout status = %d, want 422; body=%s", resp.StatusCode, body)
	}
	if !strings.Contains(body, "Field &#39;Email&#39; is invalid") {
		t.Errorf("validation error page missing Field error; body=%s", body)
	}
}

func TestProductMetaJSON(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	fe := newTestFrontend(t, pc, nil, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/product-meta/OLJCESPC7Z")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /product-meta status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	var p pb.Product
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatalf("product-meta response is not valid JSON: %v; body=%s", err, body)
	}
	if p.GetId() != "OLJCESPC7Z" || p.GetName() != "Sunglasses" {
		t.Errorf("product-meta = %+v, want OLJCESPC7Z/Sunglasses", &p)
	}
}

func TestChatBot(t *testing.T) {
	assistant := &fakeAssistantService{reply: "You might like"}
	fe := &frontendServer{}
	fe.shoppingAssistantSvcConn = dial(t, startFakeServer(t, func(s *grpc.Server) { pb.RegisterShoppingAssistantServiceServer(s, assistant) }))
	addr := startTestServer(t, fe)

	req, _ := http.NewRequest("POST", "http://"+addr+"/bot", strings.NewReader(`{"message":"recommend a gift"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := newTestClient(t).Do(req)
	if err != nil {
		t.Fatalf("POST /bot failed: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /bot status = %d, want 200; body=%s", resp.StatusCode, body)
	}
	var reply struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversation_id"`
	}
	if err := json.Unmarshal([]byte(body), &reply); err != nil {
		t.Fatalf("bot response is not valid JSON: %v; body=%s", err, body)
	}
	if !strings.Contains(reply.Message, "recommend a gift") {
		t.Errorf("bot message = %q, want it to echo the question", reply.Message)
	}
	if reply.ConversationID != "conv-1" {
		t.Errorf("bot conversation_id = %q, want conv-1", reply.ConversationID)
	}
}

func TestSessionIDCookie(t *testing.T) {
	pc := &fakeProductCatalogService{products: testCatalog()}
	cart := &fakeCartService{}
	fe := newTestFrontend(t, pc, cart, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)
	client := newTestClient(t)

	// First request assigns a session cookie.
	resp := get(t, client, "http://"+addr+"/_healthz")
	var session string
	for _, c := range resp.Cookies() {
		if c.Name == cookieSessionID {
			session = c.Value
		}
	}
	readBody(t, resp)
	if session == "" {
		t.Fatal("no session cookie set on first request")
	}

	// The cart is keyed by session id: with the same cookie jar, the
	// AddItem call must carry the session id from the cookie.
	postForm(t, client, "http://"+addr+"/cart", map[string]string{"product_id": "OLJCESPC7Z", "quantity": "1"})

	cart.mu.Lock()
	defer cart.mu.Unlock()
	if cart.userID != session {
		t.Errorf("cart user_id = %q, want session %q", cart.userID, session)
	}
}

func TestStaticFiles(t *testing.T) {
	fe := newTestFrontend(t, nil, nil, nil, nil, nil, nil, nil)
	addr := startTestServer(t, fe)

	resp := get(t, newTestClient(t), "http://"+addr+"/static/styles/styles.css")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /static/styles/styles.css status = %d, want 200", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "body") {
		t.Errorf("styles.css body does not look like css; got %d bytes", len(body))
	}

	resp = get(t, newTestClient(t), "http://"+addr+"/static/img/products/sunglasses.jpg")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /static/img/products/sunglasses.jpg status = %d, want 200", resp.StatusCode)
	}
}
