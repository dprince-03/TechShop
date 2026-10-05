package marketing

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Rules is a saved audience. Only these fields exist (an allowlist), and every value is bound
// as a parameter, so a segment can never inject SQL. This is the second sanctioned dynamic
// query in the codebase (after search) because audiences combine optional filters freely.
type Rules struct {
	States               []string   `json:"states,omitempty" doc:"Default-address state codes" maxItems:"37"`
	MinOrders            *int       `json:"minOrders,omitempty" minimum:"0"`
	MaxOrders            *int       `json:"maxOrders,omitempty" minimum:"0"`
	OrderedWithinDays    *int       `json:"orderedWithinDays,omitempty" minimum:"1" maximum:"730"`
	NotOrderedWithinDays *int       `json:"notOrderedWithinDays,omitempty" minimum:"1" maximum:"730"`
	SignedUpWithinDays   *int       `json:"signedUpWithinDays,omitempty" minimum:"1" maximum:"3650"`
	BoughtCategoryID     *uuid.UUID `json:"boughtCategoryId,omitempty" doc:"Bought anything in this category or below"`
	MinSpendKobo         *int64     `json:"minSpendKobo,omitempty" minimum:"0"`
	Consent              string     `json:"consent,omitempty" enum:"sms,email,push," doc:"Only people who opted in to this marketing channel"`
}

// paidOrders are orders that count as purchases.
const paidOrders = "o.customer_user_id = u.id and o.status not in ('pending_payment', 'cancelled')"

// SQL compiles the rules into a query returning matching user ids. Staff, suspended and
// deleted accounts never match.
func (r Rules) SQL() (string, []any) {
	var where []string
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	where = append(where, "u.status = 'active'", "not exists (select 1 from identity.staff_members sm where sm.user_id = u.id)")
	if len(r.States) > 0 {
		where = append(where, "exists (select 1 from identity.addresses a where a.user_id = u.id and a.is_default and a.state_code = any("+arg(r.States)+"::text[]))")
	}
	if r.MinOrders != nil {
		where = append(where, "(select count(*) from sales.orders o where "+paidOrders+") >= "+arg(*r.MinOrders))
	}
	if r.MaxOrders != nil {
		where = append(where, "(select count(*) from sales.orders o where "+paidOrders+") <= "+arg(*r.MaxOrders))
	}
	if r.OrderedWithinDays != nil {
		where = append(where, "exists (select 1 from sales.orders o where "+paidOrders+" and o.placed_at > now() - make_interval(days => "+arg(*r.OrderedWithinDays)+"))")
	}
	if r.NotOrderedWithinDays != nil {
		where = append(where, "not exists (select 1 from sales.orders o where "+paidOrders+" and o.placed_at > now() - make_interval(days => "+arg(*r.NotOrderedWithinDays)+"))")
	}
	if r.SignedUpWithinDays != nil {
		where = append(where, "u.created_at > now() - make_interval(days => "+arg(*r.SignedUpWithinDays)+")")
	}
	if r.BoughtCategoryID != nil {
		where = append(where, `exists (
  with recursive sub as (select id from catalog.categories where id = `+arg(*r.BoughtCategoryID)+`
                         union all select c.id from catalog.categories c join sub on c.parent_id = sub.id)
  select 1 from sales.orders o join sales.order_lines ol on ol.order_id = o.id
  join catalog.product_variants v on v.id = ol.variant_id join catalog.products p on p.id = v.product_id
  where `+paidOrders+` and p.category_id in (select id from sub))`)
	}
	if r.MinSpendKobo != nil {
		where = append(where, "(select coalesce(sum(o.total_kobo), 0) from sales.orders o where "+paidOrders+") >= "+arg(*r.MinSpendKobo))
	}
	if r.Consent != "" {
		where = append(where, `coalesce((select c.granted from identity.consents c where c.user_id = u.id and c.purpose = `+arg("marketing_"+r.Consent)+`
  order by c.recorded_at desc limit 1), false)`)
	}
	return "select u.id from identity.users u where " + strings.Join(where, "\n  and "), args
}

// Count returns how many people match.
func (s *Service) Count(ctx context.Context, r Rules) (int, error) {
	q, args := r.SQL()
	var n int
	err := s.d.Pool.QueryRow(ctx, "select count(*) from ("+q+") x", args...).Scan(&n)
	return n, err
}

// Members returns the matching user ids (used to snapshot campaign recipients).
func (s *Service) Members(ctx context.Context, r Rules) ([]uuid.UUID, error) {
	q, args := r.SQL()
	rows, err := s.d.Pool.Query(ctx, q+" order by u.id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
