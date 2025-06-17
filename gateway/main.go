package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	orderpb "github.com/maklybae/goshop/order/genproto/order"
	paymentpb "github.com/maklybae/goshop/payment/genproto/payment"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	// Прокси для order
	if err := orderpb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, "order:50051", []grpc.DialOption{grpc.WithInsecure()}); err != nil {
		log.Fatalf("failed to register order gateway: %v", err)
	}

	// Прокси для payment
	if err := paymentpb.RegisterPaymentServiceHandlerFromEndpoint(ctx, mux, "payment:50051", []grpc.DialOption{grpc.WithInsecure()}); err != nil {
		log.Fatalf("failed to register payment gateway: %v", err)
	}

	log.Println("HTTP gateway listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
