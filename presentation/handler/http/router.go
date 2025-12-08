// Package http provides HTTP controllers and routing for web endpoints.
//
// Routes are defined in routes.go following a pattern similar to Laravel's web.php/api.php.
// Each route maps an HTTP method and path to a controller handler.
package http

import (
	"net/http"
)

// Route defines a single HTTP endpoint.
type Route struct {
	Method      string           // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path        string           // URL path pattern
	Handler     http.HandlerFunc // Handler function for this route
	Description string           // Human-readable description for documentation
}

// Router manages HTTP routes and dispatches requests to handlers.
type Router struct {
	routes []Route
	mux    *http.ServeMux
}

// NewRouter creates a new HTTP router.
func NewRouter() *Router {
	return &Router{
		routes: make([]Route, 0),
		mux:    http.NewServeMux(),
	}
}

// Register adds a new route to the router.
func (r *Router) Register(method, path, description string, handler http.HandlerFunc) {
	route := Route{
		Method:      method,
		Path:        path,
		Handler:     handler,
		Description: description,
	}
	r.routes = append(r.routes, route)
}

// GET registers a GET route.
func (r *Router) GET(path, description string, handler http.HandlerFunc) {
	r.Register(http.MethodGet, path, description, handler)
}

// POST registers a POST route.
func (r *Router) POST(path, description string, handler http.HandlerFunc) {
	r.Register(http.MethodPost, path, description, handler)
}

// PUT registers a PUT route.
func (r *Router) PUT(path, description string, handler http.HandlerFunc) {
	r.Register(http.MethodPut, path, description, handler)
}

// DELETE registers a DELETE route.
func (r *Router) DELETE(path, description string, handler http.HandlerFunc) {
	r.Register(http.MethodDelete, path, description, handler)
}

// Build compiles all routes into an http.Handler.
func (r *Router) Build() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		for _, route := range r.routes {
			if req.URL.Path == route.Path && req.Method == route.Method {
				route.Handler(w, req)
				return
			}
		}
		http.NotFound(w, req)
	})
}

// Routes returns all registered routes (useful for documentation/debugging).
func (r *Router) Routes() []Route {
	return r.routes
}

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Build().ServeHTTP(w, req)
}
