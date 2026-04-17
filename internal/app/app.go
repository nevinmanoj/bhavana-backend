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
	appSchool "github.com/nevinmanoj/bhavana-backend/internal/app/school"
	appTeam "github.com/nevinmanoj/bhavana-backend/internal/app/team"
	appUser "github.com/nevinmanoj/bhavana-backend/internal/app/user"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"

	repoEvent "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/event"
	repoSchool "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/school"
	repoTeam "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/team"
	repoUser "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/user"

	domainEvent "github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	domainSchool "github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	domainTeam "github.com/nevinmanoj/bhavana-backend/internal/domain/team"
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
	eventReadRepo := repoEvent.NewEventReadRepository()
	schoolWriteRepo := repoSchool.NewSchoolWriteRepository()
	schoolReadRepo := repoSchool.NewSchoolReadRepository()
	teamWrietRepo := repoTeam.NewTeamWriteRepository()

	//Services
	userService := domainUser.NewUserService(userWriteRepo, jwtSecretbyte, dbConn)
	eventService := domainEvent.NewEventService(eventWriteRepo, userReadRepo, dbConn)
	schoolService := domainSchool.NewSchoolService(schoolWriteRepo, dbConn)
	teamService := domainTeam.NewTeamService(dbConn, teamWrietRepo, eventReadRepo, schoolReadRepo)

	//Handlers
	userHandler := appUser.NewUserHandler(userService, validator)
	eventHandler := appEvent.NewEventHandler(eventService, validator)
	schoolHandler := appSchool.NewSchoolHandler(schoolService, validator)
	teamHandler := appTeam.NewEventHandler(teamService, validator)

	//CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8081"},
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
			groupRouter.Use(authMiddleware, middleware.InjectScope)
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
	})

	//School and student routes
	r.Route("/schools", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewSchool)).Get("/", schoolHandler.GetSchools)
		router.With(middleware.RequirePermission(rbac.PermViewSchool)).Get("/{schoolId}", schoolHandler.GetSchool)
		router.With(middleware.RequirePermission(rbac.PermCreateSchool)).Post("/", schoolHandler.CreateSchool)
		router.With(middleware.RequirePermission(rbac.PermUpdateSchool)).Put("/{schoolId}", schoolHandler.UpdateSchool)

		router.Route("/{schoolId}/students", func(studentRouter chi.Router) {
			studentRouter.With(middleware.RequirePermission(rbac.PermViewStudent)).Get("/", schoolHandler.GetStudentsBySchoolID)
			studentRouter.With(middleware.RequirePermission(rbac.PermCreateStudent)).Post("/", schoolHandler.CreateStudent)
			studentRouter.With(middleware.RequirePermission(rbac.PermUpdateStudent)).Put("/{studentId}", schoolHandler.UpdateStudent)
		})
	})

	r.Route("/students", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewStudent)).Get("/", schoolHandler.GetStudents)
	})

	// Teams routes
	r.Route("/teams", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewTeam)).Get("/", teamHandler.GetTeams)
		router.With(middleware.RequirePermission(rbac.PermViewTeam)).Get("/{teamId}", teamHandler.GetTeam)
		router.With(middleware.RequirePermission(rbac.PermCreateTeam)).Post("/", teamHandler.CreateTeam)
		router.With(middleware.RequirePermission(rbac.PermUpdateTeam)).Put("/{teamId}", teamHandler.UpdateTeam)
	})

	return http.ListenAndServe(":8080", r)
}
