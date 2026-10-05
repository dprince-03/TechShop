package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

// Op describes an endpoint. Auth lists accepted audiences (empty = public); Perm is a staff
// permission; StepUp forces a recent MFA; Optional attaches a principal if present.
type Op struct {
	ID       string
	Method   string
	Path     string
	Summary  string
	Tag      string
	Auth     []string
	Optional bool
	Perm     string
	Scope    string
	StepUp   bool
	Status   int
	Hidden   bool
	MaxBody  int64
}

// Register adds an endpoint with its guards. Every staff route must declare a permission
// (enforced by a test over the generated spec).
func Register[I, O any](g *Guard, op Op, handler func(context.Context, *I) (*O, error)) {
	h := huma.Operation{
		OperationID:   op.ID,
		Method:        op.Method,
		Path:          op.Path,
		Summary:       op.Summary,
		Tags:          []string{op.Tag},
		DefaultStatus: op.Status,
		Hidden:        op.Hidden,
		MaxBodyBytes:  op.MaxBody,
		Metadata:      map[string]any{},
	}
	if op.Perm != "" {
		h.Metadata["permission"] = op.Perm
		h.Description = "Requires permission `" + op.Perm + "`."
	}
	switch {
	case len(op.Auth) > 0 && op.Optional:
		h.Middlewares = append(h.Middlewares, g.Optional(op.Auth...))
	case len(op.Auth) > 0:
		h.Security = []map[string][]string{{"bearer": {}}}
		h.Middlewares = append(h.Middlewares, g.Authenticate(op.Auth...))
	}
	if op.Perm != "" {
		h.Middlewares = append(h.Middlewares, g.Require(op.Perm))
	}
	if op.Scope != "" {
		h.Middlewares = append(h.Middlewares, g.RequireScope(op.Scope))
	}
	if op.StepUp {
		h.Middlewares = append(h.Middlewares, g.RequireStepUp())
	}
	if h.DefaultStatus == 0 {
		if op.Method == http.MethodPost {
			h.DefaultStatus = http.StatusOK
		}
	}
	huma.Register(g.API, h, handler)
}
