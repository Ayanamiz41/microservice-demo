// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"net"
	"regexp"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// TestGetQuote is a basic check on the GetQuote RPC service.
func TestGetQuote(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: &pb.Address{
			StreetAddress: "Muffin Man",
			City:          "London",
			State:         "",
			Country:       "England",
		},
		Items: []*pb.CartItem{
			{ProductId: "23", Quantity: 1},
			{ProductId: "46", Quantity: 3},
		},
	}

	res, err := s.GetQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("TestGetQuote (%v) failed", err)
	}
	if res.CostUsd.GetUnits() != 8 || res.CostUsd.GetNanos() != 990000000 {
		t.Errorf("TestGetQuote: quote value '%d.%d' does not match expected '8.990000000'",
			res.CostUsd.GetUnits(), res.CostUsd.GetNanos())
	}
	if res.CostUsd.GetCurrencyCode() != "USD" {
		t.Errorf("TestGetQuote: currency code '%s', want 'USD'", res.CostUsd.GetCurrencyCode())
	}
}

// TestGetQuoteEmptyCart verifies that an empty cart returns a zero quote.
func TestGetQuoteEmptyCart(t *testing.T) {
	s := server{}

	req := &pb.GetQuoteRequest{
		Address: &pb.Address{
			StreetAddress: "221B Baker Street",
			City:          "London",
			State:         "",
			Country:       "England",
		},
		Items: []*pb.CartItem{},
	}

	res, err := s.GetQuote(context.Background(), req)
	if err != nil {
		t.Fatalf("TestGetQuoteEmptyCart (%v) failed", err)
	}
	if res.CostUsd.GetUnits() != 0 || res.CostUsd.GetNanos() != 0 {
		t.Errorf("TestGetQuoteEmptyCart: expected zero quote for empty cart, got '%d.%d'",
			res.CostUsd.GetUnits(), res.CostUsd.GetNanos())
	}
}

// TestShipOrder is a basic check on the ShipOrder RPC service.
func TestShipOrder(t *testing.T) {
	s := server{}

	req := &pb.ShipOrderRequest{
		Address: &pb.Address{
			StreetAddress: "Muffin Man",
			City:          "London",
			State:         "",
			Country:       "England",
		},
		Items: []*pb.CartItem{
			{ProductId: "23", Quantity: 1},
			{ProductId: "46", Quantity: 3},
		},
	}

	res, err := s.ShipOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("TestShipOrder (%v) failed", err)
	}
	if len(res.TrackingId) != 18 {
		t.Errorf("TestShipOrder: tracking ID is malformed - has %d characters, %d expected",
			len(res.TrackingId), 18)
	}
}

// TestTrackingIdFormat verifies the tracking ID matches the expected pattern.
func TestTrackingIdFormat(t *testing.T) {
	pattern := regexp.MustCompile(`^[A-Z]{2}-\d+-\d+$`)

	for i := 0; i < 20; i++ {
		id := CreateTrackingId("test-salt-value")
		if !pattern.MatchString(id) {
			t.Errorf("CreateTrackingId: '%s' does not match expected pattern '[A-Z]{2}-\\d+-\\d+'", id)
		}
	}
}

// TestTrackingIdUniqueness checks that generated IDs are not all identical.
func TestTrackingIdUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		id := CreateTrackingId("same-salt")
		seen[id] = true
	}
	if len(seen) < 2 {
		t.Errorf("CreateTrackingId: expected unique IDs but got %d distinct values out of 50", len(seen))
	}
}

// TestCreateQuoteFromCount verifies count-based quote generation using
// integer-only money arithmetic.
func TestCreateQuoteFromCount(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		dollars uint32
		cents   uint32
	}{
		{"empty cart", 0, 0, 0},
		{"negative count", -3, 0, 0},
		{"single item", 1, 8, 99},
		{"many items", 42, 8, 99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := CreateQuoteFromCount(tc.count)
			if q.Dollars != tc.dollars || q.Cents != tc.cents {
				t.Errorf("CreateQuoteFromCount(%d) = $%d.%02d, want $%d.%02d",
					tc.count, q.Dollars, q.Cents, tc.dollars, tc.cents)
			}
		})
	}
}

// TestQuoteString verifies the string representation of a Quote.
func TestQuoteString(t *testing.T) {
	tests := []struct {
		name string
		q    Quote
		want string
	}{
		{"zero", Quote{}, "$0.00"},
		{"flat fee", Quote{Dollars: 8, Cents: 99}, "$8.99"},
		{"half dollar", Quote{Dollars: 0, Cents: 50}, "$0.50"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.q.String(); got != tc.want {
				t.Errorf("Quote.String() = '%s', want '%s'", got, tc.want)
			}
		})
	}
}

// TestQuoteToMoney verifies the Quote -> hipstershop.Money conversion stays
// integral: 1 cent == 10^7 nanos.
func TestQuoteToMoney(t *testing.T) {
	tests := []struct {
		name  string
		q     Quote
		units int64
		nanos int32
	}{
		{"zero", Quote{}, 0, 0},
		{"flat fee", Quote{Dollars: 8, Cents: 99}, 8, 990000000},
		{"half dollar", Quote{Dollars: 0, Cents: 50}, 0, 500000000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.q.ToMoney()
			if m.GetCurrencyCode() != "USD" {
				t.Errorf("ToMoney() currency = '%s', want 'USD'", m.GetCurrencyCode())
			}
			if m.GetUnits() != tc.units || m.GetNanos() != tc.nanos {
				t.Errorf("ToMoney() = %d.%09d, want %d.%09d",
					m.GetUnits(), m.GetNanos(), tc.units, tc.nanos)
			}
		})
	}
}

// TestGetRandomLetterCode verifies the output is a valid uppercase letter.
func TestGetRandomLetterCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := getRandomLetterCode()
		if code < 65 || code > 90 {
			t.Errorf("getRandomLetterCode: got %d (%c), expected range 65-90 (A-Z)", code, code)
		}
	}
}

// TestGetRandomNumber verifies the output has the correct number of digits.
func TestGetRandomNumber(t *testing.T) {
	for _, digits := range []int{1, 3, 5, 7, 10} {
		result := getRandomNumber(digits)
		if len(result) != digits {
			t.Errorf("getRandomNumber(%d) = '%s' (len %d), expected length %d",
				digits, result, len(result), digits)
		}
	}
}

// newTestGRPCServer starts a real in-process gRPC server (shipping service +
// health service) and returns a connected client for it.
func newTestGRPCServer(t *testing.T) (*grpc.ClientConn, *health.Server) {
	t.Helper()

	srv := grpc.NewServer()
	pb.RegisterShippingServiceServer(srv, &server{})
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthcheck)
	healthcheck.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Errorf("test gRPC server failed: %v", err)
		}
	}()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn, healthcheck
}

// TestHealthCheckOverGRPC verifies grpc.health.v1.Health/Check returns
// SERVING for the service name documented in the src/README.md matrix.
func TestHealthCheckOverGRPC(t *testing.T) {
	conn, _ := newTestGRPCServer(t)

	client := healthpb.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &healthpb.HealthCheckRequest{Service: serviceName})
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Errorf("health check status = %v, want SERVING", resp.GetStatus())
	}

	// The overall status (empty service name) must also report SERVING.
	resp, err = client.Check(context.Background(), &healthpb.HealthCheckRequest{Service: ""})
	if err != nil {
		t.Fatalf("overall health check failed: %v", err)
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		t.Errorf("overall health check status = %v, want SERVING", resp.GetStatus())
	}
}

// TestGetQuoteOverGRPC exercises the full RPC path end-to-end over an
// in-process gRPC server, mirroring the smoke test in the README.
func TestGetQuoteOverGRPC(t *testing.T) {
	conn, _ := newTestGRPCServer(t)

	client := pb.NewShippingServiceClient(conn)
	resp, err := client.GetQuote(context.Background(), &pb.GetQuoteRequest{
		Address: &pb.Address{StreetAddress: "Muffin Man", City: "London", Country: "England"},
		Items: []*pb.CartItem{
			{ProductId: "23", Quantity: 1},
			{ProductId: "46", Quantity: 3},
		},
	})
	if err != nil {
		t.Fatalf("GetQuote over gRPC failed: %v", err)
	}
	if resp.GetCostUsd().GetUnits() != 8 || resp.GetCostUsd().GetNanos() != 990000000 {
		t.Errorf("GetQuote over gRPC: got '%d.%d', want '8.990000000'",
			resp.GetCostUsd().GetUnits(), resp.GetCostUsd().GetNanos())
	}
}
