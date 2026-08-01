// Package validator implements lightweight request validation for the
// frontend forms.
//
// It mirrors the upstream Online Boutique frontend (which uses
// go-playground/validator) with the same payload shapes, the same rules
// (quantity 1..10, required fields, email, credit-card checksum, ISO 4217
// currency codes) and the same "Field 'X' is invalid: <tag>" error wording,
// but with zero external dependencies.
//
// The one intentional deviation: the credit-card check is a plain Luhn
// checksum instead of the upstream validator's card-prefix regex, so any
// syntactically valid test card (e.g. the 4432801561520454 demo card) is
// accepted.
package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	iso4217Re  = regexp.MustCompile(`^[A-Z]{3}$`)
	digitsOnly = strings.NewReplacer("-", "", " ", "")
)

// FieldError describes a single failed validation.
type FieldError struct {
	Field string
	Tag   string
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("Field '%s' is invalid: %s", e.Field, e.Tag)
}

// ValidationErrorResponse formats err as a user-facing validation error
// message (same wording as the upstream validator response).
func ValidationErrorResponse(err error) error {
	var fe *FieldError
	if errors.As(err, &fe) {
		return errors.New(fe.Error())
	}
	return errors.New("invalid validation error format")
}

// AddToCartPayload mirrors the upstream AddToCartPayload.
type AddToCartPayload struct {
	Quantity  uint64
	ProductID string
}

// Validate checks the add-to-cart form fields. Equivalent to upstream tags
// `required,gte=1,lte=10` (Quantity) and `required` (ProductID).
func (ad *AddToCartPayload) Validate() error {
	if ad.ProductID == "" {
		return &FieldError{Field: "ProductID", Tag: "required"}
	}
	if ad.Quantity < 1 || ad.Quantity > 10 {
		return &FieldError{Field: "Quantity", Tag: "gte=1,lte=10"}
	}
	return nil
}

// PlaceOrderPayload mirrors the upstream PlaceOrderPayload.
type PlaceOrderPayload struct {
	Email         string
	StreetAddress string
	ZipCode       int64
	City          string
	State         string
	Country       string
	CcNumber      string
	CcMonth       int64
	CcYear        int64
	CcCVV         int64
}

// Validate checks every checkout form field with the same rules as the
// upstream go-playground validator tags.
func (po *PlaceOrderPayload) Validate() error {
	if po.Email == "" {
		return &FieldError{Field: "Email", Tag: "required"}
	}
	if !emailRe.MatchString(po.Email) {
		return &FieldError{Field: "Email", Tag: "email"}
	}
	if po.StreetAddress == "" {
		return &FieldError{Field: "StreetAddress", Tag: "required"}
	}
	if len(po.StreetAddress) > 512 {
		return &FieldError{Field: "StreetAddress", Tag: "max=512"}
	}
	if po.ZipCode == 0 {
		return &FieldError{Field: "ZipCode", Tag: "required"}
	}
	if po.City == "" {
		return &FieldError{Field: "City", Tag: "required"}
	}
	if len(po.City) > 128 {
		return &FieldError{Field: "City", Tag: "max=128"}
	}
	if po.State == "" {
		return &FieldError{Field: "State", Tag: "required"}
	}
	if len(po.State) > 128 {
		return &FieldError{Field: "State", Tag: "max=128"}
	}
	if po.Country == "" {
		return &FieldError{Field: "Country", Tag: "required"}
	}
	if len(po.Country) > 128 {
		return &FieldError{Field: "Country", Tag: "max=128"}
	}
	if po.CcNumber == "" {
		return &FieldError{Field: "CcNumber", Tag: "required"}
	}
	if !luhnValid(digitsOnly.Replace(po.CcNumber)) {
		return &FieldError{Field: "CcNumber", Tag: "credit_card"}
	}
	if po.CcMonth < 1 || po.CcMonth > 12 {
		return &FieldError{Field: "CcMonth", Tag: "gte=1,lte=12"}
	}
	if po.CcYear == 0 {
		return &FieldError{Field: "CcYear", Tag: "required"}
	}
	if po.CcCVV == 0 {
		return &FieldError{Field: "CcCVV", Tag: "required"}
	}
	return nil
}

// SetCurrencyPayload mirrors the upstream SetCurrencyPayload.
type SetCurrencyPayload struct {
	Currency string
}

// Validate checks the currency code is a non-empty 3-letter ISO 4217 code
// (equivalent to upstream tags `required,iso4217`).
func (sc *SetCurrencyPayload) Validate() error {
	if sc.Currency == "" {
		return &FieldError{Field: "Currency", Tag: "required"}
	}
	if !iso4217Re.MatchString(sc.Currency) {
		return &FieldError{Field: "Currency", Tag: "iso4217"}
	}
	return nil
}

// luhnValid reports whether the digits pass the Luhn checksum (the same
// check behind upstream's `credit_card` validator tag).
func luhnValid(number string) bool {
	var sum int
	parity := len(number) % 2
	for i, r := range number {
		if r < '0' || r > '9' {
			return false
		}
		d := int(r - '0')
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return len(number) > 0 && sum%10 == 0
}
