package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/Ayanamiz41/microservice-demo/genproto/go/hipstershop"
	"github.com/Ayanamiz41/microservice-demo/src/frontend/money"
	"github.com/Ayanamiz41/microservice-demo/src/frontend/validator"
)

type platformDetails struct {
	css      string
	provider string
}

var (
	frontendMessage  = strings.TrimSpace(os.Getenv("FRONTEND_MESSAGE"))
	isCymbalBrand    = strings.ToLower(os.Getenv("CYMBAL_BRANDING")) == "true"
	assistantEnabled = strings.ToLower(os.Getenv("ENABLE_ASSISTANT")) == "true"
	baseURL          = os.Getenv("BASE_URL")
	templates        = template.Must(template.New("").
				Funcs(template.FuncMap{
			"renderMoney":        renderMoney,
			"renderCurrencyLogo": renderCurrencyLogo,
		}).ParseGlob("templates/*.html"))
	plat platformDetails
)

var validEnvs = []string{"local", "gcp", "azure", "aws", "onprem", "alibaba"}

// routes registers every HTTP route on mux (paths are the same as upstream,
// with `{id}`/`{ids}` wildcards handled by the stdlib ServeMux).
func (fe *frontendServer) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", fe.homeHandler)
	mux.HandleFunc("GET /product/{id}", fe.productHandler)
	mux.HandleFunc("GET /cart", fe.viewCartHandler)
	mux.HandleFunc("POST /cart", fe.addToCartHandler)
	mux.HandleFunc("POST /cart/empty", fe.emptyCartHandler)
	mux.HandleFunc("POST /setCurrency", fe.setCurrencyHandler)
	mux.HandleFunc("GET /logout", fe.logoutHandler)
	mux.HandleFunc("POST /cart/checkout", fe.placeOrderHandler)
	mux.HandleFunc("GET /assistant", fe.assistantHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "User-agent: *\nDisallow: /")
	})
	mux.HandleFunc("GET /_healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("GET /product-meta/{ids}", fe.getProductByID)
	mux.HandleFunc("POST /bot", fe.chatBotHandler)
}

func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("home page: currency=%s", currentCurrency(r))
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve currencies: %w", err), http.StatusInternalServerError)
		return
	}
	products, err := fe.getProducts(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve products: %w", err), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve cart: %w", err), http.StatusInternalServerError)
		return
	}

	type productView struct {
		Item  *pb.Product
		Price *pb.Money
	}
	ps := make([]productView, len(products))
	for i, p := range products {
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(r, w, fmt.Errorf("failed to do currency conversion for product %s: %w", p.GetId(), err), http.StatusInternalServerError)
			return
		}
		ps[i] = productView{p, price}
	}

	// ENV_PLATFORM selects the platform flag shown in the UI (default local;
	// upstream auto-detects GCP, which does not apply to this local replica).
	env := strings.ToLower(os.Getenv("ENV_PLATFORM"))
	if !stringInSlice(validEnvs, env) {
		env = "local"
	}
	plat.setPlatformDetails(env)

	if err := templates.ExecuteTemplate(w, "home", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency": true,
		"currencies":    currencies,
		"products":      ps,
		"cart_size":     cartSize(cart),
		"banner_color":  os.Getenv("BANNER_COLOR"), // illustrates canary deployments
		"ad":            fe.chooseAd(r.Context(), []string{}),
	})); err != nil {
		log.Printf("failed to render home template: %v", err)
	}
}

func (plat *platformDetails) setPlatformDetails(env string) {
	switch env {
	case "aws":
		plat.provider, plat.css = "AWS", "aws-platform"
	case "onprem":
		plat.provider, plat.css = "On-Premises", "onprem-platform"
	case "azure":
		plat.provider, plat.css = "Azure", "azure-platform"
	case "gcp":
		plat.provider, plat.css = "Google Cloud", "gcp-platform"
	case "alibaba":
		plat.provider, plat.css = "Alibaba Cloud", "alibaba-platform"
	default:
		plat.provider, plat.css = "local", "local"
	}
}

func (fe *frontendServer) productHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		renderHTTPError(r, w, fmt.Errorf("product id not specified"), http.StatusBadRequest)
		return
	}
	log.Printf("product page: id=%s currency=%s", id, currentCurrency(r))

	p, err := fe.getProduct(r.Context(), id)
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve product: %w", err), http.StatusInternalServerError)
		return
	}
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve currencies: %w", err), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve cart: %w", err), http.StatusInternalServerError)
		return
	}
	price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to convert currency: %w", err), http.StatusInternalServerError)
		return
	}

	// ignores the error retrieving recommendations since it is not critical
	recommendations, err := fe.getRecommendations(r.Context(), sessionID(r), []string{id})
	if err != nil {
		log.Printf("failed to get product recommendations: %v", err)
	}

	product := struct {
		Item  *pb.Product
		Price *pb.Money
	}{p, price}

	if err := templates.ExecuteTemplate(w, "product", injectCommonTemplateData(r, map[string]interface{}{
		"ad":              fe.chooseAd(r.Context(), p.GetCategories()),
		"show_currency":   true,
		"currencies":      currencies,
		"product":         product,
		"recommendations": recommendations,
		"cart_size":       cartSize(cart),
	})); err != nil {
		log.Printf("failed to render product template: %v", err)
	}
}

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
	productID := r.FormValue("product_id")
	payload := validator.AddToCartPayload{Quantity: quantity, ProductID: productID}
	if err := payload.Validate(); err != nil {
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Printf("adding to cart: product=%s quantity=%d", payload.ProductID, payload.Quantity)

	p, err := fe.getProduct(r.Context(), payload.ProductID)
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve product: %w", err), http.StatusInternalServerError)
		return
	}

	if err := fe.insertCart(r.Context(), sessionID(r), p.GetId(), int32(payload.Quantity)); err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to add to cart: %w", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", baseURL+"/cart")
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) emptyCartHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("emptying cart")

	if err := fe.emptyCart(r.Context(), sessionID(r)); err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to empty cart: %w", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", baseURL+"/")
	w.WriteHeader(http.StatusFound)
}

func (fe *frontendServer) viewCartHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("view cart")
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve currencies: %w", err), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(r.Context(), sessionID(r))
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve cart: %w", err), http.StatusInternalServerError)
		return
	}

	// ignores the error retrieving recommendations since it is not critical
	recommendations, err := fe.getRecommendations(r.Context(), sessionID(r), cartIDs(cart))
	if err != nil {
		log.Printf("failed to get cart recommendations: %v", err)
	}

	shippingCost, err := fe.getShippingQuote(r.Context(), cart, currentCurrency(r))
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to get shipping quote: %w", err), http.StatusInternalServerError)
		return
	}

	type cartItemView struct {
		Item     *pb.Product
		Quantity int32
		Price    *pb.Money
	}
	items := make([]cartItemView, len(cart))
	totalPrice := pb.Money{CurrencyCode: currentCurrency(r)}
	for i, item := range cart {
		p, err := fe.getProduct(r.Context(), item.GetProductId())
		if err != nil {
			renderHTTPError(r, w, fmt.Errorf("could not retrieve product #%s: %w", item.GetProductId(), err), http.StatusInternalServerError)
			return
		}
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(r, w, fmt.Errorf("could not convert currency for product #%s: %w", item.GetProductId(), err), http.StatusInternalServerError)
			return
		}

		multPrice := money.MultiplySlow(*price, uint32(item.GetQuantity()))
		items[i] = cartItemView{
			Item:     p,
			Quantity: item.GetQuantity(),
			Price:    &multPrice,
		}
		totalPrice = money.Must(money.Sum(totalPrice, multPrice))
	}
	totalPrice = money.Must(money.Sum(totalPrice, *shippingCost))
	year := time.Now().Year()

	if err := templates.ExecuteTemplate(w, "cart", injectCommonTemplateData(r, map[string]interface{}{
		"currencies":       currencies,
		"recommendations":  recommendations,
		"cart_size":        cartSize(cart),
		"shipping_cost":    shippingCost,
		"show_currency":    true,
		"total_cost":       totalPrice,
		"items":            items,
		"expiration_years": []int{year, year + 1, year + 2, year + 3, year + 4},
	})); err != nil {
		log.Printf("failed to render cart template: %v", err)
	}
}

func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("placing order")

	var (
		email         = r.FormValue("email")
		streetAddress = r.FormValue("street_address")
		zipCode, _    = strconv.ParseInt(r.FormValue("zip_code"), 10, 32)
		city          = r.FormValue("city")
		state         = r.FormValue("state")
		country       = r.FormValue("country")
		ccNumber      = r.FormValue("credit_card_number")
		ccMonth, _    = strconv.ParseInt(r.FormValue("credit_card_expiration_month"), 10, 32)
		ccYear, _     = strconv.ParseInt(r.FormValue("credit_card_expiration_year"), 10, 32)
		ccCVV, _      = strconv.ParseInt(r.FormValue("credit_card_cvv"), 10, 32)
	)

	payload := validator.PlaceOrderPayload{
		Email:         email,
		StreetAddress: streetAddress,
		ZipCode:       zipCode,
		City:          city,
		State:         state,
		Country:       country,
		CcNumber:      ccNumber,
		CcMonth:       ccMonth,
		CcYear:        ccYear,
		CcCVV:         ccCVV,
	}
	if err := payload.Validate(); err != nil {
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}

	order, err := pb.NewCheckoutServiceClient(fe.checkoutSvcConn).
		PlaceOrder(r.Context(), &pb.PlaceOrderRequest{
			Email: payload.Email,
			CreditCard: &pb.CreditCardInfo{
				CreditCardNumber:          payload.CcNumber,
				CreditCardExpirationMonth: int32(payload.CcMonth),
				CreditCardExpirationYear:  int32(payload.CcYear),
				CreditCardCvv:             int32(payload.CcCVV),
			},
			UserId:       sessionID(r),
			UserCurrency: currentCurrency(r),
			Address: &pb.Address{
				StreetAddress: payload.StreetAddress,
				City:          payload.City,
				State:         payload.State,
				ZipCode:       int32(payload.ZipCode),
				Country:       payload.Country,
			},
		})
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to complete the order: %w", err), http.StatusInternalServerError)
		return
	}
	log.Printf("order placed: order_id=%s", order.GetOrder().GetOrderId())

	recommendations, _ := fe.getRecommendations(r.Context(), sessionID(r), nil)

	totalPaid := *order.GetOrder().GetShippingCost()
	for _, v := range order.GetOrder().GetItems() {
		multPrice := money.MultiplySlow(*v.GetCost(), uint32(v.GetItem().GetQuantity()))
		totalPaid = money.Must(money.Sum(totalPaid, multPrice))
	}

	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve currencies: %w", err), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "order", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency":   false,
		"currencies":      currencies,
		"order":           order.GetOrder(),
		"total_paid":      &totalPaid,
		"recommendations": recommendations,
	})); err != nil {
		log.Printf("failed to render order template: %v", err)
	}
}

func (fe *frontendServer) assistantHandler(w http.ResponseWriter, r *http.Request) {
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve currencies: %w", err), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "assistant", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency": false,
		"currencies":    currencies,
	})); err != nil {
		log.Printf("failed to render assistant template: %v", err)
	}
}

func (fe *frontendServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("logging out")
	for _, c := range r.Cookies() {
		c.Expires = time.Now().Add(-time.Hour * 24 * 365)
		c.MaxAge = -1
		http.SetCookie(w, c)
	}
	w.Header().Set("Location", baseURL+"/")
	w.WriteHeader(http.StatusFound)
}

// getProductByID serves a single product as JSON; used by the assistant UI
// to fetch product metadata for recommendation cards.
func (fe *frontendServer) getProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("ids")
	if id == "" {
		return
	}
	p, err := fe.getProduct(r.Context(), id)
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("could not retrieve product: %w", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("failed to encode product %s as json: %v", id, err)
	}
}

// chatBotHandler forwards the assistant chat message to the
// ShoppingAssistantService (gRPC GetCompletion) and returns the reply as
// JSON in the shape the assistant UI expects.
func (fe *frontendServer) chatBotHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Message string `json:"message"`
		Image   string `json:"image"`
	}
	type response struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversation_id,omitempty"`
	}

	var req request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to read request body: %w", err), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to parse request body: %w", err), http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		renderHTTPError(r, w, fmt.Errorf("empty assistant message"), http.StatusBadRequest)
		return
	}

	reply, err := fe.getAssistantCompletion(r.Context(), sessionID(r), req.Message)
	if err != nil {
		renderHTTPError(r, w, fmt.Errorf("failed to get assistant completion: %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response{
		Message:        reply.GetMessage(),
		ConversationID: reply.GetConversationId(),
	}); err != nil {
		log.Printf("failed to encode assistant response: %v", err)
	}
}

func (fe *frontendServer) setCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	cur := r.FormValue("currency_code")
	payload := validator.SetCurrencyPayload{Currency: cur}
	if err := payload.Validate(); err != nil {
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Printf("setting currency: old=%s new=%s", currentCurrency(r), payload.Currency)

	if payload.Currency != "" {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieCurrency,
			Value:  payload.Currency,
			MaxAge: cookieMaxAge,
		})
	}
	referer := r.Header.Get("referer")
	if referer == "" {
		referer = baseURL + "/"
	}
	w.Header().Set("Location", referer)
	w.WriteHeader(http.StatusFound)
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, ctxKeys []string) *pb.Ad {
	ads, err := fe.getAd(ctx, ctxKeys)
	if err != nil {
		log.Printf("failed to retrieve ads: %v", err)
		return nil
	}
	if len(ads) == 0 {
		return nil
	}
	return ads[rand.Intn(len(ads))]
}

func renderHTTPError(r *http.Request, w http.ResponseWriter, err error, code int) {
	log.Printf("request error: status=%d err=%v", code, err)
	errMsg := fmt.Sprintf("%+v", err)

	w.WriteHeader(code)

	if templateErr := templates.ExecuteTemplate(w, "error", injectCommonTemplateData(r, map[string]interface{}{
		"error":       errMsg,
		"status_code": code,
		"status":      http.StatusText(code),
	})); templateErr != nil {
		log.Printf("failed to render error template: %v", templateErr)
	}
}

func injectCommonTemplateData(r *http.Request, payload map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"assistant_enabled": assistantEnabled,
		"frontendMessage":   frontendMessage,
		"currentYear":       time.Now().Year(),
		"baseUrl":           baseURL,
	}

	for k, v := range payload {
		data[k] = v
	}

	return data
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func cartIDs(c []*pb.CartItem) []string {
	out := make([]string, len(c))
	for i, v := range c {
		out[i] = v.GetProductId()
	}
	return out
}

// cartSize returns the total # of items in the cart.
func cartSize(c []*pb.CartItem) int {
	cartSize := 0
	for _, item := range c {
		cartSize += int(item.GetQuantity())
	}
	return cartSize
}

func renderMoney(money pb.Money) string {
	currencyLogo := renderCurrencyLogo(money.GetCurrencyCode())
	return fmt.Sprintf("%s%d.%02d", currencyLogo, money.GetUnits(), money.GetNanos()/10000000)
}

func renderCurrencyLogo(currencyCode string) string {
	logos := map[string]string{
		"USD": "$",
		"CAD": "$",
		"JPY": "¥",
		"EUR": "€",
		"TRY": "₺",
		"GBP": "£",
	}

	logo := "$" //default
	if val, ok := logos[currencyCode]; ok {
		logo = val
	}
	return logo
}

func stringInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
