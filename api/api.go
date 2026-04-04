// Package api provides a self-documenting API route registry.
package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Route represents an API endpoint with documentation.
type Route struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Desc    string            `json:"description"`
	Group   string            `json:"group"`
	Handler httprouter.Handle `json:"-"`
}

// Registry holds all registered routes.
type Registry struct {
	routes []Route
}

// NewRegistry creates a new route registry.
func NewRegistry() *Registry {
	return &Registry{routes: []Route{}}
}

// Add registers a route with documentation.
func (r *Registry) Add(method, path, group, desc string, handler httprouter.Handle) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Path:    path,
		Desc:    desc,
		Group:   group,
		Handler: handler,
	})
}

// GET registers a GET route.
func (r *Registry) GET(path, group, desc string, handler httprouter.Handle) {
	r.Add("GET", path, group, desc, handler)
}

// POST registers a POST route.
func (r *Registry) POST(path, group, desc string, handler httprouter.Handle) {
	r.Add("POST", path, group, desc, handler)
}

// DELETE registers a DELETE route.
func (r *Registry) DELETE(path, group, desc string, handler httprouter.Handle) {
	r.Add("DELETE", path, group, desc, handler)
}

// Register applies all routes to an httprouter.
func (r *Registry) Register(router *httprouter.Router) {
	for _, route := range r.routes {
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Path, route.Handler)
		case http.MethodPost:
			router.POST(route.Path, route.Handler)
		case http.MethodDelete:
			router.DELETE(route.Path, route.Handler)
		case http.MethodPut:
			router.PUT(route.Path, route.Handler)
		case http.MethodPatch:
			router.PATCH(route.Path, route.Handler)
		}
	}
}

// Routes returns all registered routes (for documentation).
func (r *Registry) Routes() []Route {
	return r.routes
}

// GroupedRoutes returns routes grouped by category.
func (r *Registry) GroupedRoutes() map[string][]Route {
	grouped := make(map[string][]Route)
	for _, route := range r.routes {
		grouped[route.Group] = append(grouped[route.Group], route)
	}
	return grouped
}

// DocsHandler returns an HTTP handler that serves API documentation.
func (r *Registry) DocsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		accept := req.Header.Get("Accept")

		// Return JSON if requested
		if strings.Contains(accept, "application/json") {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(r.routes); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Return HTML documentation
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(r.generateHTML()))
	}
}

func (r *Registry) generateHTML() string {
	grouped := r.GroupedRoutes()

	// Sort groups
	groups := make([]string, 0, len(grouped))
	for g := range grouped {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
<title>API Documentation</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
.container { max-width: 1000px; margin: 0 auto; }
h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
h2 { color: #555; margin-top: 30px; }
.route { background: white; border-radius: 8px; padding: 15px; margin: 10px 0; box-shadow: 0 1px 3px rgba(0,0,0,0.1); display: flex; align-items: center; gap: 15px; }
.method { font-weight: bold; padding: 4px 10px; border-radius: 4px; font-size: 12px; min-width: 60px; text-align: center; }
.GET { background: #61affe; color: white; }
.POST { background: #49cc90; color: white; }
.DELETE { background: #f93e3e; color: white; }
.PUT { background: #fca130; color: white; }
.PATCH { background: #50e3c2; color: white; }
.path { font-family: monospace; color: #333; flex: 1; }
.desc { color: #666; font-size: 14px; }
.param { color: #007bff; }
a { color: #007bff; text-decoration: none; }
a:hover { text-decoration: underline; }
.header { display: flex; justify-content: space-between; align-items: center; }
.json-link { font-size: 14px; }
.auth-section { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.auth-section h3 { margin-top: 0; color: #333; }
.auth-section code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; font-size: 13px; }
.auth-section pre { background: #2d2d2d; color: #f8f8f2; padding: 15px; border-radius: 6px; overflow-x: auto; font-size: 13px; }
.auth-section pre .comment { color: #6272a4; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>API Documentation</h1>
<a class="json-link" href="/api/v1?format=json" onclick="fetchJSON(event)">View as JSON</a>
</div>
<p>Base URL: <code>/api/v1</code></p>

<div class="auth-section">
<h3>Authentication</h3>
<p>All API endpoints require authentication. Two methods are supported:</p>

<p><strong>1. Session Cookie</strong> (browser usage)<br>
Login via POST to <code>/</code> with <code>email</code> and <code>pass</code> form fields. A session cookie is set automatically.</p>

<p><strong>2. API Key</strong> (programmatic access)<br>
Add your API key to the <code>X-API-Key</code> header. Generate a key with <code>contrib/gen</code> and add it to your user in <code>entities.toml</code>.</p>

<pre><span class="comment"># List entities</span>
curl -H "X-API-Key: YOUR_API_KEY" http://localhost:9999/api/v1/entities

<span class="comment"># Get invoices for a company/year</span>
curl -H "X-API-Key: YOUR_API_KEY" http://localhost:9999/api/v1/invoices/mycompany/2024

<span class="comment"># Create an invoice (POST)</span>
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"customer":"Acme Corp",...}' \
     http://localhost:9999/api/v1/invoice/mycompany/2024</pre>
</div>
`)

	for _, group := range groups {
		routes := grouped[group]
		sb.WriteString("<h2>" + group + "</h2>\n")

		for _, route := range routes {
			// Highlight path parameters
			path := route.Path
			path = strings.ReplaceAll(path, ":", `<span class="param">:`)
			path = strings.ReplaceAll(path, "/", `</span>/`)
			path = strings.TrimSuffix(path, `</span>/`)
			// Fix any remaining unclosed spans
			paramCount := strings.Count(path, `<span class="param">`)
			closeCount := strings.Count(path, `</span>`)
			for i := closeCount; i < paramCount; i++ {
				path += "</span>"
			}

			sb.WriteString(`<div class="route">`)
			sb.WriteString(`<span class="method ` + route.Method + `">` + route.Method + `</span>`)
			sb.WriteString(`<span class="path">` + path + `</span>`)
			sb.WriteString(`<span class="desc">` + route.Desc + `</span>`)
			sb.WriteString("</div>\n")
		}
	}

	sb.WriteString(`
<script>
function fetchJSON(e) {
	e.preventDefault();
	fetch('/api/v1', {headers: {'Accept': 'application/json'}})
		.then(r => r.json())
		.then(d => {
			document.body.innerHTML = '<pre style="padding:20px">' + JSON.stringify(d, null, 2) + '</pre>';
		});
}
</script>
</div>
</body>
</html>`)

	return sb.String()
}
