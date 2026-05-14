package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/go-co-op/gocron/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tink3rlabs/magic/errors"
	"github.com/tink3rlabs/magic/health"
	"github.com/tink3rlabs/magic/leadership"
	"github.com/tink3rlabs/magic/logger"
	"github.com/tink3rlabs/magic/middlewares"
	"github.com/tink3rlabs/magic/observability"
	"github.com/tink3rlabs/magic/pubsub"
	"github.com/tink3rlabs/magic/storage"
	"github.com/tink3rlabs/magic/telemetry"

	"todo-service/pkg/routes"
)

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Run the ToDo server",
	RunE:  runServer,
}

func init() {
	serverCommand.Flags().StringP("port", "p", "8080", "The port on which the Todo server will listen on")
}

func initRoutes(obs *observability.Observer, todosCreated telemetry.Counter, auth routes.AuthConfig, pubSub routes.PubSubConfig) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger,          // Log API request calls
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
		middlewares.ObservabilityWithOptions(obs, middlewares.ObservabilityOptions{
			SkipPaths:        []string{"/metrics"},
			SkipPathPrefixes: []string{"/health/"},
		}),
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	)

	t := routes.NewTodoRouter(todosCreated, auth, pubSub)
	router.Route("/", func(r chi.Router) {
		r.Mount("/todos", t.Router)
	})

	return router
}

func createScheduler() {
	slog.Info("strating scheduler")
	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("failed to create scheduler", slog.Any("error", err))
	}
	// add a job to the scheduler
	_, err = s.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(
			func(param string) {
				slog.Info("scheduled job says", slog.String("param", param))
			},
			"hello",
		),
	)
	if err != nil {
		logger.Fatal("failed to create scheduled job", slog.Any("error", err))
	}

	// start the scheduler
	s.Start()
}

func runServer(cmd *cobra.Command, args []string) error {
	openApiSpec, err := ConfigFS.ReadFile("config/openapi.json")

	if err != nil {
		return fmt.Errorf("failed to load OpenAPI definition, did you forget to run go generate?: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	obsCfg := observability.DefaultConfig()
	obsCfg.ServiceName = viper.GetString("observability.service_name")
	if obsCfg.ServiceName == "" {
		obsCfg.ServiceName = "todo-service"
	}
	switch viper.GetString("observability.metrics_mode") {
	case "otlp":
		obsCfg.MetricsMode = observability.MetricsModeOTLP
		obsCfg.MetricsOTLPEndpoint = viper.GetString("observability.metrics_otlp_endpoint")
	default:
		obsCfg.MetricsMode = observability.MetricsModePrometheus
	}
	obsCfg.EnableTracing = viper.GetBool("observability.enable_tracing")
	obsCfg.TracesOTLPEndpoint = viper.GetString("observability.traces_otlp_endpoint")

	obs, err := observability.Init(ctx, obsCfg)
	if err != nil {
		logger.Fatal("failed to initialise observability", slog.String("error", err.Error()))
	}
	defer func() { _ = obs.Shutdown(context.Background()) }()

	todosCreated, err := obs.Counter(telemetry.MetricDefinition{
		Name: "todo_service_todos_created_total",
		Help: "Total number of todo items created.",
		Kind: telemetry.KindCounter,
	})
	if err != nil {
		logger.Fatal("failed to register todos_created counter", slog.String("error", err.Error()))
	}

	storageAdapter, err := storage.StorageAdapterFactory{}.GetInstance(
		storage.StorageAdapterType(viper.GetString("storage.type")),
		viper.GetStringMapString("storage.config"),
	)

	if err != nil {
		panic("failed to get storage adapter instance")
	}

	storage.NewDatabaseMigration(storageAdapter).Migrate()

	var publisher pubsub.Publisher
	if viper.GetBool("pubsub.enabled") {
		publisher, err = pubsub.PublisherFactory{}.GetInstance(pubsub.SNS, map[string]string{
			"region": viper.GetString("pubsub.region"),
		})
		if err != nil {
			logger.Fatal("failed to create pub/sub publisher", slog.String("error", err.Error()))
		}
	}

	electionProps := leadership.LeaderElectionProps{
		HeartbeatInterval: viper.GetDuration("leadership.heartbeat"),
		StorageAdapter:    storageAdapter,
		AdditionalProps: map[string]any{
			"global": viper.GetBool("storage.config.global"),
			"region": viper.GetString("storage.config.region"),
			"regios": viper.GetStringSlice("storage.config.regions"),
		},
	}
	election := leadership.NewLeaderElection(electionProps)
	election.Start()

	go func() {
		for result := range election.Results {
			if result == leadership.RESULT_ELECTED {
				createScheduler()
			}
		}
	}()

	authEnabled := viper.GetBool("auth.enabled")
	authMiddleware := middlewares.EnsureValidToken(middlewares.EnsureValidTokenConfig{
		Enabled:   authEnabled,
		IssuerURL: viper.GetString("auth.issuer_url"),
		Audience:  viper.GetStringSlice("auth.audience"),
	})
	authCfg := routes.AuthConfig{
		Middleware: authMiddleware,
		Enabled:    authEnabled,
		WriteRole:  viper.GetString("auth.write_role"),
	}

	pubSubCfg := routes.PubSubConfig{
		Publisher: publisher,
		TopicARN:  viper.GetString("pubsub.topic_arn"),
	}

	router := initRoutes(obs, todosCreated, authCfg, pubSubCfg)

	router.Handle("/metrics", obs.MetricsHandler())

	router.Get("/api-docs", func(w http.ResponseWriter, r *http.Request) {
		if _, responseFailed := w.Write(openApiSpec); responseFailed != nil {
			slog.Error("failed responding to /api-docs:", slog.Any("error", responseFailed))
		}
	})

	//health check - liveness
	router.Get("/health/liveness", func(w http.ResponseWriter, r *http.Request) {
		render.Status(r, http.StatusNoContent)
		render.NoContent(w, r)
	})

	//health check - readiness
	healthChecker := health.NewHealthChecker(storageAdapter)
	h := middlewares.ErrorHandler{}
	router.Get("/health/readiness", h.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		err := healthChecker.Check(viper.GetBool("health.storage"), viper.GetStringSlice("health.dependencies"))
		if err != nil {
			slog.Error("health check readiness failed", slog.Any("error", err.Error()))
			return &errors.ServiceUnavailable{Message: err.Error()}
		} else {
			render.Status(r, http.StatusNoContent)
			render.NoContent(w, r)
			return nil
		}
	}))

	port := viper.GetString("service.port")
	listenAddress := fmt.Sprintf(":%s", port)

	srv := &http.Server{Addr: listenAddress, Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", slog.Any("error", err))
		}
	}()
	slog.Info("todo-service listening", slog.String("address", listenAddress))

	<-ctx.Done()
	slog.Info("shutdown signal received, stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", slog.Any("error", err))
	}
	return nil
}
