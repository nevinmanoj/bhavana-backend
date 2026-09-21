package app

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	postgres "github.com/nevinmanoj/bhavana-backend/internal/db/postgres"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/validation"

	appEntry "github.com/nevinmanoj/bhavana-backend/internal/app/entry"
	appEvent "github.com/nevinmanoj/bhavana-backend/internal/app/event"
	appResult "github.com/nevinmanoj/bhavana-backend/internal/app/result"
	appSchool "github.com/nevinmanoj/bhavana-backend/internal/app/school"
	appScore "github.com/nevinmanoj/bhavana-backend/internal/app/score"
	appUser "github.com/nevinmanoj/bhavana-backend/internal/app/user"

	"github.com/nevinmanoj/bhavana-backend/internal/rbac"

	repoAccess "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/access"
	repoEntry "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/entry"
	repoEvent "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/event"
	repoResult "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/result"
	repoSchool "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/school"
	repoScore "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/score"
	repoUser "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/user"

	domainAccess "github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	domainEntry "github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
	domainEvent "github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	domainResult "github.com/nevinmanoj/bhavana-backend/internal/domain/result"
	domainSchool "github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	domainScore "github.com/nevinmanoj/bhavana-backend/internal/domain/score"
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
	repoAccess := repoAccess.NewAccessRepository()
	userReadRepo := repoUser.NewUserReadRepository()
	userWriteRepo := repoUser.NewUserWriteRepository()
	eventWriteRepo := repoEvent.NewEventWriteRepository()
	eventReadRepo := repoEvent.NewEventReadRepository()
	schoolWriteRepo := repoSchool.NewSchoolWriteRepository()
	schoolReadRepo := repoSchool.NewSchoolReadRepository()
	entryWriteRepo := repoEntry.NewEntryWriteRepository()
	scoreWriteRepo := repoScore.NewScoreWriteRepository()
	resultWriteRepo := repoResult.NewResultWriteRepository()

	//Services
	accessService := domainAccess.NewAccessService(dbConn, repoAccess)
	userService := domainUser.NewUserService(dbConn, jwtSecretbyte, userWriteRepo)
	resultService := domainResult.NewResultService(dbConn, resultWriteRepo)
	eventService := domainEvent.NewEventService(dbConn, eventWriteRepo, userReadRepo, resultService)
	schoolService := domainSchool.NewSchoolService(dbConn, accessService, schoolWriteRepo)
	entryService := domainEntry.NewEntryService(dbConn, accessService, entryWriteRepo, eventReadRepo, schoolReadRepo)
	scoreService := domainScore.NewScoreService(dbConn, accessService, scoreWriteRepo)

	//Handlers
	userHandler := appUser.NewUserHandler(userService, validator)
	eventHandler := appEvent.NewEventHandler(eventService, validator)
	schoolHandler := appSchool.NewSchoolHandler(schoolService, validator)
	entryHandler := appEntry.NewEntryHandler(entryService, validator)
	scoreHandler := appScore.NewSchoolHandler(scoreService, validator)
	resultHandler := appResult.NewResultHandler(resultService, validator)

	//CORS
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {

			if strings.Contains(origin, "521d7c39-a391-4c5e-9803-b598925e3ada") {
				return true
			}
			if origin == "http://localhost:8081" {
				return true
			}
			if origin == "https://gray-stone-0cf1a6200.2.azurestaticapps.net" {
				return true
			}
			return false
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))

	//health and root endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("HI there! Welcome to Bhavana Backend"))
	})

	//User routes
	r.Route("/users", func(router chi.Router) {
		// public
		router.Post("/login", userHandler.LoginUser)
		router.Post("/register", userHandler.CreateUser)

		// protected
		router.Group(func(groupRouter chi.Router) {
			groupRouter.Use(authMiddleware, middleware.InjectScope)
			groupRouter.Get("/me", userHandler.GetMe)
			groupRouter.With(middleware.RequirePermission(rbac.PermViewUser)).Get("/", userHandler.GetUsers)
			groupRouter.With(middleware.RequirePermission(rbac.PermViewUser)).Get("/{userId}", userHandler.GetUser)
		})
	})

	//Event routes
	r.Route("/events", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewEvent)).Get("/", eventHandler.GetEvents)
		router.With(middleware.RequirePermission(rbac.PermViewEvent)).Get("/{eventId}", eventHandler.GetEvent)
		router.With(middleware.RequirePermission(rbac.PermCreateEvent)).Post("/", eventHandler.CreateEvent)
		router.With(middleware.RequirePermission(rbac.PermUpdateEvent)).Put("/{eventId}", eventHandler.UpdateEvent)
		router.With(middleware.RequirePermission(rbac.PermUpdateEventStatus)).Put("/{eventId}/status", eventHandler.UpdateEventStatus)
		router.With(middleware.RequirePermission(rbac.PermDeleteEvent)).Delete("/{eventId}", eventHandler.DeleteEvent)
		router.With(middleware.RequirePermission(rbac.PermViewScore)).Get("/{eventId}/scores", scoreHandler.GetScoresByEventID)
		router.With(middleware.RequirePermission(rbac.PermViewResult)).Get("/{eventId}/results", resultHandler.GetEventResults)
		router.With(middleware.RequirePermission(rbac.PermViewResult)).Get("/{eventId}/results/readiness", resultHandler.GetReadiness)
	})

	//School and student routes
	r.Route("/schools", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewSchool)).Get("/", schoolHandler.GetSchools)
		router.With(middleware.RequirePermission(rbac.PermViewSchool)).Get("/{schoolId}", schoolHandler.GetSchool)
		router.With(middleware.RequirePermission(rbac.PermCreateSchool)).Post("/", schoolHandler.CreateSchool)
		router.With(middleware.RequirePermission(rbac.PermUpdateSchool)).Put("/{schoolId}", schoolHandler.UpdateSchool)
		router.With(middleware.RequirePermission(rbac.PermDeleteSchool)).Delete("/{schoolId}", schoolHandler.DeleteSchool)

		router.Route("/{schoolId}/students", func(studentRouter chi.Router) {
			studentRouter.With(middleware.RequirePermission(rbac.PermViewStudent)).Get("/", schoolHandler.GetStudentsBySchoolID)
			studentRouter.With(middleware.RequirePermission(rbac.PermCreateStudent)).Post("/", schoolHandler.CreateStudent)
			studentRouter.With(middleware.RequirePermission(rbac.PermUpdateStudent)).Put("/{studentId}", schoolHandler.UpdateStudent)
			studentRouter.With(middleware.RequirePermission(rbac.PermDeleteStudent)).Delete("/{studentId}", schoolHandler.DeleteStudent)

		})
	})

	r.Route("/students", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewStudent)).Get("/", schoolHandler.GetStudents)
	})

	// Entries routes
	r.Route("/entries", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewEntry)).Get("/", entryHandler.GetEntries)
		router.With(middleware.RequirePermission(rbac.PermViewEntry)).Get("/{entryId}", entryHandler.GetEntry)
		router.With(middleware.RequirePermission(rbac.PermCreateEntry)).Post("/", entryHandler.CreateEntry)
		router.With(middleware.RequirePermission(rbac.PermUpdateEntry)).Put("/{entryId}", entryHandler.UpdateEntry)
		router.With(middleware.RequirePermission(rbac.PermDeleteEntry)).Delete("/{entryId}", entryHandler.DeleteEntry)

	})
	r.Route("/scores", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		// router.With(middleware.RequirePermission(rbac.PermViewEntry)).Get("/", entryHandler.GetEntries)
		router.With(middleware.RequirePermission(rbac.PermViewScore)).Get("/{scoreId}", scoreHandler.GetScore)
		router.With(middleware.RequirePermission(rbac.PermCreateScore)).Post("/", scoreHandler.CreateScores)
		router.With(middleware.RequirePermission(rbac.PermUpdateScore)).Put("/", scoreHandler.UpdateScores)
		router.With(middleware.RequirePermission(rbac.PermDeleteScore)).Delete("/{scoreId}", scoreHandler.DeleteScore)

	})

	// Leaderboard route
	r.Route("/leaderboard", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewResult)).Get("/", resultHandler.GetLeaderboard)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local run
	}

	fmt.Println("Serving on port " + port)
	return http.ListenAndServe(":"+port, r)
}
