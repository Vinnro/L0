package http

import (
	"encoding/json"
	"net/http"

	"L0_main/internal/usecase"

	"github.com/gorilla/mux"
)

func GetOrderHandler(s *usecase.OrdService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		order, err := s.GetOrd(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(order); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
