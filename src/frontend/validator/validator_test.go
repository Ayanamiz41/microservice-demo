package validator

import (
	"errors"
	"strings"
	"testing"
)

func TestAddToCartValidate(t *testing.T) {
	tests := []struct {
		name    string
		payload AddToCartPayload
		wantErr string
	}{
		{"valid", AddToCartPayload{Quantity: 1, ProductID: "OLJCESPC7Z"}, ""},
		{"valid max quantity", AddToCartPayload{Quantity: 10, ProductID: "OLJCESPC7Z"}, ""},
		{"missing product id", AddToCartPayload{Quantity: 1}, "Field 'ProductID' is invalid: required"},
		{"zero quantity", AddToCartPayload{Quantity: 0, ProductID: "OLJCESPC7Z"}, "Field 'Quantity' is invalid: gte=1,lte=10"},
		{"quantity over 10", AddToCartPayload{Quantity: 11, ProductID: "OLJCESPC7Z"}, "Field 'Quantity' is invalid: gte=1,lte=10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestPlaceOrderValidate(t *testing.T) {
	valid := PlaceOrderPayload{
		Email:         "someone@example.com",
		StreetAddress: "1600 Amphitheatre Parkway",
		ZipCode:       94043,
		City:          "Mountain View",
		State:         "CA",
		Country:       "United States",
		CcNumber:      "4432801561520454",
		CcMonth:       1,
		CcYear:        2030,
		CcCVV:         672,
	}

	tests := []struct {
		name    string
		mutate  func(*PlaceOrderPayload)
		wantErr string
	}{
		{"valid", func(p *PlaceOrderPayload) {}, ""},
		{"valid cc with dashes", func(p *PlaceOrderPayload) { p.CcNumber = "4432-8015-6152-0454" }, ""},
		{"missing email", func(p *PlaceOrderPayload) { p.Email = "" }, "Field 'Email' is invalid: required"},
		{"bad email", func(p *PlaceOrderPayload) { p.Email = "not-an-email" }, "Field 'Email' is invalid: email"},
		{"missing street", func(p *PlaceOrderPayload) { p.StreetAddress = "" }, "Field 'StreetAddress' is invalid: required"},
		{"missing zip", func(p *PlaceOrderPayload) { p.ZipCode = 0 }, "Field 'ZipCode' is invalid: required"},
		{"missing city", func(p *PlaceOrderPayload) { p.City = "" }, "Field 'City' is invalid: required"},
		{"missing state", func(p *PlaceOrderPayload) { p.State = "" }, "Field 'State' is invalid: required"},
		{"missing country", func(p *PlaceOrderPayload) { p.Country = "" }, "Field 'Country' is invalid: required"},
		{"missing cc", func(p *PlaceOrderPayload) { p.CcNumber = "" }, "Field 'CcNumber' is invalid: required"},
		{"invalid cc", func(p *PlaceOrderPayload) { p.CcNumber = "1234567890123456" }, "Field 'CcNumber' is invalid: credit_card"},
		{"month zero", func(p *PlaceOrderPayload) { p.CcMonth = 0 }, "Field 'CcMonth' is invalid: gte=1,lte=12"},
		{"month 13", func(p *PlaceOrderPayload) { p.CcMonth = 13 }, "Field 'CcMonth' is invalid: gte=1,lte=12"},
		{"missing year", func(p *PlaceOrderPayload) { p.CcYear = 0 }, "Field 'CcYear' is invalid: required"},
		{"missing cvv", func(p *PlaceOrderPayload) { p.CcCVV = 0 }, "Field 'CcCVV' is invalid: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.mutate(&p)
			err := p.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestSetCurrencyValidate(t *testing.T) {
	tests := []struct {
		name    string
		payload SetCurrencyPayload
		wantErr string
	}{
		{"valid", SetCurrencyPayload{Currency: "USD"}, ""},
		{"valid eur", SetCurrencyPayload{Currency: "EUR"}, ""},
		{"empty", SetCurrencyPayload{}, "Field 'Currency' is invalid: required"},
		{"lowercase", SetCurrencyPayload{Currency: "usd"}, "Field 'Currency' is invalid: iso4217"},
		{"not a code", SetCurrencyPayload{Currency: "US"}, "Field 'Currency' is invalid: iso4217"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidationErrorResponse(t *testing.T) {
	err := ValidationErrorResponse(&FieldError{Field: "Email", Tag: "required"})
	if err == nil || !strings.Contains(err.Error(), "Field 'Email' is invalid: required") {
		t.Errorf("ValidationErrorResponse = %v, want formatted FieldError message", err)
	}
	err = ValidationErrorResponse(errors.New("other"))
	if err == nil || err.Error() != "invalid validation error format" {
		t.Errorf("ValidationErrorResponse = %v, want 'invalid validation error format'", err)
	}
}

func TestLuhnValid(t *testing.T) {
	if !luhnValid("4432801561520454") {
		t.Error("luhnValid(4432801561520454) = false, want true (demo test card)")
	}
	if luhnValid("1234567890123456") {
		t.Error("luhnValid(1234567890123456) = true, want false")
	}
	if luhnValid("") {
		t.Error("luhnValid(\"\") = true, want false")
	}
}
