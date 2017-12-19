package cms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"google.golang.org/appengine/log"
)

const responseKey = "result"

type Result map[string]interface{}

func printError(w http.ResponseWriter, err error, code int) {
	write(w, "", code, err.Error(), responseKey, nil)
}

func printData(w http.ResponseWriter, response interface{}) {
	write(w, "", http.StatusOK, "", responseKey, response)
}

func (ctx *Context) Print(w http.ResponseWriter, response interface{}) {
	write(w, ctx.Token(), http.StatusOK, "", responseKey, response)
}

func (ctx *Context) PrintError(w http.ResponseWriter, err error, code int) {
	log.Errorf(ctx.Context, "Internal Error: %v", err)
	write(w, ctx.Token(), code, err.Error(), responseKey, nil)
}

func write(w http.ResponseWriter, token string, status int, message string, responseKey string, response interface{}) {
	var out = Result{
		"status": status,
	}

	if len(token) != 0 {
		out["token"] = token
	}

	if len(message) > 0 {
		out["message"] = message
	}

	if response != nil {
		out[responseKey] = response
	}

	printOut(w, out)
}

func printOut(w http.ResponseWriter, out Result) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		fmt.Fprint(w, map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
	}
}
