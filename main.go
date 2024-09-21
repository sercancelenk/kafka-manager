package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	observability "gitlab.trendyol.com/devx/observability/sdk/opentelemetry-go"
	_ "gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/docs"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/controller/consumer"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/controller/topic"
	consumer2 "gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/consumer"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/manager"
	topic2 "gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/topic"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/config"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)

	l, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("fail to build log. err: %s", err))
	}

	zap.ReplaceGlobals(l.With(zap.String("app", "kafka-janitor")))
}

//go:generate swag init --parseDependency --parseInternal

// @title			Kafka Stream Api
// @version		    1.0
// @description	    For any questions, please contact with PaaS - Data Streaming Team
// @contact.name	Infrastructure - PaaS - Data Streaming
// @contact.email	messaging-development@trendyol.com
// @BasePath		/api/v1/
func main() {
	configFile := flag.String("config", "./local/config.yml", "Path to config file")
	secretFile := flag.String("secret", "./local/secret.yml", "Path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configFile, *secretFile)
	if err != nil {
		zap.L().Panic("config read failed", zap.Error(err))
	}

	ctx := context.Background()
	closeFunc := observability.Initialize()
	defer closeFunc(ctx)

	app := fiber.New(fiber.Config{
		ErrorHandler: http.ErrorHandler(),
	})

	prometheus := fiberprometheus.New("kafka-stream-api")
	prometheus.RegisterAt(app, "/metrics")

	app.Use(pprof.New())
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: zap.L(),
	}))
	app.Use(prometheus.Middleware)
	app.Use(otelfiber.Middleware())
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(cors.New())
	app.Use(healthcheck.New())

	app.Get("/swagger/*", swagger.HandlerDefault)

	api := app.Group("/api")

	clientService := manager.NewAdminClientService(cfg.Clusters)

	consumerService, err := consumer2.NewConsumerService(zap.L(), cfg, clientService)
	if err != nil {
		zap.L().Fatal("failed to create consumer service", zap.Error(err))
	}
	consumer.NewConsumerController(api, consumerService)

	m := make(map[string]manager.ConsumerPool, len(cfg.Clusters))
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	for id, cluster := range cfg.Clusters {
		pool, err := manager.NewConsumerPool(10, cluster.Brokers, config)
		if err != nil {
			zap.L().Fatal("failed to create consumer pool", zap.Error(err))
		}
		m[id] = pool

	}
	poolManager := manager.NewConsumerPoolManager(m)
	topicService := topic2.NewTopicService(zap.L(), cfg, poolManager, clientService)
	topic.NewTopicController(api, topicService)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	doneCh := make(chan struct{})
	go func() {
		_ = <-c
		fmt.Println("Gracefully shutting down...")
		_ = app.Shutdown()
		close(doneCh)
	}()

	if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		zap.L().Panic("server listen failed", zap.Error(err))
	}

	<-doneCh
}
