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

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maklybae/goshop/payment/db"
	paymentpb "github.com/maklybae/goshop/payment/genproto/payment"
	"github.com/maklybae/goshop/payment/internal/config"
	"github.com/maklybae/goshop/payment/internal/infrastructure/repository/postgres"
	"github.com/maklybae/goshop/payment/internal/producer"
	"github.com/maklybae/goshop/payment/internal/service"
	grpcapi "github.com/maklybae/goshop/payment/internal/transport/grpc"
	"github.com/maklybae/goshop/shared/inbox"
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
	inboxRepo := inbox.NewPostgresRepository(pool)
	transactor := txs.NewTxBeginner(pool)
	paymentService := service.NewPaymentService(repo)

	// Kafka producer для outbox worker
	prod, err := producer.NewKafkaSyncProducer([]string{config.KafkaBroker})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer prod.Close()

	// Dummy workers (Kafka -> Inbox)
	groupID := "payment-inbox-group"
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup([]string{config.KafkaBroker}, groupID, saramaCfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer group: %v", err)
	}
	defer consumerGroup.Close()

	// Dummy workers (Kafka -> Inbox)
	for range 1 {
		worker := inbox.NewDummyWorker(consumerGroup, inboxRepo, "order.payments", groupID)
		go func() {
			if err := worker.Start(ctx); err != nil {
				log.Printf("dummy worker error: %v", err)
			}
		}()
	}

	// Clever workers (Inbox -> бизнес-логика)
	handler := service.NewPaymentHandler(repo, outboxRepo)
	for range 1 {
		clever := inbox.NewCleverWorker(inboxRepo, 10, handler.Handle, transactor)
		go func() {
			clever.Start(ctx, 5*time.Second)
		}()
	}

	// Outbox workers (Outbox -> Kafka)
	outboxConfig := &outbox.Config{Period: 5 * time.Second, BatchSize: 10}
	for range 1 {
		worker := outbox.NewWorker(prod, transactor, outboxRepo, outboxConfig)
		worker.StartProcessingEvents(ctx)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v\n", config.Port, err)
	}

	grpcServer := grpc.NewServer()
	paymentHandler := grpcapi.NewPaymentGRPCServer(paymentService)
	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentHandler)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	grpcServer.GracefulStop()
	log.Println("Received shutdown signal, stopping payment server...")
}
