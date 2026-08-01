// checkoutservice — the CheckoutService for the Online Boutique replica.
//
// Wire-compatible with upstream GoogleCloudPlatform/microservices-demo
// src/checkoutservice. Implements the PlaceOrder/OrderResult flow:
// cart -> product catalog + currency -> shipping quote -> payment -> shipping
// -> empty cart -> email confirmation.
//
// All money math goes through the money package (integer units + nanos,
// no floating point).
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
	money "github.com/Ayanamiz41/microservice-demo/src/checkoutservice/money"
)

const (
	listenPort = "5050"

	// healthServiceName is the gRPC service name the health server is
	// registered under; grpcurl probes use it:
	//   grpcurl -plaintext -d '{"service":"hipstershop.CheckoutService"}' \
	//     localhost:5050 grpc.health.v1.Health/Check
	healthServiceName = "hipstershop.CheckoutService"
)

// Local-first defaults for the downstream services (matches src/README.md
// port matrix) so `go run ./...` starts standalone with no env vars set.
// In a deployment these are overridden via env vars.
const (
	defaultProductCatalogAddr = "localhost:3550"
	defaultCartAddr           = "localhost:7070"
	defaultCurrencyAddr       = "localhost:7000"
	defaultShippingAddr       = "localhost:50051"
	defaultEmailAddr          = "localhost:8080"
	defaultPaymentAddr        = "localhost:50051"
)

type checkoutService struct {
	pb.UnimplementedCheckoutServiceServer

	productCatalogSvcConn *grpc.ClientConn
	cartSvcConn           *grpc.ClientConn
	currencySvcConn       *grpc.ClientConn
	shippingSvcConn       *grpc.ClientConn
	emailSvcConn          *grpc.ClientConn
	paymentSvcConn        *grpc.ClientConn
}

func main() {
	port := envOr("PORT", listenPort)

	svc := &checkoutService{
		productCatalogSvcConn: mustConnGRPC(envOr("PRODUCT_CATALOG_SERVICE_ADDR", defaultProductCatalogAddr)),
		cartSvcConn:           mustConnGRPC(envOr("CART_SERVICE_ADDR", defaultCartAddr)),
		currencySvcConn:       mustConnGRPC(envOr("CURRENCY_SERVICE_ADDR", defaultCurrencyAddr)),
		shippingSvcConn:       mustConnGRPC(envOr("SHIPPING_SERVICE_ADDR", defaultShippingAddr)),
		emailSvcConn:          mustConnGRPC(envOr("EMAIL_SERVICE_ADDR", defaultEmailAddr)),
		paymentSvcConn:        mustConnGRPC(envOr("PAYMENT_SERVICE_ADDR", defaultPaymentAddr)),
	}
	log.Printf("service config: %+v", svc)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on :%s: %v", port, err)
	}

	srv := grpc.NewServer()
	pb.RegisterCheckoutServiceServer(srv, svc)
	healthcheck := health.NewServer()
	healthcheck.SetServingStatus(healthServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(srv, healthcheck)

	log.Printf("checkoutservice starting to listen on tcp: %q", lis.Addr().String())
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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

// PlaceOrder runs the full checkout flow and returns an OrderResult.
func (cs *checkoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	log.Printf("[PlaceOrder] user_id=%q user_currency=%q", req.GetUserId(), req.GetUserCurrency())

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate order uuid: %v", err)
	}

	prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(ctx, req.GetUserId(), req.GetUserCurrency(), req.GetAddress())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	total := pb.Money{CurrencyCode: req.GetUserCurrency()}
	total = money.Must(money.Sum(total, *prep.shippingCostLocalized))
	for _, it := range prep.orderItems {
		multPrice := money.MultiplySlow(*it.GetCost(), uint32(it.GetItem().GetQuantity()))
		total = money.Must(money.Sum(total, multPrice))
	}

	txID, err := cs.chargeCard(ctx, &total, req.GetCreditCard())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to charge card: %v", err)
	}
	log.Printf("payment went through (transaction_id: %s)", txID)

	shippingTrackingID, err := cs.shipOrder(ctx, req.GetAddress(), prep.cartItems)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "shipping error: %v", err)
	}

	if err := cs.emptyUserCart(ctx, req.GetUserId()); err != nil {
		log.Printf("warning: failed to empty user cart: %v", err)
	}

	orderResult := &pb.OrderResult{
		OrderId:            orderID.String(),
		ShippingTrackingId: shippingTrackingID,
		ShippingCost:       prep.shippingCostLocalized,
		ShippingAddress:    req.GetAddress(),
		Items:              prep.orderItems,
	}

	if err := cs.sendOrderConfirmation(ctx, req.GetEmail(), orderResult); err != nil {
		log.Printf("warning: failed to send order confirmation to %q: %v", req.GetEmail(), err)
	} else {
		log.Printf("order confirmation email sent to %q", req.GetEmail())
	}
	return &pb.PlaceOrderResponse{Order: orderResult}, nil
}

type orderPrep struct {
	orderItems            []*pb.OrderItem
	cartItems             []*pb.CartItem
	shippingCostLocalized *pb.Money
}

func (cs *checkoutService) prepareOrderItemsAndShippingQuoteFromCart(ctx context.Context, userID, userCurrency string, address *pb.Address) (orderPrep, error) {
	var out orderPrep

	cartItems, err := cs.getUserCart(ctx, userID)
	if err != nil {
		return out, fmt.Errorf("cart failure: %v", err)
	}
	orderItems, err := cs.prepOrderItems(ctx, cartItems, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to prepare order: %v", err)
	}
	shippingUSD, err := cs.quoteShipping(ctx, address, cartItems)
	if err != nil {
		return out, fmt.Errorf("shipping quote failure: %v", err)
	}
	shippingPrice, err := cs.convertCurrency(ctx, shippingUSD, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to convert shipping cost to currency: %v", err)
	}

	out.shippingCostLocalized = shippingPrice
	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (cs *checkoutService) getUserCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	cart, err := pb.NewCartServiceClient(cs.cartSvcConn).GetCart(ctx, &pb.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart during checkout: %v", err)
	}
	return cart.GetItems(), nil
}

func (cs *checkoutService) emptyUserCart(ctx context.Context, userID string) error {
	if _, err := pb.NewCartServiceClient(cs.cartSvcConn).EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userID}); err != nil {
		return fmt.Errorf("failed to empty user cart during checkout: %v", err)
	}
	return nil
}

func (cs *checkoutService) prepOrderItems(ctx context.Context, items []*pb.CartItem, userCurrency string) ([]*pb.OrderItem, error) {
	out := make([]*pb.OrderItem, len(items))
	cl := pb.NewProductCatalogServiceClient(cs.productCatalogSvcConn)

	for i, item := range items {
		product, err := cl.GetProduct(ctx, &pb.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			return nil, fmt.Errorf("failed to get product %q", item.GetProductId())
		}
		price, err := cs.convertCurrency(ctx, product.GetPriceUsd(), userCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to convert price of %q to %s", item.GetProductId(), userCurrency)
		}
		out[i] = &pb.OrderItem{
			Item: item,
			Cost: price,
		}
	}
	return out, nil
}

func (cs *checkoutService) quoteShipping(ctx context.Context, address *pb.Address, items []*pb.CartItem) (*pb.Money, error) {
	shippingQuote, err := pb.NewShippingServiceClient(cs.shippingSvcConn).
		GetQuote(ctx, &pb.GetQuoteRequest{Address: address, Items: items})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping quote: %v", err)
	}
	return shippingQuote.GetCostUsd(), nil
}

func (cs *checkoutService) convertCurrency(ctx context.Context, from *pb.Money, toCurrency string) (*pb.Money, error) {
	result, err := pb.NewCurrencyServiceClient(cs.currencySvcConn).Convert(ctx, &pb.CurrencyConversionRequest{
		From:   from,
		ToCode: toCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %v", err)
	}
	return result, nil
}

func (cs *checkoutService) chargeCard(ctx context.Context, amount *pb.Money, paymentInfo *pb.CreditCardInfo) (string, error) {
	paymentResp, err := pb.NewPaymentServiceClient(cs.paymentSvcConn).Charge(ctx, &pb.ChargeRequest{
		Amount:     amount,
		CreditCard: paymentInfo,
	})
	if err != nil {
		return "", fmt.Errorf("could not charge the card: %v", err)
	}
	return paymentResp.GetTransactionId(), nil
}

func (cs *checkoutService) shipOrder(ctx context.Context, address *pb.Address, items []*pb.CartItem) (string, error) {
	resp, err := pb.NewShippingServiceClient(cs.shippingSvcConn).ShipOrder(ctx, &pb.ShipOrderRequest{
		Address: address,
		Items:   items,
	})
	if err != nil {
		return "", fmt.Errorf("shipment failed: %v", err)
	}
	return resp.GetTrackingId(), nil
}

func (cs *checkoutService) sendOrderConfirmation(ctx context.Context, email string, order *pb.OrderResult) error {
	_, err := pb.NewEmailServiceClient(cs.emailSvcConn).SendOrderConfirmation(ctx, &pb.SendOrderConfirmationRequest{
		Email: email,
		Order: order,
	})
	return err
}
