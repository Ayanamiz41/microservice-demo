package main

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// --- in-memory fakes for every downstream service ---------------------------

type fakeCartService struct {
	pb.UnimplementedCartServiceServer
	items     []*pb.CartItem
	emptied   bool
	emptiedMu sync.Mutex
}

func (f *fakeCartService) GetCart(_ context.Context, _ *pb.GetCartRequest) (*pb.Cart, error) {
	return &pb.Cart{Items: f.items}, nil
}

func (f *fakeCartService) EmptyCart(_ context.Context, _ *pb.EmptyCartRequest) (*pb.Empty, error) {
	f.emptiedMu.Lock()
	defer f.emptiedMu.Unlock()
	f.emptied = true
	return &pb.Empty{}, nil
}

type fakeProductCatalogService struct {
	pb.UnimplementedProductCatalogServiceServer
	products map[string]*pb.Product
}

func (f *fakeProductCatalogService) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	p, ok := f.products[req.GetId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "product %q not found", req.GetId())
	}
	return p, nil
}

// fakeCurrencyService is an identity converter: it returns the input amount
// tagged with the requested target currency code.
type fakeCurrencyService struct {
	pb.UnimplementedCurrencyServiceServer
}

func (f *fakeCurrencyService) Convert(_ context.Context, req *pb.CurrencyConversionRequest) (*pb.Money, error) {
	return &pb.Money{
		CurrencyCode: req.GetToCode(),
		Units:        req.GetFrom().GetUnits(),
		Nanos:        req.GetFrom().GetNanos(),
	}, nil
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

type fakePaymentService struct {
	pb.UnimplementedPaymentServiceServer
	chargedMu sync.Mutex
	charged   []*pb.Money
}

func (f *fakePaymentService) Charge(_ context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, error) {
	f.chargedMu.Lock()
	defer f.chargedMu.Unlock()
	f.charged = append(f.charged, req.GetAmount())
	return &pb.ChargeResponse{TransactionId: "TXN-42"}, nil
}

type fakeEmailService struct {
	pb.UnimplementedEmailServiceServer
	sentMu sync.Mutex
	sent   []*pb.SendOrderConfirmationRequest
}

func (f *fakeEmailService) SendOrderConfirmation(_ context.Context, req *pb.SendOrderConfirmationRequest) (*pb.Empty, error) {
	f.sentMu.Lock()
	defer f.sentMu.Unlock()
	f.sent = append(f.sent, req)
	return &pb.Empty{}, nil
}

// startFakeServer runs a gRPC server on a random localhost port.
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

func startTestService(t *testing.T, svc *checkoutService) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterCheckoutServiceServer(srv, svc)
	// Same health registration as main(): explicit SERVING status for the
	// canonical service name.
	healthcheck := health.NewServer()
	healthcheck.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, healthcheck)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

func TestHealthCheckServing(t *testing.T) {
	addr := startTestService(t, &checkoutService{})
	conn := dial(t, addr)
	hc := grpc_health_v1.NewHealthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := hc.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: healthServiceName})
	if err != nil {
		t.Fatalf("health Check(%q) failed: %v", healthServiceName, err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("health Check(%q) = %v, want SERVING", healthServiceName, resp.GetStatus())
	}
}

func TestPlaceOrderFullFlow(t *testing.T) {
	// Downstream fakes.
	cart := &fakeCartService{items: []*pb.CartItem{
		{ProductId: "OLJCESPC7Z", Quantity: 2},
		{ProductId: "L9ECAV7KIM", Quantity: 1},
	}}
	pc := &fakeProductCatalogService{products: map[string]*pb.Product{
		"OLJCESPC7Z": {Id: "OLJCESPC7Z", Name: "Vintage Typewriter", PriceUsd: &pb.Money{CurrencyCode: "USD", Units: 1, Nanos: 500000000}}, // $1.50
		"L9ECAV7KIM": {Id: "L9ECAV7KIM", Name: "Sunglasses", PriceUsd: &pb.Money{CurrencyCode: "USD", Units: 2, Nanos: 0}},                 // $2.00
	}}
	currency := &fakeCurrencyService{}
	shipping := &fakeShippingService{quoteUSD: pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 0}} // $10.00
	payment := &fakePaymentService{}
	email := &fakeEmailService{}

	cartAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterCartServiceServer(s, cart) })
	pcAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterProductCatalogServiceServer(s, pc) })
	currencyAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterCurrencyServiceServer(s, currency) })
	shippingAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterShippingServiceServer(s, shipping) })
	paymentAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterPaymentServiceServer(s, payment) })
	emailAddr := startFakeServer(t, func(s *grpc.Server) { pb.RegisterEmailServiceServer(s, email) })

	svc := &checkoutService{
		cartSvcConn:           dial(t, cartAddr),
		productCatalogSvcConn: dial(t, pcAddr),
		currencySvcConn:       dial(t, currencyAddr),
		shippingSvcConn:       dial(t, shippingAddr),
		paymentSvcConn:        dial(t, paymentAddr),
		emailSvcConn:          dial(t, emailAddr),
	}
	addr := startTestService(t, svc)
	conn := dial(t, addr)
	client := pb.NewCheckoutServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.PlaceOrderRequest{
		UserId:       "user-1",
		UserCurrency: "USD",
		Address: &pb.Address{
			StreetAddress: "1 Main St",
			City:          "Springfield",
			State:         "IL",
			Country:       "US",
			ZipCode:       62701,
		},
		Email:      "buyer@example.com",
		CreditCard: &pb.CreditCardInfo{CreditCardNumber: "4432-8015-6152-0454", CreditCardCvv: 672, CreditCardExpirationYear: 2030, CreditCardExpirationMonth: 1},
	}

	resp, err := client.PlaceOrder(ctx, req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	order := resp.GetOrder()
	if order.GetOrderId() == "" {
		t.Error("order_id is empty")
	}
	if order.GetShippingTrackingId() != "TRACK-42" {
		t.Errorf("shipping_tracking_id = %q, want TRACK-42", order.GetShippingTrackingId())
	}
	wantShipping := &pb.Money{CurrencyCode: "USD", Units: 10, Nanos: 0}
	if !moneyEqual(order.GetShippingCost(), wantShipping) {
		t.Errorf("shipping_cost = %v, want %v", order.GetShippingCost(), wantShipping)
	}
	if order.GetShippingAddress().GetCity() != "Springfield" {
		t.Errorf("shipping_address city = %q, want Springfield", order.GetShippingAddress().GetCity())
	}
	if len(order.GetItems()) != 2 {
		t.Fatalf("order items = %d, want 2", len(order.GetItems()))
	}
	// OrderItem.Cost is the per-unit converted price (quantity is applied only
	// when computing the total): item 1 = $1.50, item 2 = $2.00.
	wantCosts := []*pb.Money{
		{CurrencyCode: "USD", Units: 1, Nanos: 500000000},
		{CurrencyCode: "USD", Units: 2, Nanos: 0},
	}
	for i, want := range wantCosts {
		if !moneyEqual(order.GetItems()[i].GetCost(), want) {
			t.Errorf("order items[%d].cost = %v, want %v", i, order.GetItems()[i].GetCost(), want)
		}
	}

	// Total charged = items ($5.00) + shipping ($10.00) = $15.00.
	payment.chargedMu.Lock()
	charged := append([]*pb.Money(nil), payment.charged...)
	payment.chargedMu.Unlock()
	if len(charged) != 1 {
		t.Fatalf("payment charged %d times, want 1", len(charged))
	}
	wantTotal := &pb.Money{CurrencyCode: "USD", Units: 15, Nanos: 0}
	if !moneyEqual(charged[0], wantTotal) {
		t.Errorf("charged amount = %v, want %v", charged[0], wantTotal)
	}

	if !cart.emptied {
		t.Error("cart was not emptied after checkout")
	}

	email.sentMu.Lock()
	sent := append([]*pb.SendOrderConfirmationRequest(nil), email.sent...)
	email.sentMu.Unlock()
	if len(sent) != 1 {
		t.Fatalf("email sent %d times, want 1", len(sent))
	}
	if sent[0].GetEmail() != "buyer@example.com" {
		t.Errorf("confirmation email = %q, want buyer@example.com", sent[0].GetEmail())
	}
	if sent[0].GetOrder().GetOrderId() != order.GetOrderId() {
		t.Errorf("confirmation order_id mismatch: %q vs %q", sent[0].GetOrder().GetOrderId(), order.GetOrderId())
	}
}

func TestPlaceOrderCartFailure(t *testing.T) {
	// cart service down -> PlaceOrder must fail with an Internal error,
	// not panic or hang.
	svc := &checkoutService{
		cartSvcConn:           dial(t, "127.0.0.1:1"), // nothing listening
		productCatalogSvcConn: dial(t, "127.0.0.1:1"),
		currencySvcConn:       dial(t, "127.0.0.1:1"),
		shippingSvcConn:       dial(t, "127.0.0.1:1"),
		paymentSvcConn:        dial(t, "127.0.0.1:1"),
		emailSvcConn:          dial(t, "127.0.0.1:1"),
	}
	addr := startTestService(t, svc)
	conn := dial(t, addr)
	client := pb.NewCheckoutServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
		UserId:       "user-2",
		UserCurrency: "USD",
		Address:      &pb.Address{City: "X"},
	})
	if err == nil {
		t.Fatal("PlaceOrder succeeded, want error when cart service is unavailable")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("PlaceOrder error code = %v, want Internal; err=%v", status.Code(err), err)
	}
}

func moneyEqual(l, r *pb.Money) bool {
	if l == nil || r == nil {
		return l == r
	}
	return l.GetCurrencyCode() == r.GetCurrencyCode() &&
		l.GetUnits() == r.GetUnits() && l.GetNanos() == r.GetNanos()
}
