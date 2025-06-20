package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/internal/api/handlers"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

func (s *Server) RegisterApiRoutes() {

	errorMiddleware := func(h middleware.HandlerFunc) http.HandlerFunc {
		return middleware.ErrorMiddleware(h).ServeHTTP
	}

	s.mux.Route("/v1", func(r chi.Router) {
		accountsHandler := handlers.NewAccountHandler(s.container.Repositories.AccountRepo, s.container.AWSClient, s.container.Config)
		commentHandler := handlers.NewCommentHandler(s.container.Repositories.CommentRepo, s.container.Repositories.SolutionRepo)
		problemHandler := handlers.NewProblemHandler(s.container.Repositories.ProblemRepo)
		adminHandler := handlers.NewAdminHandler(s.container.Repositories.AdminRepo, s.container.Repositories.ProblemRepo, s.container.Judge0)
		submissionHandler := handlers.NewSubmissionHandler(s.container.Repositories.SubmissionRepo, s.container.Repositories.ProblemRepo, s.container.Judge0)
		solutionsHandler := handlers.NewSolutionsHandler(s.container.Repositories.SolutionRepo)
		starredHandler := handlers.NewStarredHandler(s.container.Repositories.StarredRepo)
		runHandler := handlers.NewRunHandler(s.container.Repositories.ProblemRepo, s.container.Judge0)
		friendHandler := handlers.NewFriendHandler(s.container.Repositories.FriendRepo, s.container.Repositories.AccountRepo)

		// solutions
		r.Route("/solutions", func(r chi.Router) {
			r.Get("/", errorMiddleware(solutionsHandler.GetSolutions))
			r.Post("/", errorMiddleware(solutionsHandler.CreateSolution))
			r.Route("/{solutionId}", func(r chi.Router) {
				r.Get("/", errorMiddleware(solutionsHandler.GetSolution))
				r.Put("/", errorMiddleware(solutionsHandler.UpdateSolution))
				r.Delete("/", errorMiddleware(solutionsHandler.DeleteSolution))
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", errorMiddleware(solutionsHandler.VoteSolution))
				})
			})
		})
		// comments
		r.Route("/comments", func(r chi.Router) {
			r.Get("/", errorMiddleware(commentHandler.GetComments))
			r.Post("/", errorMiddleware(commentHandler.CreateComment))
			r.Route("/{commentId}", func(r chi.Router) {
				r.Get("/", errorMiddleware(commentHandler.GetComment))
				r.Put("/", errorMiddleware(commentHandler.UpdateComment))
				r.Delete("/", errorMiddleware(commentHandler.DeleteComment))
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", errorMiddleware(commentHandler.VoteComment))
				})
			})
		})
		//accounts
		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", errorMiddleware(accountsHandler.CreateAccount))
			r.Get("/", errorMiddleware(accountsHandler.GetAccounts))
			r.Route("/avatar", func(r chi.Router) {
				r.Post("/", errorMiddleware(accountsHandler.UploadAccountAvatar))
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", errorMiddleware(accountsHandler.GetAccount))
				r.Put("/", errorMiddleware(accountsHandler.UpdateAccount))
				r.Delete("/", errorMiddleware(accountsHandler.DeleteAccount))
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", errorMiddleware(accountsHandler.GetAccountByUsername))
			})
			r.Route("/validate", func(r chi.Router) {
				r.Get("/", errorMiddleware(accountsHandler.GetAccountValidation))
			})
		})
		r.Route("/friends", func(r chi.Router) {
			r.Get("/", errorMiddleware(friendHandler.GetFriends))
			r.Post("/", errorMiddleware(friendHandler.CreateFriendRequest))
			r.Post("/accept", errorMiddleware(friendHandler.AcceptFriendRequest))
			r.Post("/block", errorMiddleware(friendHandler.BlockFriendRequest))
			r.Post("/unblock", errorMiddleware(friendHandler.UnblockFriendRequest))
			r.Post("/deny", errorMiddleware(friendHandler.DeleteFriend))
			r.Route("/{username}", func(r chi.Router) {
				r.Delete("/", errorMiddleware(friendHandler.DeleteFriend))
			})
			r.Route("/requests", func(r chi.Router) {
				r.Get("/sent", errorMiddleware(friendHandler.GetFriendRequestsSent))
				r.Get("/received", errorMiddleware(friendHandler.GetFriendRequestsReceived))
				r.Route("/{username}", func(r chi.Router) {
					r.Delete("/", errorMiddleware(friendHandler.DeleteFriendRequest))
				})
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", errorMiddleware(friendHandler.GetFriendsUsername))
			})
		})
		//problems
		r.Route("/problems", func(r chi.Router) {
			r.Get("/", errorMiddleware(problemHandler.GetProblems))
			r.Route("/{problemId}", func(r chi.Router) {
				r.Get("/", errorMiddleware(problemHandler.GetProblem))
			})
		})
		//submissions
		r.Route("/submissions", func(r chi.Router) {
			r.Route("/{token}", func(r chi.Router) {
				r.Get("/", errorMiddleware(submissionHandler.GetSubmission))
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", errorMiddleware(submissionHandler.GetSubmissionsByUsername))
			})
			r.Post("/", errorMiddleware(submissionHandler.CreateSubmission))
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
		r.Route("/runs", func(r chi.Router) {
			r.Post("/", errorMiddleware(runHandler.CreateRun))
		})
		//starred
		r.Route("/starred", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", errorMiddleware(starredHandler.GetStarredProblems))
				r.Put("/", errorMiddleware(starredHandler.PutStarProblem))
			})
			r.Route("/solutions", func(r chi.Router) {
				r.Get("/", errorMiddleware(starredHandler.GetStarredSolutions))
				r.Put("/", errorMiddleware(starredHandler.PutStarSolution))
			})
			r.Route("/submissions", func(r chi.Router) {
				r.Get("/", errorMiddleware(starredHandler.GetStarredSubmissions))
				r.Put("/", errorMiddleware(starredHandler.PutStarSubmission))
			})
		})
		r.Route("/admin", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", errorMiddleware(adminHandler.GetAdminProblems))
				r.Post("/", errorMiddleware(adminHandler.CreateAdminProblem))
				r.Post("/run", errorMiddleware(adminHandler.CreateAdminProblemRun))
			})
			r.Get("/validate", errorMiddleware(adminHandler.GetAdminValidation))
		})
		// catch anything else under /v1
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error":"route not found"}`))
		})
	})
	s.mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"route not found"}`))
	})
	s.mux.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"route not found"}`))
	})
}
