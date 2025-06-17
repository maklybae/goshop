package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/order/db"
	orderpb "github.com/maklybae/goshop/order/genproto/order"
	"github.com/maklybae/goshop/order/internal/config"
	"github.com/maklybae/goshop/order/internal/consumer"
	"github.com/maklybae/goshop/order/internal/infrastructure/repository/postgres"
	"github.com/maklybae/goshop/order/internal/producer"
	"github.com/maklybae/goshop/order/internal/service"
	grpcapi "github.com/maklybae/goshop/order/internal/transport/grpc"
	"github.com/maklybae/goshop/shared/outbox"
	"github.com/maklybae/goshop/shared/txs"
	"google.golang.org/grpc"
)

func main() {
	config := config.NewConfig()

	connConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v\n", err)
	}

	db.MustMigrate(connConfig.ConnConfig)

	repo := postgres.NewRepository(pool)
	outboxRepo := outbox.NewPostgresRepository(pool)
	transactor := txs.NewTxBeginner(pool)
	orderService := service.NewOrderService(outboxRepo, transactor, repo)

	producer, err := producer.NewKafkaSyncProducer([]string{config.KafkaBroker})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	groupID := "order-consumer-group"
	kafkaConsumer, err := consumer.NewKafkaGroupConsumer([]string{config.KafkaBroker}, groupID, "payment.completed", orderService.HandlePaymentEvent)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer group: %v", err)
	}
	defer kafkaConsumer.Close()
	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer stopped: %v", err)
		}
	}()

	outboxConfig := &outbox.Config{
		Period:    2 * time.Second,
		BatchSize: 10,
	}

	for range 5 {
		worker := outbox.NewWorker(producer, transactor, outboxRepo, outboxConfig)
		worker.StartProcessingEvents(ctx)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v\n", config.Port, err)
	}

	grpcServer := grpc.NewServer()
	orderHandler := grpcapi.NewOrderGRPCServer(orderService)
	orderpb.RegisterOrderServiceServer(grpcServer, orderHandler)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	grpcServer.GracefulStop()
	log.Println("Received shutdown signal, stopping order server...")
}
