package app

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	postgres "github.com/nevinmanoj/bhavana-backend/internal/db/postgres"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"

	appUser "github.com/nevinmanoj/bhavana-backend/internal/app/user"
	repoUser "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/user"
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

	// Global middleware
	r.Use(chimiddle.StripSlashes)

	//auth middleware
	authMiddleware := middleware.Authorization(jwtSecretbyte)

	//Repos
	// userReadRepo := repoUser.NewUserReadRepository(dbConn)
	userWriteRepo := repoUser.NewUserWriteRepository(dbConn)

	//Services
	userService := domainUser.NewUserService(userWriteRepo, jwtSecretbyte)

	//Handlers
	userHandler := appUser.NewUserHandler(userService)

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
		router.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Get("/", userHandler.GetUsers)
			r.Get("/{userId}", userHandler.GetUser)
		})
	})
	return http.ListenAndServe(":8080", r)
}
