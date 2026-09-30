package response

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, resCode int, payload any) {
	resData, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("error encoding response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resCode)
	_, err = w.Write(resData)
	if err != nil {
		log.Printf("error writing response: %v", err)
	}
}
