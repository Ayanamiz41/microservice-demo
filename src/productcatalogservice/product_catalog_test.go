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
	"context"
	"os"
	"testing"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	mockProductCatalog *productCatalog
)

func TestMain(m *testing.M) {
	mockProductCatalog = &productCatalog{
		catalog: pb.ListProductsResponse{
			Products: []*pb.Product{},
		},
	}

	mockProductCatalog.catalog.Products = append(mockProductCatalog.catalog.Products, &pb.Product{
		Id:   "abc001",
		Name: "Product Alpha One",
	})
	mockProductCatalog.catalog.Products = append(mockProductCatalog.catalog.Products, &pb.Product{
		Id:   "abc002",
		Name: "Product Delta",
	})
	mockProductCatalog.catalog.Products = append(mockProductCatalog.catalog.Products, &pb.Product{
		Id:   "abc003",
		Name: "Product Alpha Two",
	})
	mockProductCatalog.catalog.Products = append(mockProductCatalog.catalog.Products, &pb.Product{
		Id:   "abc004",
		Name: "Product Gamma",
	})

	os.Exit(m.Run())
}

func TestGetProductExists(t *testing.T) {
	product, err := mockProductCatalog.GetProduct(context.Background(),
		&pb.GetProductRequest{Id: "abc003"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := product.Name, "Product Alpha Two"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestGetProductNotFound(t *testing.T) {
	_, err := mockProductCatalog.GetProduct(context.Background(),
		&pb.GetProductRequest{Id: "abc005"},
	)
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestListProducts(t *testing.T) {
	products, err := mockProductCatalog.ListProducts(context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(products.Products), 4; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestSearchProducts(t *testing.T) {
	products, err := mockProductCatalog.SearchProducts(context.Background(),
		&pb.SearchProductsRequest{Query: "alpha"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(products.Results), 2; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestLoadCatalog(t *testing.T) {
	var catalog pb.ListProductsResponse
	if err := loadCatalog(&catalog); err != nil {
		t.Fatalf("loadCatalog() error = %v", err)
	}
	if got, want := len(catalog.Products), 9; got != want {
		t.Errorf("got %d products, want %d", got, want)
	}

	// Spot check a known product: Sunglasses.
	var sunglasses *pb.Product
	for _, p := range catalog.Products {
		if p.Id == "OLJCESPC7Z" {
			sunglasses = p
			break
		}
	}
	if sunglasses == nil {
		t.Fatal("expected product OLJCESPC7Z in catalog")
	}
	if got, want := sunglasses.Name, "Sunglasses"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := sunglasses.PriceUsd.GetUnits(), int64(19); got != want {
		t.Errorf("got units %d, want %d", got, want)
	}
	if got, want := sunglasses.PriceUsd.GetNanos(), int32(990000000); got != want {
		t.Errorf("got nanos %d, want %d", got, want)
	}
}
