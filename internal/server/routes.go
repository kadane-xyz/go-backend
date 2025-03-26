package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) RegisterApiRoutes() {
	s.mux.Route("/v1", func(r chi.Router) {
		// solutions
		s.mux.Route("/solutions", func(r chi.Router) {
			r.Get("/", s.container.APIHandlers.AccountHandler.GetSolutions)
			r.Post("/", s.container.APIHandlers.AccountHandler.CreateSolution)
			r.Route("/{solutionId}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.AccountHandler.GetSolution)
				r.Put("/", s.container.APIHandlers.AccountHandler.UpdateSolution)
				r.Delete("/", s.container.APIHandlers.AccountHandler.DeleteSolution)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", s.container.APIHandlers.AccountHandler.VoteSolution)
				})
			})
		})
		// comments
		s.mux.Route("/comments", func(r chi.Router) {
			r.Get("/", s.container.APIHandlers.GetComments)
			r.Post("/", s.container.APIHandlers.CreateComment)
			r.Route("/{commentId}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetComment)
				r.Put("/", s.container.APIHandlers.UpdateComment)
				r.Delete("/", s.container.APIHandlers.DeleteComment)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", s.container.APIHandlers.VoteComment)
				})
			})
		})
		//accounts
		s.mux.Route("/accounts", func(r chi.Router) {
			r.Post("/", s.container.APIHandlers.AccountHandler.CreateAccount)
			r.Get("/", s.container.APIHandlers.AccountHandler.GetAccounts)
			r.Route("/avatar", func(r chi.Router) {
				r.Post("/", s.container.APIHandlers.AccountHandler.UploadAvatar)
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.AccountHandler.GetAccount)
				r.Put("/", s.container.APIHandlers.AccountHandler.UpdateAccount)
				r.Delete("/", s.container.APIHandlers.AccountHandler.DeleteAccount)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.AccountHandler.GetAccountByUsername)
			})
			r.Route("/validate", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.AccountHandler.GetAccountValidation)
			})
		})
		s.mux.Route("/friends", func(r chi.Router) {
			r.Get("/", s.container.APIHandlers.GetFriends)
			r.Post("/", s.container.APIHandlers.CreateFriendRequest)
			r.Post("/accept", s.container.APIHandlers.AcceptFriendRequest)
			r.Post("/block", s.container.APIHandlers.BlockFriendRequest)
			r.Post("/unblock", s.container.APIHandlers.UnblockFriendRequest)
			r.Post("/deny", s.container.APIHandlers.DeleteFriend)
			r.Route("/{username}", func(r chi.Router) {
				r.Delete("/", s.container.APIHandlers.DeleteFriend)
			})
			r.Route("/requests", func(r chi.Router) {
				r.Get("/sent", s.container.APIHandlers.GetFriendRequestsSent)
				r.Get("/received", s.container.APIHandlers.GetFriendRequestsReceived)
				r.Route("/{username}", func(r chi.Router) {
					r.Delete("/", s.container.APIHandlers.DeleteFriendRequest)
				})
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetFriendsUsername)
			})
		})
		//problems
		s.mux.Route("/problems", func(r chi.Router) {
			r.Get("/", s.container.APIHandlers.GetProblemsRoute)
			r.Route("/{problemId}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetProblem)
			})
		})
		//submissions
		s.mux.Route("/submissions", func(r chi.Router) {
			r.Route("/{token}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetSubmission)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetSubmissionsByUsername)
			})
			r.Post("/", s.container.APIHandlers.CreateSubmissionRoute)
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
			r.Post("/", s.container.APIHandlers.CreateRunRoute)
		})
		//starred
		s.mux.Route("/starred", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetStarredProblems)
				r.Put("/", s.container.APIHandlers.PutStarProblem)
			})
			r.Route("/solutions", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetStarredSolutions)
				r.Put("/", s.container.APIHandlers.PutStarSolution)
			})
			r.Route("/submissions", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetStarredSubmissions)
				r.Put("/", s.container.APIHandlers.PutStarSubmission)
			})
		})
		s.mux.Route("/admin", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", s.container.APIHandlers.GetAdminProblems)
				r.Post("/", s.container.APIHandlers.CreateAdminProblem)
				r.Post("/run", s.container.APIHandlers.CreateAdminProblemRun)
			})
			s.mux.Get("/validate", s.container.APIHandlers.GetAdminValidation)
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
