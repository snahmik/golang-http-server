package response

import (
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func InternalServerError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}

	JSON(w, http.StatusInternalServerError, errorResponse{
		Error: "Internal Server Error",
	})
}

func UnauthorizedError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("auth error: %v", err)
	}

	JSON(w, http.StatusUnauthorized, errorResponse{
		Error: "Unauthorized",
	})
}

func BadRequestError(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		log.Printf("request error: %v", err)
	}

	errMsg := msg
	if errMsg == ""{
		errMsg = "Invalid Request"
	}

	JSON(w, http.StatusBadRequest, errorResponse{
		Error: errMsg,
	})
}

func GenericError(w http.ResponseWriter, resCode int, msg string, err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}

	if resCode > 499 {
		log.Printf("Responding with 5xx error: %s", msg)
	}

	JSON(w, resCode, errorResponse{
		Error: msg,
	})
}
