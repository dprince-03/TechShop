// Package seed creates development data (accounts per role, sample catalogue, warehouses,
// zones, sellers). It is only called by cmd/seed, which refuses to run outside development.
package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/rbac"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/pkg/crypto"
)

// Account is a seeded login.
type Account struct {
	Role       string    `json:"role"`
	UserID     uuid.UUID `json:"userId"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Password   string    `json:"password"`
	TOTPSecret string    `json:"totpSecret,omitempty"`
	Audience   string    `json:"audience"`
}

// Output is printed by cmd/seed.
type Output struct {
	Note     string            `json:"note"`
	Accounts []Account         `json:"accounts"`
	Extra    map[string]string `json:"extra"`
}

// Run seeds everything idempotently.
func Run(ctx context.Context, d *kit.Deps, password string) (*Output, error) {
	if password == "" {
		password = "dev-" + crypto.RandomToken(12)
	}
	if err := rbac.Sync(ctx, d.Q); err != nil {
		return nil, err
	}
	out := &Output{Note: "Development accounts only. Never use in staging or production.", Extra: map[string]string{}}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	q := d.Q

	user := func(email, phone, first, last string) (store.IdentityUser, error) {
		if u, err := q.IdentityGetUserByEmail(ctx, &email); err == nil {
			_ = q.IdentitySetPassword(ctx, store.IdentitySetPasswordParams{ID: u.ID, PasswordHash: &hash})
			return u, nil
		}
		now := time.Now()
		var ph *string
		if phone != "" {
			ph = &phone
		}
		return q.IdentityCreateUser(ctx, store.IdentityCreateUserParams{Email: &email, Phone: ph, PasswordHash: &hash, FirstName: first, LastName: last, PhoneVerifiedAt: &now, EmailVerifiedAt: &now})
	}
	staff := func(u store.IdentityUser, no, dept, title string, roles ...string) (string, error) {
		if _, err := q.IdentityGetStaffMember(ctx, u.ID); err != nil {
			if _, err := q.IdentityCreateStaffMember(ctx, store.IdentityCreateStaffMemberParams{UserID: u.ID, EmployeeNo: no, Department: dept, JobTitle: title}); err != nil {
				return "", fmt.Errorf("staff %s: %w", no, err)
			}
		}
		for _, r := range roles {
			role, err := q.IdentityGetRoleByKey(ctx, r)
			if err != nil {
				return "", fmt.Errorf("role %s: %w", r, err)
			}
			if err := q.IdentityGrantRole(ctx, store.IdentityGrantRoleParams{UserID: u.ID, RoleID: role.ID}); err != nil {
				return "", err
			}
		}
		secret, _, err := auth.NewTOTPSecret("TechShop", *u.Email)
		if err != nil {
			return "", err
		}
		sealed, err := d.Sealer.Seal([]byte(secret), "identity.user_mfa_factors:"+u.ID.String())
		if err != nil {
			return "", err
		}
		return secret, q.IdentitySeedTOTP(ctx, store.IdentitySeedTOTPParams{UserID: u.ID, SecretEncrypted: sealed})
	}

	// Proposed role templates are DEV DATA ONLY until the owner approves the role list.
	for key, t := range rbac.RoleTemplates {
		desc := "Proposed template (dev data; owner approval pending)"
		role, err := q.IdentityUpsertRole(ctx, store.IdentityUpsertRoleParams{Key: key, Name: t.Name, Description: &desc})
		if err != nil {
			return nil, err
		}
		if err := q.IdentityClearRolePermissions(ctx, role.ID); err != nil {
			return nil, err
		}
		if err := q.IdentityAddRolePermissions(ctx, store.IdentityAddRolePermissionsParams{RoleID: role.ID, Keys: t.Perms}); err != nil {
			return nil, err
		}
	}

	// Customers.
	cust, err := user("customer@dev.techshop.ng", "+2348031234567", "Adaeze", "Okafor")
	if err != nil {
		return nil, err
	}
	out.Accounts = append(out.Accounts, Account{Role: "customer", UserID: cust.ID, Email: "customer@dev.techshop.ng", Phone: "+2348031234567", Password: password, Audience: auth.AudMarket})
	cust2, err := user("customer2@dev.techshop.ng", "+2348031234570", "Halima", "Bello")
	if err != nil {
		return nil, err
	}
	out.Accounts = append(out.Accounts, Account{Role: "customer_b", UserID: cust2.ID, Email: "customer2@dev.techshop.ng", Phone: "+2348031234570", Password: password, Audience: auth.AudMarket})

	// Staff: an admin and one account per proposed role.
	adminU, err := user("admin@dev.techshop.ng", "", "Bola", "Ade")
	if err != nil {
		return nil, err
	}
	sec, err := staff(adminU, "DEV-0001", "IT", "Administrator", "admin")
	if err != nil {
		return nil, err
	}
	out.Accounts = append(out.Accounts, Account{Role: "admin", UserID: adminU.ID, Email: "admin@dev.techshop.ng", Password: password, TOTPSecret: sec, Audience: auth.AudStaff})
	i := 2
	for key, t := range rbac.RoleTemplates {
		email := key + "@dev.techshop.ng"
		u, err := user(email, "", t.Name, "(dev)")
		if err != nil {
			return nil, err
		}
		sec, err := staff(u, fmt.Sprintf("DEV-%04d", i), "Dev", t.Name, key)
		if err != nil {
			return nil, err
		}
		i++
		out.Accounts = append(out.Accounts, Account{Role: key, UserID: u.ID, Email: email, Password: password, TOTPSecret: sec, Audience: auth.AudStaff})
	}
	// A second finance manager, so dual-control flows can be demonstrated end to end.
	fm2, err := user("finance_manager2@dev.techshop.ng", "", "Second", "Finance manager")
	if err != nil {
		return nil, err
	}
	sec, err = staff(fm2, "DEV-0099", "Finance", "Finance manager", "finance_manager")
	if err != nil {
		return nil, err
	}
	out.Accounts = append(out.Accounts, Account{Role: "finance_manager_b", UserID: fm2.ID, Email: "finance_manager2@dev.techshop.ng", Password: password, TOTPSecret: sec, Audience: auth.AudStaff})

	if err := seedCommerce(ctx, d, out, user, staff, password); err != nil {
		return nil, err
	}
	return out, nil
}
