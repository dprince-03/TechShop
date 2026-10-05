package rbac

import (
	"context"

	"github.com/dprince-03/techshop/backend/internal/store"
)

// Sync writes the permission catalogue and the system "admin" role (every permission) to the
// database. It runs at API start-up so code and data never drift; it is idempotent.
func Sync(ctx context.Context, q *store.Queries) error {
	for _, p := range Catalogue {
		desc := p.Description
		if err := q.IdentityUpsertPermission(ctx, store.IdentityUpsertPermissionParams{Key: p.Key, Module: p.Module, Description: &desc, IsSensitive: p.Sensitive}); err != nil {
			return err
		}
	}
	desc := "Break-glass administrator: every permission. Grant sparingly."
	role, err := q.IdentityUpsertRole(ctx, store.IdentityUpsertRoleParams{Key: "admin", Name: "Administrator", Description: &desc, IsSystem: true})
	if err != nil {
		return err
	}
	return q.IdentityAddRolePermissions(ctx, store.IdentityAddRolePermissionsParams{RoleID: role.ID, Keys: Keys()})
}
