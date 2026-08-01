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

// Command shippingservice implements the hipstershop.ShippingService gRPC
// server for the Online Boutique replica (microservice-demo). It is
// wire-compatible with the upstream GoogleCloudPlatform/microservices-demo
// shippingservice.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

const (
	defaultPort = "50051"
	// serviceName is the gRPC health check service name for this server
	// (see the src/README.md service matrix).
	serviceName = "hipstershop.ShippingService"
)

func main() {
	port := defaultPort
	if value, ok := os.LookupEnv("PORT"); ok && value != "" {
		port = value
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	srv := grpc.NewServer()
	pb.RegisterShippingServiceServer(srv, &server{})

	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthcheck)
	healthcheck.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Register the reflection service to ease grpcurl-based smoke testing.
	reflection.Register(srv)

	log.Printf("ShippingService listening on port %s", port)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// server implements the hipstershop.ShippingService gRPC interface.
type server struct {
	pb.UnimplementedShippingServiceServer
}

// GetQuote produces a shipping quote (cost) in USD.
func (s *server) GetQuote(ctx context.Context, in *pb.GetQuoteRequest) (*pb.GetQuoteResponse, error) {
	log.Printf("[GetQuote] received request with %d items", len(in.GetItems()))

	// 1. Generate a quote based on the total number of items to be shipped.
	count := 0
	for _, item := range in.GetItems() {
		count += int(item.GetQuantity())
	}
	quote := CreateQuoteFromCount(count)

	// 2. Generate a response.
	return &pb.GetQuoteResponse{
		CostUsd: quote.ToMoney(),
	}, nil
}

// ShipOrder mocks that the requested items will be shipped.
// It supplies a tracking ID for notional lookup of shipment delivery status.
func (s *server) ShipOrder(ctx context.Context, in *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	log.Printf("[ShipOrder] received request for %d items", len(in.GetItems()))

	// 1. Create a tracking ID seeded by the destination address.
	baseAddress := fmt.Sprintf("%s, %s, %s",
		in.GetAddress().GetStreetAddress(),
		in.GetAddress().GetCity(),
		in.GetAddress().GetState(),
	)

	// 2. Generate a response.
	return &pb.ShipOrderResponse{
		TrackingId: CreateTrackingId(baseAddress),
	}, nil
}
