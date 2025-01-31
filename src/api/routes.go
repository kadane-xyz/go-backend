package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/apierror"
)

func RegisterApiRoutes(h *Handler, r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		// solutions
		r.Route("/solutions", func(r chi.Router) {
			r.Get("/", h.GetSolutions)
			r.Post("/", h.CreateSolution)
			r.Route("/{solutionId}", func(r chi.Router) {
				r.Get("/", h.GetSolution)
				r.Put("/", h.UpdateSolution)
				r.Delete("/", h.DeleteSolution)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", h.VoteSolution)
				})
			})
		})
		// comments
		r.Route("/comments", func(r chi.Router) {
			r.Get("/", h.GetComments)
			r.Post("/", h.CreateComment)
			r.Route("/{commentId}", func(r chi.Router) {
				r.Get("/", h.GetComment)
				r.Put("/", h.UpdateComment)
				r.Delete("/", h.DeleteComment)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", h.VoteComment)
				})
			})
		})
		//accounts
		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", h.CreateAccount)
			r.Get("/", h.GetAccounts)
			r.Route("/avatar", func(r chi.Router) {
				r.Post("/", h.UploadAvatar)
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetAccount)
				r.Put("/", h.UpdateAccount)
				r.Delete("/", h.DeleteAccount)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", h.GetAccountByUsername)
			})
		})
		r.Route("/friends", func(r chi.Router) {
			r.Get("/", h.GetFriends)
			r.Post("/", h.CreateFriendRequest)
			r.Post("/accept", h.AcceptFriendRequest)
			r.Post("/block", h.BlockFriendRequest)
			r.Post("/unblock", h.UnblockFriendRequest)
			r.Post("/deny", h.DeleteFriend)
			r.Route("/{username}", func(r chi.Router) {
				r.Delete("/", h.DeleteFriend)
			})
			r.Route("/requests", func(r chi.Router) {
				r.Get("/sent", h.GetFriendRequestsSent)
				r.Get("/received", h.GetFriendRequestsReceived)
				r.Route("/{username}", func(r chi.Router) {
					r.Delete("/", h.DeleteFriendRequest)
				})
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", h.GetFriendsUsername)
			})
		})
		//problems
		r.Route("/problems", func(r chi.Router) {
			r.Get("/", h.GetProblems)
			r.Post("/", h.CreateProblem)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetProblem)
			})
		})
		//submissions
		r.Route("/submissions", func(r chi.Router) {
			r.Route("/{token}", func(r chi.Router) {
				r.Get("/", h.GetSubmission)
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", h.GetSubmissionsByUsername)
			})
			r.Post("/", h.CreateSubmission)
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
			r.Post("/", h.CreateRun)
		})
		//starred
		r.Route("/starred", func(r chi.Router) {
			r.Route("/problems", func(r chi.Router) {
				r.Get("/", h.GetStarredProblems)
				r.Put("/", h.PutStarProblem)
			})
			r.Route("/solutions", func(r chi.Router) {
				r.Get("/", h.GetStarredSolutions)
				r.Put("/", h.PutStarSolution)
			})
			r.Route("/submissions", func(r chi.Router) {
				r.Get("/", h.GetStarredSubmissions)
				r.Put("/", h.PutStarSubmission)
			})
		})
	})
	//generate a route to catch anything not defined and error/block spam
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		apierror.SendError(w, http.StatusNotFound, "Not Found")
	})
	r.Put("/*", func(w http.ResponseWriter, r *http.Request) {
		apierror.SendError(w, http.StatusNotFound, "Not Found")
	})
	r.Post("/*", func(w http.ResponseWriter, r *http.Request) {
		apierror.SendError(w, http.StatusNotFound, "Not Found")
	})
	r.Delete("/*", func(w http.ResponseWriter, r *http.Request) {
		apierror.SendError(w, http.StatusNotFound, "Not Found")
	})
}
