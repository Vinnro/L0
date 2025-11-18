package http

import (
	"L0_main/internal/domain"
	"encoding/json"
	"io"
	"net/http"

	"L0_main/internal/usecase"

	"github.com/gorilla/mux"
)

func CreateOrderHandler(s *usecase.OrdService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)

		file, handler, err := r.FormFile("order_file")
		if err != nil {
			http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
			return
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Ошибка чтения файла", http.StatusBadRequest)
			return
		}

		var ord domain.Order
		if err := json.Unmarshal(bytes, &ord); err != nil {
			http.Error(w, "Неправильный JSON", http.StatusBadRequest)
			return
		}

		if err := s.InsertOrd(&ord); err != nil {
			http.Error(w, "Ошибка сервиса: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"status":    "success",
			"filename":  handler.Filename,
			"order_uid": ord.OrderUID,
		})
	}
}

func GetOrderHandler(s *usecase.OrdService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]

		order, err := s.GetOrd(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
