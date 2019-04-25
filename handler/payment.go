package handler

import (
	"encoding/json"
	"github.com/clD11/form3-payments/model"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

// GET /v1/payments/{id}
func GetPayment(db *pg.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid, err := uuid.FromString(vars["id"])
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	payment := model.Payment{ID: uuid}
	if err := db.Select(&payment); err != nil {
		if err == pg.ErrNoRows {
			writeErrorResponse(w, http.StatusNotFound, "Payment not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Server failed to return payment")
		return
	}

	writeResponse(w, http.StatusOK, &payment)
}

// POST /v1/payments
func CreatePayment(db *pg.DB, w http.ResponseWriter, r *http.Request) {
	var payment model.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Could not decode request body")
		return
	}
	defer r.Body.Close()

	if err := db.Select(&payment); err != pg.ErrNoRows {
		writeErrorResponse(w, http.StatusBadRequest, "Cannot create payment already exists")
		return
	}

	if err := db.Insert(&payment); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not insert payment")
		return
	}

	writeResponse(w, http.StatusCreated, payment)
}

// DELETE "/v1/payments/{id}"
func DeletePayment(db *pg.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid, err := uuid.FromString(vars["id"])
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	payment := model.Payment{ID: uuid}
	if err := db.Select(&payment); err != nil {
		writeErrorResponse(w, http.StatusNotFound, "Payment not found cannot delete")
		return
	}

	if err := db.Delete(&payment); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Payment could not be deleted")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PUT /v1/payments/{id}
func UpdatePayment(db *pg.DB, w http.ResponseWriter, r *http.Request) {
	// get variable
	vars := mux.Vars(r)

	uuid, err := uuid.FromString(vars["id"])
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// decode body
	requestPayment := model.Payment{}
	if err := json.NewDecoder(r.Body).Decode(&requestPayment); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Could not decode request body")
		return
	}
	defer r.Body.Close()

	// validate request
	if uuid != requestPayment.ID {
		writeErrorResponse(w, http.StatusBadRequest, "Could not update payment - request id does not match update payment")
	}

	// check payment exists
	currentPayment := model.Payment{ID: uuid}
	if err := db.Select(&currentPayment); err != nil {
		writeErrorResponse(w, http.StatusNotFound, "Could not update payment as not found")
		return
	}

	// update record
	if err := db.Update(&requestPayment); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not update payment")
		return
	}

	// return status
	w.WriteHeader(http.StatusCreated)
}

// GET /v1/payments
func GetPayments(db *pg.DB, w http.ResponseWriter, r *http.Request) {
	payments := []model.Payment{}

	if err := db.Model(&payments).Select(); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not get all payments")
	}

	writeResponse(w, http.StatusOK, &payments)
}

// Could be moved to handler utils for use with other handlers
func writeResponse(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	writeResponse(w, code, map[string]string{"error": message})
}
