package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterApiRoutes(h *Handler, r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome to the API"))
		})
		// solutions
		r.Route("/solutions", func(r chi.Router) {
			r.Get("/", h.GetSolutions)
			r.Post("/", h.CreateSolution)
			r.Route("/{solutionID}", func(r chi.Router) {
				r.Get("/", h.GetSolution)
				r.Put("/", h.UpdateSolution)
				r.Delete("/", h.DeleteSolution)
				r.Route("/vote", func(r chi.Router) {
					r.Patch("/", h.VoteSolution)
				})
			})
		})
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the API"))
	})
	//generate a route to catch anything not defined and error/block spam
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})
}
