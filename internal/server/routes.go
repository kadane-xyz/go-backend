package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/internal/api/handlers"
)

func (s *Server) RegisterApiRoutes() {
	s.mux.Route("/v1", func(r chi.Router) {
		accountsHandler := handlers.NewAccountHandler(s.container.Services.AccountService, s.container.AWSClient, s.container.Config)
		adminHandler := handlers.NewAdminHandler(s.container.Services.AdminService, s.container.Services.ProblemService, s.container.Judge0)
		problemHandler := handlers.NewProblemHandler(s.container.Services.ProblemService)
		submissionHandler := handlers.NewSubmissionHandler(s.container.Services.SubmissionService)
		solutionsHandler := handlers.NewSolutionsHandler(s.container.Repositories.SolutionRepo)
		starredHandler := handlers.NewStarredHandler(s.container.Repositories.StarredRepo)
		friendHandler := handlers.NewFriendHandler(s.container.Repositories.FriendRepo)

		// solutions
		s.mux.Route("/solutions", func(r chi.Router) {
			r.Get("/", solutionsHandler.GetSolutions)
			r.Post("/", solutionsHandler.CreateSolution)
			r.Route("/{solutionId}", func(r chi.Router) {
				r.Get("/", solutionsHandler.GetSolution)
				r.Put("/", solutionsHandler.UpdateSolution)
				r.Delete("/", solutionsHandler.DeleteSolution)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", solutionsHandler.VoteSolution)
				})
			})
		})
		// comments
		s.mux.Route("/comments", func(r chi.Router) {
			r.Get("/", commentHandler.GetComments)
			r.Post("/", commentHandler.CreateComment)
			r.Route("/{commentId}", func(r chi.Router) {
				r.Get("/", commentHandler.GetComment)
				r.Put("/", commentHandler.UpdateComment)
				r.Delete("/", commentHandler.DeleteComment)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", commentHandler.VoteComment)
				})
			})
		})
		//accounts
		s.mux.Route("/accounts", func(r chi.Router) {
			r.Post("/", accountsHandler.CreateAccount)
			r.Get("/", accountsHandler.GetAccounts)
			r.Route("/avatar", func(r chi.Router) {
				r.Post("/", accountsHandler.UploadAccountAvatar)
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", accountsHandler.GetAccount)
				r.Put("/", accountsHandler.UpdateAccount)
				r.Delete("/", accountsHandler.DeleteAccount)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", accountsHandler.GetAccountByUsername)
			})
			r.Route("/validate", func(r chi.Router) {
				r.Get("/", accountsHandler.GetAccountValidation)
			})
		})
		s.mux.Route("/friends", func(r chi.Router) {
			r.Get("/", friendHandler.GetFriends)
			r.Post("/", friendHandler.CreateFriendRequest)
			r.Post("/accept", friendHandler.AcceptFriendRequest)
			r.Post("/block", friendHandler.BlockFriendRequest)
			r.Post("/unblock", friendHandler.UnblockFriendRequest)
			r.Post("/deny", friendHandler.DeleteFriend)
			r.Route("/{username}", func(r chi.Router) {
				r.Delete("/", friendHandler.DeleteFriend)
			})
			r.Route("/requests", func(r chi.Router) {
				r.Get("/sent", friendHandler.GetFriendRequestsSent)
				r.Get("/received", friendHandler.GetFriendRequestsReceived)
				r.Route("/{username}", func(r chi.Router) {
					r.Delete("/", friendHandler.DeleteFriendRequest)
				})
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", friendHandler.GetFriendsUsername)
			})
		})
		//problems
		s.mux.Route("/problems", func(r chi.Router) {
			r.Get("/", problemHandler.GetProblems)
			r.Route("/{problemId}", func(r chi.Router) {
				r.Get("/", problemHandler.GetProblem)
			})
		})
		//submissions
		s.mux.Route("/submissions", func(r chi.Router) {
			r.Route("/{token}", func(r chi.Router) {
				r.Get("/", submissionHandler.GetSubmission)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", submissionHandler.GetSubmissionsByUsername)
			})
			r.Post("/", submissionHandler.CreateSubmission)
		})
		//rooms
		/*r.Route("/rooms", func(r chi.Router) {
			r.Get("/", h.GetRooms)
			r.Post("/", h.CreateRoom)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetRoom)
			})
		})*/
		//runs
		s.mux.Route("/runs", func(r chi.Router) {
			r.Post("/", runHandler.CreateRun)
		})
		//starred
		s.mux.Route("/starred", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", starredHandler.GetStarredProblems)
				r.Put("/", starredHandler.PutStarProblem)
			})
			r.Route("/solutions", func(r chi.Router) {
				r.Get("/", starredHandler.GetStarredSolutions)
				r.Put("/", starredHandler.PutStarSolution)
			})
			r.Route("/submissions", func(r chi.Router) {
				r.Get("/", starredHandler.GetStarredSubmissions)
				r.Put("/", starredHandler.PutStarSubmission)
			})
		})
		s.mux.Route("/admin", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", adminHandler.GetAdminProblems)
				r.Post("/", adminHandler.CreateAdminProblem)
				r.Post("/run", adminHandler.CreateAdminProblemRun)
			})
			s.mux.Get("/validate", adminHandler.GetAdminValidation)
		})
	})
	//generate a route to catch anything not defined and error/block spam
	s.mux.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		SendError(w, http.StatusNotFound, "Not Found")
	})
	s.mux.Put("/*", func(w http.ResponseWriter, r *http.Request) {
		SendError(w, http.StatusNotFound, "Not Found")
	})
	s.mux.Post("/*", func(w http.ResponseWriter, r *http.Request) {
		SendError(w, http.StatusNotFound, "Not Found")
	})
	s.mux.Delete("/*", func(w http.ResponseWriter, r *http.Request) {
		SendError(w, http.StatusNotFound, "Not Found")
	})
}
