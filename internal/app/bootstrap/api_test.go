package bootstrap

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestBuildRouterRegistersCoreRoutes(t *testing.T) {
	t.Parallel()

	r := BuildRouter(RouterConfig{
		JWTSecret:      "secret",
		SwaggerDocPath: "./docs/swagger.json",
		Observability:  RouterConfig{}.Observability,
	}, RouteRegistrars{
		UserNoAuth: func(r chi.Router) { r.Post("/auth/register", noop) },
		UserAuth:   func(r chi.Router) { r.Get("/users/{id}", noop) },
		Genre:      func(r chi.Router) { r.Get("/genres/", noop) },
		Movie:      func(r chi.Router) { r.Get("/movies/", noop) },
	})

	routes := collectRoutes(t, r)

	assertHasRoute(t, routes, "GET", "/metrics/*")
	assertHasRoute(t, routes, "GET", "/api/v1/swagger/*")
	assertHasRoute(t, routes, "GET", "/api/v1/swagger/doc.json")
	assertHasRoute(t, routes, "POST", "/api/v1/auth/register")
	assertHasRoute(t, routes, "GET", "/api/v1/genres/")
	assertHasRoute(t, routes, "GET", "/api/v1/movies/")
	assertHasRoute(t, routes, "GET", "/api/v1/users/{id}")
}

func TestBuildRouterKeepsAuthRoutesUnderApiV1(t *testing.T) {
	t.Parallel()

	r := BuildRouter(RouterConfig{JWTSecret: "secret"}, RouteRegistrars{
		UserAuth: func(r chi.Router) { r.Patch("/users/password", noop) },
	})

	routes := collectRoutes(t, r)
	assertHasRoute(t, routes, "PATCH", "/api/v1/users/password")
}

func collectRoutes(t *testing.T, r *chi.Mux) map[string][]string {
	t.Helper()

	out := map[string][]string{}
	if err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		out[method] = append(out[method], route)
		return nil
	}); err != nil {
		t.Fatalf("walk routes: %v", err)
	}
	return out
}

func assertHasRoute(t *testing.T, routes map[string][]string, method, route string) {
	t.Helper()
	if !contains(routes[method], route) {
		t.Fatalf("route %s %s not found; got %v", method, route, routes[method])
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func noop(_ http.ResponseWriter, _ *http.Request) {}
