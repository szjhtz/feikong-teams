package router

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterAPIRoutesIncludesCoreEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	registerAPIRoutes(engine, false)

	routes := routeSet(engine.Routes())
	for _, route := range []string{
		"GET /health",
		"GET /ws",
		"GET /v1/models",
		"POST /v1/chat/completions",
		"GET /api/fkteams/version",
		"GET /api/fkteams/agents",
		"POST /api/fkteams/chat",
		"POST /api/fkteams/stream/start",
		"PATCH /api/fkteams/stream/queue/:sessionID/:queueID",
		"DELETE /api/fkteams/stream/queue/:sessionID/:queueID",
		"GET /api/fkteams/files/serve/*filepath",
		"GET /api/fkteams/preview/:linkId/render/*filepath",
		"POST /api/fkteams/session-shares",
		"GET /api/fkteams/public/session-shares/:shareID/info",
		"GET /api/fkteams/sessions/:sessionID",
		"GET /api/fkteams/schedules/:id/history/:filename",
		"GET /api/fkteams/skills/:slug/file",
		"POST /api/fkteams/memory/clear",
		"GET /api/fkteams/config/template-vars",
		"POST /api/fkteams/providers/models",
		"POST /api/fkteams/shutdown",
		"POST /api/fkteams/restart",
	} {
		if !routes[route] {
			t.Fatalf("route %s was not registered", route)
		}
	}

	if routes["POST /api/fkteams/login"] {
		t.Fatal("login route should not be registered when auth is disabled")
	}
}

func TestRegisterAPIRoutesAddsLoginWhenAuthEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	registerAPIRoutes(engine, true)

	routes := routeSet(engine.Routes())
	if !routes["POST /api/fkteams/login"] {
		t.Fatal("login route should be registered when auth is enabled")
	}
}

func TestNewEngineAddsMiddlewareAndRoutesCanBeRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newEngine(false)
	registerAPIRoutes(engine, false)

	if len(engine.Routes()) == 0 {
		t.Fatal("engine should have registered routes")
	}
}

func routeSet(routes gin.RoutesInfo) map[string]bool {
	result := make(map[string]bool, len(routes))
	for _, route := range routes {
		result[route.Method+" "+route.Path] = true
	}
	return result
}
