package cmd

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
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
	"github.com/tink3rlabs/magic/storage"

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

func initRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.Logger,          // Log API request calls
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
	)

	t := routes.NewTodoRouter()
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

	// Random sleep between 0 to 30 seconds to handle multiple instances starting at the same time
	sleep := rand.IntN(30)
	slog.Info("Sleeping to handle multiple instances starting at the same time", slog.Int("sleep_duration_sec", sleep))
	time.Sleep(time.Duration(sleep) * time.Second)

	storageAdapter, err := storage.StorageAdapterFactory{}.GetInstance(
		storage.StorageAdapterType(viper.GetString("storage.type")),
		viper.GetStringMapString("storage.config"),
	)

	if err != nil {
		panic("failed to get storage adapter instance")
	}

	storage.NewDatabaseMigration(storageAdapter).Migrate()

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

	router := initRoutes()

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
	return http.ListenAndServe(listenAddress, router)
}
