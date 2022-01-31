package controllers

import "net/http"

// Template interface for execution
type Template interface {
	Execute(w http.ResponseWriter, data interface{})
}
