package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
)

// avoidNoopCurrencyConversionRPC skips the currency RPC when the amount is
// already expressed in the target currency. Upstream keeps this false
// (always ask the currencyservice); this replica turns it on so the default
// USD flow keeps working even while currencyservice is not deployed, while
// real conversions (e.g. USD -> EUR when the user switches currency) still
// go through the service.
const avoidNoopCurrencyConversionRPC = true

// getCurrencies returns the whitelisted currencies offered by the currency
// service. When currencyservice is unreachable it falls back to the
// whitelisted defaults (USD is the default currency and needs no rates), so
// the storefront keeps rendering locally.
func (fe *frontendServer) getCurrencies(ctx context.Context) ([]string, error) {
	currs, err := pb.NewCurrencyServiceClient(fe.currencySvcConn).
		GetSupportedCurrencies(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("warning: failed to retrieve currencies from currencyservice: %v", err)
		out := make([]string, 0, len(whitelistedCurrencies))
		for c := range whitelistedCurrencies {
			out = append(out, c)
		}
		sort.Strings(out)
		return out, nil
	}
	var out []string
	for _, c := range currs.GetCurrencyCodes() {
		if whitelistedCurrencies[c] {
			out = append(out, c)
		}
	}
	return out, nil
}

func (fe *frontendServer) getProducts(ctx context.Context) ([]*pb.Product, error) {
	resp, err := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn).
		ListProducts(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return resp.GetProducts(), nil
}

func (fe *frontendServer) getProduct(ctx context.Context, id string) (*pb.Product, error) {
	return pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn).
		GetProduct(ctx, &pb.GetProductRequest{Id: id})
}

func (fe *frontendServer) getCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	resp, err := pb.NewCartServiceClient(fe.cartSvcConn).GetCart(ctx, &pb.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return resp.GetItems(), nil
}

func (fe *frontendServer) emptyCart(ctx context.Context, userID string) error {
	_, err := pb.NewCartServiceClient(fe.cartSvcConn).EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userID})
	return err
}

func (fe *frontendServer) insertCart(ctx context.Context, userID, productID string, quantity int32) error {
	_, err := pb.NewCartServiceClient(fe.cartSvcConn).AddItem(ctx, &pb.AddItemRequest{
		UserId: userID,
		Item:   &pb.CartItem{ProductId: productID, Quantity: quantity},
	})
	return err
}

func (fe *frontendServer) convertCurrency(ctx context.Context, m *pb.Money, currency string) (*pb.Money, error) {
	if avoidNoopCurrencyConversionRPC && m.GetCurrencyCode() == currency {
		return m, nil
	}
	return pb.NewCurrencyServiceClient(fe.currencySvcConn).Convert(ctx, &pb.CurrencyConversionRequest{
		From:   m,
		ToCode: currency,
	})
}

func (fe *frontendServer) getShippingQuote(ctx context.Context, items []*pb.CartItem, currency string) (*pb.Money, error) {
	quote, err := pb.NewShippingServiceClient(fe.shippingSvcConn).GetQuote(ctx, &pb.GetQuoteRequest{
		Address: nil,
		Items:   items,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping quote: %w", err)
	}
	localized, err := fe.convertCurrency(ctx, quote.GetCostUsd(), currency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency for shipping cost: %w", err)
	}
	return localized, nil
}

func (fe *frontendServer) getRecommendations(ctx context.Context, userID string, productIDs []string) ([]*pb.Product, error) {
	resp, err := pb.NewRecommendationServiceClient(fe.recommendationSvcConn).
		ListRecommendations(ctx, &pb.ListRecommendationsRequest{UserId: userID, ProductIds: productIDs})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Product, 0, len(resp.GetProductIds()))
	for _, v := range resp.GetProductIds() {
		p, err := fe.getProduct(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("failed to get recommended product info (#%s): %w", v, err)
		}
		out = append(out, p)
	}
	if len(out) > 4 {
		out = out[:4] // take only first four to fit the UI
	}
	return out, nil
}

func (fe *frontendServer) getAd(ctx context.Context, ctxKeys []string) ([]*pb.Ad, error) {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	resp, err := pb.NewAdServiceClient(fe.adSvcConn).GetAds(ctx, &pb.AdRequest{ContextKeys: ctxKeys})
	if err != nil {
		return nil, fmt.Errorf("failed to get ads: %w", err)
	}
	return resp.GetAds(), nil
}

func (fe *frontendServer) getAssistantCompletion(ctx context.Context, userID, message string) (*pb.ShoppingAssistantResponse, error) {
	return pb.NewShoppingAssistantServiceClient(fe.shoppingAssistantSvcConn).
		GetCompletion(ctx, &pb.ShoppingAssistantRequest{UserId: userID, Message: message})
}
