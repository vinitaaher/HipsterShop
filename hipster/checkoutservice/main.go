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
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"

	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/trace"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/checkoutservice/genproto"
	money "github.com/GoogleCloudPlatform/microservices-demo/src/checkoutservice/money"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	usdCurrency = "USD"
)

var log *logrus.Logger
var serviceName string
var serviceNameSpace string

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
}

type checkoutService struct {
	productCatalogSvcAddr string
	cartSvcAddr           string
	currencySvcAddr       string
	shippingSvcAddr       string
	emailSvcAddr          string
	paymentSvcAddr        string
}

func checkoutserviceConstructor(productCatalogSvcAddr string, currencySvcAddr string, cartSvcAddr string, shippingSvcAddr string, paymentSvcAddr string, emailSvcAddr string) *checkoutService {
	obj := new(checkoutService)

	obj.productCatalogSvcAddr = productCatalogSvcAddr
	obj.currencySvcAddr = currencySvcAddr
	obj.cartSvcAddr = cartSvcAddr
	obj.shippingSvcAddr = shippingSvcAddr
	obj.paymentSvcAddr = paymentSvcAddr
	obj.emailSvcAddr = emailSvcAddr

	return obj
}
func detectResource() (*resource.Resource, error) {
	var instID label.KeyValue
	if host, ok := os.LookupEnv("HOSTNAME"); ok && host != "" {
		instID = semconv.ServiceInstanceIDKey.String(host)
	} else {
		instID = semconv.ServiceInstanceIDKey.String(uuid.New().String())
	}
    hostName:=os.Getenv("MY_POD_NAME")
   hostIp:=os.Getenv("MY_POD_IP")
   resourceType:=os.Getenv("RESOURCE_TYPE")
   return resource.New(
      context.Background(),
      resource.WithAttributes(
         instID,
         semconv.ServiceNameKey.String(serviceName),
         semconv.HostNameKey.String(hostName),
         label.String("service.namespace",serviceNameSpace),
         label.String("ip",hostIp),
         label.String("resource.type",resourceType),
      ),
   )
}
func spanExporter() (exporttrace.SpanExporter, error) {

	var user = os.Getenv("JAEGER_USER")
	var password = os.Getenv("JAEGER_PASSWORD")
	export_type := os.Getenv("EXPORT_TYPE")
	if export_type == "JAEGER" {
		log.Info("exporting with JAEGER logger")
		addr1 := os.Getenv("JAEGER_ENDPOINT")
		return jaeger.NewRawExporter(
			jaeger.WithCollectorEndpoint(addr1,jaeger.WithUsername(user),jaeger.WithPassword(password)),
			jaeger.WithProcess(jaeger.Process{
				ServiceName: serviceName,
			}),
		)
	}
	if export_type == "OTLP" {
		log.Info("exporting with OTLP logger")
		if os.Getenv("OTLP_ENDPOINT") != "" {
			addr1 := os.Getenv("OTLP_ENDPOINT")
			return otlp.NewExporter(
				context.Background(),
				otlp.WithInsecure(),
				otlp.WithAddress(addr1),
			)
		}
	}

	log.Info("exporting with STDOUT logger")
	return stdout.NewExporter(
		stdout.WithPrettyPrint(),
		stdout.WithWriter(log.Writer()),
	)
}
func initTracing() {
	if os.Getenv("DISABLE_TRACING") != "" {
		log.Info("tracing disabled")
		return
	}

	res, err := detectResource()
	if err != nil {
		log.WithError(err).Fatal("failed to detect environment resource")
	}

	exp, err := spanExporter()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize Span exporter")
		return
	}

	log.Info("tracing enabled")
	otel.SetTracerProvider(
		trace.NewTracerProvider(
			trace.WithConfig(
				trace.Config{
					DefaultSampler: trace.AlwaysSample(),
					Resource:       res,
				},
			),
			trace.WithSpanProcessor(
				trace.NewBatchSpanProcessor(exp),
			),
		),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func main() {

    if serviceName=os.Getenv("SERVICE_NAME");serviceName==""{
        serviceName = "checkout-service"
    }

    if serviceNameSpace=os.Getenv("SERVICE_NAMESPACE");serviceNameSpace==""{
        serviceNameSpace = "hipster"
    }

	var port = os.Getenv("PORT")
	initTracing()

	var PRODUCT_CATALOG_SERVICE_ADDR = os.Getenv("PRODUCT_CATALOG_SERVICE_ADDR") // "productcatlog:4000"
	var CURRENCY_SERVICE_ADDR = os.Getenv("CURRENCY_SERVICE_ADDR")               //"currency:9000"
	var CART_SERVICE_ADDR = os.Getenv("CART_SERVICE_ADDR")                       //"cart:80"
	var SHIPPING_SERVICE_ADDR = os.Getenv("SHIPPING_SERVICE_ADDR")               //"shipping:50051"
	var PAYMENT_SERVICE_ADDR = os.Getenv("PAYMENT_SERVICE_ADDR")                 //"payment:8011"
	var EMAIL_SERVICE_ADDR = os.Getenv("EMAIL_SERVICE_ADDR")                     //"email:4009"

	svc := checkoutserviceConstructor(PRODUCT_CATALOG_SERVICE_ADDR, CURRENCY_SERVICE_ADDR, CART_SERVICE_ADDR, SHIPPING_SERVICE_ADDR, PAYMENT_SERVICE_ADDR, EMAIL_SERVICE_ADDR)

	log.Infof("service config: %+v", svc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	var srv *grpc.Server
	srv = grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	pb.RegisterCheckoutServiceServer(srv, svc)
	healthpb.RegisterHealthServer(srv, svc)
	log.Infof("starting to listen on tcp: %q", lis.Addr().String())
	err = srv.Serve(lis)
	log.Fatal(err)
}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", envKey))
	}
	*target = v
}

func (cs *checkoutService) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (cs *checkoutService) Watch(req *healthpb.HealthCheckRequest, ws healthpb.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

func (cs *checkoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	log.Infof("[PlaceOrder] user_id=%q user_currency=%q", req.UserId, req.UserCurrency)

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate order uuid")
	}

	prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(ctx, req.UserId, req.UserCurrency, req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	total := pb.Money{CurrencyCode: req.UserCurrency,
		Units: 0,
		Nanos: 0}
	total = money.Must(money.Sum(total, *prep.shippingCostLocalized))
	for _, it := range prep.orderItems {
		multPrice := money.MultiplySlow(*it.Cost, uint32(it.GetItem().GetQuantity()))
		total = money.Must(money.Sum(total, multPrice))
	}

	txID, err := cs.chargeCard(ctx, &total, req.CreditCard)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to charge card: %+v", err)
	}
	log.Infof("payment went through (transaction_id: %s)", txID)

	shippingTrackingID, err := cs.shipOrder(ctx, req.Address, prep.cartItems)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "shipping error: %+v", err)
	}

	_ = cs.emptyUserCart(ctx, req.UserId)

	orderResult := &pb.OrderResult{
		OrderId:            orderID.String(),
		ShippingTrackingId: shippingTrackingID,
		ShippingCost:       prep.shippingCostLocalized,
		ShippingAddress:    req.Address,
		Items:              prep.orderItems,
	}

	if err := cs.sendOrderConfirmation(ctx, req.Email, orderResult); err != nil {
		log.Warnf("failed to send order confirmation to %q: %+v", req.Email, err)
	} else {
		log.Infof("order confirmation email sent to %q", req.Email)
	}
	resp := &pb.PlaceOrderResponse{Order: orderResult}
	return resp, nil
}

type orderPrep struct {
	orderItems            []*pb.OrderItem
	cartItems             []*pb.CartItem
	shippingCostLocalized *pb.Money
}

func (cs *checkoutService) prepareOrderItemsAndShippingQuoteFromCart(ctx context.Context, userID, userCurrency string, address *pb.Address) (orderPrep, error) {
	var out orderPrep
	cartItems, err := cs.getUserCart(ctx, userID)
	if err != nil {
		return out, fmt.Errorf("cart failure: %+v", err)
	}
	orderItems, err := cs.prepOrderItems(ctx, cartItems, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to prepare order: %+v", err)
	}
	shippingUSD, err := cs.quoteShipping(ctx, address, cartItems)
	if err != nil {
		return out, fmt.Errorf("shipping quote failure: %+v", err)
	}
	shippingPrice, err := cs.convertCurrency(ctx, shippingUSD, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to convert shipping cost to currency: %+v", err)
	}

	out.shippingCostLocalized = shippingPrice
	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (cs *checkoutService) quoteShipping(ctx context.Context, address *pb.Address, items []*pb.CartItem) (*pb.Money, error) {
	/*conn, err := grpc.DialContext(ctx, cs.shippingSvcAddr,
	grpc.WithInsecure())*/
	conn, err := grpc.Dial(cs.shippingSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("could not connect shipping service: %+v", err)
	}
	defer conn.Close()

	shippingQuote, err := pb.NewShippingServiceClient(conn).
		GetQuote(ctx, &pb.GetQuoteRequest{
			Address: address,
			Items:   items})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping quote: %+v", err)
	}
	return shippingQuote.GetCostUsd(), nil
}

func (cs *checkoutService) getUserCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	//conn, err := grpc.DialContext(ctx, cs.cartSvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.cartSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return nil, fmt.Errorf("could not connect cart service: %+v", err)
	}
	defer conn.Close()

	cart, err := pb.NewCartServiceClient(conn).GetCart(ctx, &pb.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart during checkout: %+v", err)
	}
	return cart.GetItems(), nil
}

func (cs *checkoutService) emptyUserCart(ctx context.Context, userID string) error {
	//conn, err := grpc.DialContext(ctx, cs.cartSvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.cartSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return fmt.Errorf("could not connect cart service: %+v", err)
	}
	defer conn.Close()

	if _, err = pb.NewCartServiceClient(conn).EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userID}); err != nil {
		return fmt.Errorf("failed to empty user cart during checkout: %+v", err)
	}
	return nil
}

func (cs *checkoutService) prepOrderItems(ctx context.Context, items []*pb.CartItem, userCurrency string) ([]*pb.OrderItem, error) {
	out := make([]*pb.OrderItem, len(items))

	//conn, err := grpc.DialContext(ctx, cs.productCatalogSvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.productCatalogSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return nil, fmt.Errorf("could not connect product catalog service: %+v", err)
	}
	defer conn.Close()
	cl := pb.NewProductCatalogServiceClient(conn)

	for i, item := range items {
		product, err := cl.GetProduct(ctx, &pb.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			return nil, fmt.Errorf("failed to get product #%q", item.GetProductId())
		}
		price, err := cs.convertCurrency(ctx, product.GetPriceUsd(), userCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to convert price of %q to %s", item.GetProductId(), userCurrency)
		}
		out[i] = &pb.OrderItem{
			Item: item,
			Cost: price}
	}
	return out, nil
}

func (cs *checkoutService) convertCurrency(ctx context.Context, from *pb.Money, toCurrency string) (*pb.Money, error) {
	//conn, err := grpc.DialContext(ctx, cs.currencySvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.currencySvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return nil, fmt.Errorf("could not connect currency service: %+v", err)
	}
	defer conn.Close()
	result, err := pb.NewCurrencyServiceClient(conn).Convert(context.TODO(), &pb.CurrencyConversionRequest{
		From:   from,
		ToCode: toCurrency})
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %+v", err)
	}
	return result, err
}

func (cs *checkoutService) chargeCard(ctx context.Context, amount *pb.Money, paymentInfo *pb.CreditCardInfo) (string, error) {
	//conn, err := grpc.DialContext(ctx, cs.paymentSvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.paymentSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return "", fmt.Errorf("failed to connect payment service: %+v", err)
	}
	defer conn.Close()

	paymentResp, err := pb.NewPaymentServiceClient(conn).Charge(ctx, &pb.ChargeRequest{
		Amount:     amount,
		CreditCard: paymentInfo})
	if err != nil {
		return "", fmt.Errorf("could not charge the card: %+v", err)
	}
	return paymentResp.GetTransactionId(), nil
}

func (cs *checkoutService) sendOrderConfirmation(ctx context.Context, email string, order *pb.OrderResult) error {
	//conn, err := grpc.DialContext(ctx, cs.emailSvcAddr, grpc.WithInsecure())

	conn, err := grpc.Dial(cs.emailSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		return fmt.Errorf("failed to connect email service: %+v", err)
	}
	defer conn.Close()
	_, err = pb.NewEmailServiceClient(conn).SendOrderConfirmation(ctx, &pb.SendOrderConfirmationRequest{
		Email: email,
		Order: order})
	return err
}

func (cs *checkoutService) shipOrder(ctx context.Context, address *pb.Address, items []*pb.CartItem) (string, error) {
	//conn, err := grpc.DialContext(ctx, cs.shippingSvcAddr, grpc.WithInsecure())
	conn, err := grpc.Dial(cs.shippingSvcAddr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to connect email service: %+v", err)
	}
	defer conn.Close()
	resp, err := pb.NewShippingServiceClient(conn).ShipOrder(ctx, &pb.ShipOrderRequest{
		Address: address,
		Items:   items})
	if err != nil {
		return "", fmt.Errorf("shipment failed: %+v", err)
	}
	return resp.GetTrackingId(), nil
}

// TODO: Dial and create client once, reuse.
