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
	_ "embed"
	"log"

	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// productsJSON is the built-in product catalog, embedded at build time.
// The data mirrors the upstream Online Boutique product catalog.
//
//go:embed products.json
var productsJSON []byte

// loadCatalog parses the embedded products.json into catalog.
func loadCatalog(catalog *pb.ListProductsResponse) error {
	log.Println("loading catalog from embedded products.json...")

	if err := protojson.Unmarshal(productsJSON, catalog); err != nil {
		log.Printf("failed to parse the product catalog json: %v", err)
		return err
	}

	log.Printf("successfully parsed product catalog json (%d products)", len(catalog.Products))
	return nil
}
