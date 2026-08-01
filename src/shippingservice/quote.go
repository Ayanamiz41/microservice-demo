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
	"fmt"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// Quote represents a shipping quote in USD.
//
// All money arithmetic is done in integer minimal units (dollars and cents);
// no floating point is involved in any price calculation.
type Quote struct {
	Dollars uint32
	Cents   uint32
}

// String returns the human-readable representation of the Quote.
func (q Quote) String() string {
	return fmt.Sprintf("$%d.%02d", q.Dollars, q.Cents)
}

// CreateQuoteFromCount takes the total number of items to ship and returns a
// shipping quote: free for an empty cart, otherwise a flat fee of USD 8.99
// (matching the upstream Online Boutique shippingservice behavior).
func CreateQuoteFromCount(count int) Quote {
	if count <= 0 {
		return Quote{}
	}
	return Quote{Dollars: 8, Cents: 99}
}

// ToMoney converts the Quote to the hipstershop.Money wire type.
// One cent equals 10^7 nanos, so the conversion stays purely integral.
func (q Quote) ToMoney() *pb.Money {
	return &pb.Money{
		CurrencyCode: "USD",
		Units:        int64(q.Dollars),
		Nanos:        int32(q.Cents) * 10000000,
	}
}
