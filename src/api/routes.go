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
			})
			r.Route("/username/{username}", func(r chi.Router) {
				r.Get("/", h.GetAccountByUsername)
			})
		})
		//submissions
		r.Route("/submissions", func(r chi.Router) {
			r.Get("/{token}", h.GetSubmission)
			r.Post("/", h.CreateSubmission)
		})
	})
	//generate a route to catch anything not defined and error/block spam
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		apierror.SendError(w, http.StatusNotFound, "Not Found")
	})
}
