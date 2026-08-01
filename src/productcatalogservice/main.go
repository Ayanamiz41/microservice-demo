// Copyright 2026 Ayanamiz41
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
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	logger *log.Logger

	port = "3550"
)

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

func main() {
	flag.Parse()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	logger.Printf("starting productcatalogservice grpc server at :%s", port)

	run(port)

	// Wait for a termination signal before exiting.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	logger.Println("shutting down productcatalogservice")
}

// run starts the gRPC server on the given port and blocks until the server
// exits. Returns the address the listener was bound to.
func run(port string) string {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatalf("failed to listen on port %s: %v", port, err)
	}

	srv := grpc.NewServer()

	svc := &productCatalog{}
	pb.RegisterProductCatalogServiceServer(srv, svc)

	// Register the standard gRPC health service and mark both the overall
	// server and the product catalog service as SERVING.
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthcheck)
	healthcheck.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthcheck.SetServingStatus("hipstershop.ProductCatalogService", healthpb.HealthCheckResponse_SERVING)

	// Reflection makes the server introspectable with grpcurl.
	reflection.Register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			logger.Fatalf("failed to serve: %v", err)
		}
	}()

	return lis.Addr().String()
}
