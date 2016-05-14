package http

import (
	"net/http"
)

type Route interface {
	HandleFunc(w http.ResponseWriter, req *http.Request)
}

type Protocol interface {
	RouteLoad(*http.ServeMux)
}
