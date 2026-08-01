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

// Command productcatalogservice is the Go implementation of the Online
// Boutique product catalog service. It serves the hipstershop.ProductCatalogService
// gRPC contract (ListProducts / GetProduct / SearchProducts) backed by a
// built-in catalog of products, plus the standard grpc.health.v1.Health
// service.
package main

import (
	"context"
	"strings"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// productCatalog implements hipstershop.ProductCatalogService.
//
// The catalog is loaded once at startup (and lazily on first request) from
// products.json, which is embedded into the binary via go:embed.
type productCatalog struct {
	pb.UnimplementedProductCatalogServiceServer
	catalog pb.ListProductsResponse
}

func (p *productCatalog) ListProducts(context.Context, *pb.Empty) (*pb.ListProductsResponse, error) {
	return &pb.ListProductsResponse{Products: p.parseCatalog()}, nil
}

func (p *productCatalog) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	catalog := p.parseCatalog()
	for _, product := range catalog {
		if req.Id == product.Id {
			return product, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
}

func (p *productCatalog) SearchProducts(_ context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	var ps []*pb.Product
	query := strings.ToLower(req.Query)
	for _, product := range p.parseCatalog() {
		if strings.Contains(strings.ToLower(product.Name), query) ||
			strings.Contains(strings.ToLower(product.Description), query) {
			ps = append(ps, product)
		}
	}

	return &pb.SearchProductsResponse{Results: ps}, nil
}

func (p *productCatalog) parseCatalog() []*pb.Product {
	if len(p.catalog.Products) == 0 {
		if err := loadCatalog(&p.catalog); err != nil {
			return []*pb.Product{}
		}
	}

	return p.catalog.Products
}
