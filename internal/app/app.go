package app

import (
	"fmt"
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
	appScore "github.com/nevinmanoj/bhavana-backend/internal/app/score"
	appTeam "github.com/nevinmanoj/bhavana-backend/internal/app/team"
	appUser "github.com/nevinmanoj/bhavana-backend/internal/app/user"

	"github.com/nevinmanoj/bhavana-backend/internal/rbac"

	repoAccess "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/access"
	repoEvent "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/event"
	repoSchool "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/school"
	repoScore "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/score"
	repoTeam "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/team"
	repoUser "github.com/nevinmanoj/bhavana-backend/internal/db/postgres/user"

	domainAccess "github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	domainEvent "github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	domainSchool "github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	domainScore "github.com/nevinmanoj/bhavana-backend/internal/domain/score"
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
	repoAccess := repoAccess.NewAccessRepository()
	userReadRepo := repoUser.NewUserReadRepository()
	userWriteRepo := repoUser.NewUserWriteRepository()
	eventWriteRepo := repoEvent.NewEventWriteRepository()
	eventReadRepo := repoEvent.NewEventReadRepository()
	schoolWriteRepo := repoSchool.NewSchoolWriteRepository()
	schoolReadRepo := repoSchool.NewSchoolReadRepository()
	teamWrietRepo := repoTeam.NewTeamWriteRepository()
	scoreWriteRepo := repoScore.NewScoreWriteRepository()

	//Services
	accessService := domainAccess.NewAccessService(dbConn, repoAccess)
	userService := domainUser.NewUserService(dbConn, jwtSecretbyte, userWriteRepo)
	eventService := domainEvent.NewEventService(dbConn, eventWriteRepo, userReadRepo)
	schoolService := domainSchool.NewSchoolService(dbConn, accessService, schoolWriteRepo)
	teamService := domainTeam.NewTeamService(dbConn, accessService, teamWrietRepo, eventReadRepo, schoolReadRepo)
	scoreService := domainScore.NewScoreService(dbConn, accessService, scoreWriteRepo)

	//Handlers
	userHandler := appUser.NewUserHandler(userService, validator)
	eventHandler := appEvent.NewEventHandler(eventService, validator)
	schoolHandler := appSchool.NewSchoolHandler(schoolService, validator)
	teamHandler := appTeam.NewTeamHandler(teamService, validator)
	scoreHandler := appScore.NewSchoolHandler(scoreService, validator)

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
		router.With(middleware.RequirePermission(rbac.PermUpdateEvent)).Put("/{eventId}/status", eventHandler.UpdateEventStatus)
		router.With(middleware.RequirePermission(rbac.PermDeleteEvent)).Delete("/{eventId}", eventHandler.DeleteEvent)
		router.With(middleware.RequirePermission(rbac.PermViewScore)).Get("/{eventId}/scores", scoreHandler.GetScoresByEventID)
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

	// Teams routes
	r.Route("/teams", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		router.With(middleware.RequirePermission(rbac.PermViewTeam)).Get("/", teamHandler.GetTeams)
		router.With(middleware.RequirePermission(rbac.PermViewTeam)).Get("/{teamId}", teamHandler.GetTeam)
		router.With(middleware.RequirePermission(rbac.PermCreateTeam)).Post("/", teamHandler.CreateTeam)
		router.With(middleware.RequirePermission(rbac.PermUpdateTeam)).Put("/{teamId}", teamHandler.UpdateTeam)
		router.With(middleware.RequirePermission(rbac.PermDeleteTeam)).Delete("/{teamId}", teamHandler.DeleteTeam)

	})
	r.Route("/scores", func(router chi.Router) {
		router.Use(authMiddleware, middleware.InjectScope)
		// router.With(middleware.RequirePermission(rbac.PermViewTeam)).Get("/", teamHandler.GetTeams)
		router.With(middleware.RequirePermission(rbac.PermViewScore)).Get("/{scoreId}", scoreHandler.GetScore)
		router.With(middleware.RequirePermission(rbac.PermCreateScore)).Post("/", scoreHandler.CreateScores)
		router.With(middleware.RequirePermission(rbac.PermUpdateScore)).Put("/", scoreHandler.UpdateScores)
		router.With(middleware.RequirePermission(rbac.PermDeleteScore)).Delete("/{scoreId}", scoreHandler.DeleteScore)

	})

	fmt.Println("Serving on port 8080")
	return http.ListenAndServe(":8080", r)
}
