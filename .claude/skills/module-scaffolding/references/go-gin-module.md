# Go/Gin Module Template

> This template shows a **multi-tenant** app. For single-company projects such as **TechShop**, drop `TenantID` and `mw.Tenant`, replace tenant filters with ownership filters (e.g. `WHERE customer_user_id = $2` or a seller-membership check), and write SQL in `backend/db/queries/` for sqlc instead of a hand-written repository. Follow `docs/backend.md` for module layout and error format.

Files under `internal/<feature>/`:

```go
// model.go
type Invoice struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Number      string
    AmountMinor int64
    Currency    string
    Status      Status
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

```go
// dto.go
type CreateInvoiceRequest struct {
    Number      string `json:"number" binding:"required,max=50"`
    AmountMinor int64  `json:"amount_minor" binding:"required,gt=0"`
    Currency    string `json:"currency" binding:"required,len=3"`
}
type InvoiceResponse struct { /* only fields the client needs */ }
type ListQuery struct {
    Page  int `form:"page,default=1" binding:"min=1"`
    Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}
```

```go
// repository.go
type Repository interface {
    Create(ctx context.Context, inv *Invoice) error
    GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Invoice, error)
    List(ctx context.Context, tenantID uuid.UUID, q ListQuery) ([]Invoice, int, error)
    Update(ctx context.Context, inv *Invoice) error
    SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error
}
// SQL always includes: WHERE tenant_id = $1 AND deleted_at IS NULL
```

```go
// service.go
type Service struct { repo Repository; audit audit.Logger }
func (s *Service) Create(ctx context.Context, actor auth.Actor, req CreateInvoiceRequest) (*Invoice, error) {
    if !actor.Can("invoices:write") { return nil, ErrForbidden }
    inv := &Invoice{ID: uuid.New(), TenantID: actor.TenantID, ...}
    if err := s.repo.Create(ctx, inv); err != nil { return nil, fmt.Errorf("create invoice: %w", err) }
    s.audit.Log(ctx, actor, "invoice.create", inv.ID)
    return inv, nil
}
```

```go
// handler.go + routes
func RegisterRoutes(r *gin.RouterGroup, h *Handler, mw middleware.Set) {
    g := r.Group("/invoices", mw.Auth, mw.Tenant)
    g.GET("", mw.Require("invoices:read"), h.List)
    g.GET("/:id", mw.Require("invoices:read"), h.Get)
    g.POST("", mw.Require("invoices:write"), h.Create)
    g.PATCH("/:id", mw.Require("invoices:write"), h.Update)
    g.DELETE("/:id", mw.Require("invoices:delete"), h.Delete)
}
```

Tests: `service_test.go` (table-driven, mock repo) and `handler_test.go` with `httptest` + a real test DB (testcontainers-go) for integration.
