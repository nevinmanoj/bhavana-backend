package app

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	postgres "github.com/nevinmanoj/bhavana-backend/internal/db/postgres"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/validation"

	appEvent "github.com/nevinmanoj/bhavana-backend/internal/app/event"
	appUser "github.com/nevinmanoj/bhavana-backend/internal/app/user"

	repoEvent "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/event"
	repoUser "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/user"

	domainEvent "github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	domainUser "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

func Start() error {
	//Router and db connection
	var r *chi.Mux = chi.NewRouter()

	//get connection strings and jwt secret
	dsn := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtSecretbyte := []byte(jwtSecret)

	//postgres
	dbConn := postgres.NewPostgres(dsn)

	//validator
	validator := validation.NewValidator()

	// Global middleware
	r.Use(chimiddle.StripSlashes)

	//auth middleware
	authMiddleware := middleware.Authorization(jwtSecretbyte)

	//Repos
	userReadRepo := repoUser.NewUserReadRepository()
	userWriteRepo := repoUser.NewUserWriteRepository()
	eventWriteRepo := repoEvent.NewEventWriteRepository()

	//Services
	userService := domainUser.NewUserService(userWriteRepo, jwtSecretbyte, dbConn)
	eventService := domainEvent.NewEventService(eventWriteRepo, userReadRepo, dbConn)

	//Handlers
	userHandler := appUser.NewUserHandler(userService, validator)
	eventHandler := appEvent.NewEventHandler(eventService, validator)

	//CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))

	//User routes
	r.Route("/users", func(router chi.Router) {
		// public
		router.Post("/login", userHandler.LoginUser)
		router.Post("/register", userHandler.CreateUser)

		// protected
		router.Group(func(groupRouter chi.Router) {
			groupRouter.Use(authMiddleware)
			groupRouter.Get("/", userHandler.GetUsers)
			groupRouter.Get("/{userId}", userHandler.GetUser)
		})
	})

	//Event routes
	r.Route("/events", func(router chi.Router) {
		router.Use(authMiddleware)
		router.Get("/", eventHandler.GetEvents)
		router.Get("/{eventId}", eventHandler.GetEvent)
		router.Post("/", eventHandler.CreateEvent)
		router.Put("/{eventId}", eventHandler.UpdateEvent)

	})

	return http.ListenAndServe(":8080", r)
}
