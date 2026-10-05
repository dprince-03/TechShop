package inventory

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

var staff = []string{auth.AudStaff}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	op := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Inventory"; o.Auth = staff; return o }

	httpx.Register(g, op(httpx.Op{ID: "staffListWarehouses", Method: http.MethodGet, Path: "/api/v1/staff/inventory/warehouses", Perm: "inventory.view", Summary: "Warehouses, stores and hubs"}),
		func(ctx context.Context, _ *struct{}) (*struct{ Body []store.InventoryWarehouse }, error) {
			rows, err := s.d.Q.InventoryListWarehouses(ctx)
			return &struct{ Body []store.InventoryWarehouse }{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffCreateWarehouse", Method: http.MethodPost, Path: "/api/v1/staff/inventory/warehouses", Perm: "inventory.adjust", Status: 201, Summary: "Add a location"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Code      string   `json:"code" pattern:"^[A-Z0-9-]{2,12}$"`
				Name      string   `json:"name" minLength:"2"`
				Kind      string   `json:"kind" enum:"warehouse,store,hub"`
				Address   string   `json:"address" minLength:"3"`
				City      string   `json:"city" minLength:"2"`
				StateCode string   `json:"stateCode" pattern:"^[A-Z]{2}$"`
				Latitude  *float64 `json:"latitude,omitempty"`
				Longitude *float64 `json:"longitude,omitempty"`
			}
		}) (*struct{ Body store.InventoryWarehouse }, error) {
			var out store.InventoryWarehouse
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.InventoryCreateWarehouse(ctx, store.InventoryCreateWarehouseParams{Code: in.Body.Code, Name: in.Body.Name, Kind: in.Body.Kind, Address: in.Body.Address,
					City: in.Body.City, StateCode: in.Body.StateCode, Latitude: in.Body.Latitude, Longitude: in.Body.Longitude})
				if err != nil {
					return httpx.DB(err, "warehouse")
				}
				return tx.Audit("warehouse.saved", "warehouse", out.ID.String(), in.Body)
			})
			return &struct{ Body store.InventoryWarehouse }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffListStock", Method: http.MethodGet, Path: "/api/v1/staff/inventory/levels", Perm: "inventory.view", Summary: "On hand, reserved and available"}),
		func(ctx context.Context, in *struct {
			WarehouseID string `query:"warehouseId"`
			VariantID   string `query:"variantId"`
		}) (*struct {
			Body []store.InventoryListLevelsRow
		}, error) {
			p := store.InventoryListLevelsParams{Limit: 500}
			if id, err := uuid.Parse(in.WarehouseID); err == nil {
				p.WarehouseID = &id
			}
			if id, err := uuid.Parse(in.VariantID); err == nil {
				p.VariantID = &id
			}
			rows, err := s.d.Q.InventoryListLevels(ctx, p)
			return &struct {
				Body []store.InventoryListLevelsRow
			}{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffListMovements", Method: http.MethodGet, Path: "/api/v1/staff/inventory/movements", Perm: "inventory.view", Summary: "Stock movement ledger"}),
		func(ctx context.Context, in *struct {
			VariantID string `query:"variantId"`
		}) (*struct {
			Body []store.InventoryStockMovement
		}, error) {
			p := store.InventoryListMovementsParams{Limit: 200}
			if id, err := uuid.Parse(in.VariantID); err == nil {
				p.VariantID = &id
			}
			rows, err := s.d.Q.InventoryListMovements(ctx, p)
			return &struct {
				Body []store.InventoryStockMovement
			}{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffAdjustStock", Method: http.MethodPost, Path: "/api/v1/staff/inventory/adjustments", Perm: "inventory.adjust", Summary: "Adjust stock (writes a movement; needs step-up)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				WarehouseID   uuid.UUID  `json:"warehouseId"`
				VariantID     uuid.UUID  `json:"variantId"`
				Condition     string     `json:"condition" enum:"new,uk_used,refurbished,open_box"`
				OwnerSellerID *uuid.UUID `json:"ownerSellerId,omitempty" doc:"Defaults to TechShop"`
				Delta         int32      `json:"delta"`
				Reason        string     `json:"reason" enum:"adjustment,write_off,purchase_receipt"`
				Note          string     `json:"note" minLength:"3"`
			}
		}) (*struct{ Body store.InventoryInventoryLevel }, error) {
			if in.Body.Delta == 0 {
				return nil, httpx.Invalid("zero_delta", "Delta can't be zero.")
			}
			p := httpx.MustPrincipal(ctx)
			var out store.InventoryInventoryLevel
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				owner := in.Body.OwnerSellerID
				if owner == nil {
					fp, err := tx.Q.SellersGetFirstParty(ctx)
					if err != nil {
						return err
					}
					owner = &fp.ID
				}
				var err error
				out, err = s.Apply(ctx, tx, Move{WarehouseID: in.Body.WarehouseID, VariantID: in.Body.VariantID, Condition: in.Body.Condition, OwnerSellerID: *owner,
					Delta: in.Body.Delta, Reason: in.Body.Reason, RefType: "adjustment", Actor: &p.UserID})
				if err != nil {
					return err
				}
				return tx.Audit("stock.adjusted", "variant", in.Body.VariantID.String(), in.Body)
			})
			return &struct{ Body store.InventoryInventoryLevel }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffReceiveUnits", Method: http.MethodPost, Path: "/api/v1/staff/inventory/device-units", Perm: "purchasing.manage", Status: 201, Summary: "Receive serialised devices (IMEI Luhn + blocklist checked)"}),
		func(ctx context.Context, in *struct {
			Body struct {
				WarehouseID uuid.UUID `json:"warehouseId"`
				VariantID   uuid.UUID `json:"variantId"`
				Condition   string    `json:"condition" enum:"new,uk_used,refurbished,open_box"`
				CostKobo    *int64    `json:"costKobo,omitempty"`
				Units       []struct {
					IMEI   string `json:"imei,omitempty" pattern:"^([0-9]{15})?$"`
					Serial string `json:"serial,omitempty"`
				} `json:"units" minItems:"1" maxItems:"200"`
			}
		}) (*struct{ Body []store.InventoryDeviceUnit }, error) {
			p := httpx.MustPrincipal(ctx)
			var out []store.InventoryDeviceUnit
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				fp, err := tx.Q.SellersGetFirstParty(ctx)
				if err != nil {
					return err
				}
				for _, u := range in.Body.Units {
					if u.IMEI == "" && u.Serial == "" {
						return httpx.Invalid("imei_or_serial_required", "Each unit needs an IMEI or a serial number.")
					}
					unit, err := s.ReceiveUnit(ctx, tx, in.Body.VariantID, in.Body.WarehouseID, fp.ID, in.Body.Condition, u.IMEI, u.Serial, "purchase", in.Body.CostKobo, "receipt", nil, &p.UserID)
					if err != nil {
						return err
					}
					out = append(out, unit)
				}
				return tx.Audit("stock.units_received", "variant", in.Body.VariantID.String(), map[string]any{"count": len(out)})
			})
			return &struct{ Body []store.InventoryDeviceUnit }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffListDeviceUnits", Method: http.MethodGet, Path: "/api/v1/staff/inventory/device-units", Perm: "inventory.view", Summary: "Serialised units"}),
		func(ctx context.Context, in *struct {
			VariantID string `query:"variantId"`
			Status    string `query:"status"`
		}) (*struct{ Body []store.InventoryDeviceUnit }, error) {
			p := store.InventoryListDeviceUnitsParams{Limit: 200}
			if id, err := uuid.Parse(in.VariantID); err == nil {
				p.VariantID = &id
			}
			if in.Status != "" {
				p.Status = &in.Status
			}
			rows, err := s.d.Q.InventoryListDeviceUnits(ctx, p)
			return &struct{ Body []store.InventoryDeviceUnit }{rows}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffCreateTransfer", Method: http.MethodPost, Path: "/api/v1/staff/inventory/transfers", Perm: "inventory.adjust", Status: 201, Summary: "Move stock between locations"}),
		func(ctx context.Context, in *struct {
			Body struct {
				From  uuid.UUID `json:"fromWarehouseId"`
				To    uuid.UUID `json:"toWarehouseId"`
				Lines []struct {
					VariantID uuid.UUID `json:"variantId"`
					Condition string    `json:"condition" enum:"new,uk_used,refurbished,open_box"`
					Quantity  int32     `json:"quantity" minimum:"1"`
				} `json:"lines" minItems:"1"`
			}
		}) (*struct{ Body store.InventoryStockTransfer }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.InventoryStockTransfer
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				fp, err := tx.Q.SellersGetFirstParty(ctx)
				if err != nil {
					return err
				}
				out, err = tx.Q.InventoryCreateTransfer(ctx, store.InventoryCreateTransferParams{FromWarehouseID: in.Body.From, ToWarehouseID: in.Body.To, CreatedBy: p.UserID})
				if err != nil {
					return httpx.DB(err, "transfer")
				}
				for _, l := range in.Body.Lines {
					if _, err := tx.Q.InventoryAddTransferLine(ctx, store.InventoryAddTransferLineParams{TransferID: out.ID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: fp.ID, Quantity: l.Quantity}); err != nil {
						return err
					}
				}
				return tx.Audit("transfer.created", "transfer", out.ID.String(), in.Body)
			})
			return &struct{ Body store.InventoryStockTransfer }{out}, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffAdvanceTransfer", Method: http.MethodPost, Path: "/api/v1/staff/inventory/transfers/{id}/{action}", Perm: "inventory.adjust", Summary: "Ship or receive a transfer (writes movements)"}),
		func(ctx context.Context, in *struct {
			ID     uuid.UUID `path:"id"`
			Action string    `path:"action" enum:"ship,receive,cancel"`
		}) (*struct{ Body store.InventoryStockTransfer }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.InventoryStockTransfer
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				t, err := tx.Q.InventoryGetTransfer(ctx, in.ID)
				if err != nil {
					return httpx.DB(err, "transfer")
				}
				lines, err := tx.Q.InventoryTransferLines(ctx, in.ID)
				if err != nil {
					return err
				}
				next := map[string][2]string{"ship": {"draft", "in_transit"}, "receive": {"in_transit", "received"}, "cancel": {"draft", "cancelled"}}[in.Action]
				if t.Status != next[0] {
					return httpx.Conflict("bad_transition", "This transfer is "+t.Status+".")
				}
				for _, l := range lines {
					switch in.Action {
					case "ship":
						_, err = s.Apply(ctx, tx, Move{WarehouseID: t.FromWarehouseID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: l.OwnerSellerID, Delta: -l.Quantity, Reason: "transfer_out", RefType: "transfer", RefID: &t.ID, Actor: &p.UserID})
					case "receive":
						_, err = s.Apply(ctx, tx, Move{WarehouseID: t.ToWarehouseID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: l.OwnerSellerID, Delta: l.Quantity, Reason: "transfer_in", RefType: "transfer", RefID: &t.ID, Actor: &p.UserID})
					}
					if err != nil {
						return err
					}
				}
				out, err = tx.Q.InventorySetTransferStatus(ctx, store.InventorySetTransferStatusParams{ID: in.ID, Status: next[1]})
				if err != nil {
					return err
				}
				return tx.Audit("transfer."+in.Action, "transfer", in.ID.String(), nil)
			})
			return &struct{ Body store.InventoryStockTransfer }{out}, err
		})
	s.countRoutes(g)
}

// countRoutes: blind cycle counts; variances are approved by a different person.
func (s *Service) countRoutes(g *httpx.Guard) {
	op := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Inventory"; o.Auth = staff; return o }
	httpx.Register(g, op(httpx.Op{ID: "staffCreateCount", Method: http.MethodPost, Path: "/api/v1/staff/inventory/counts", Perm: "fulfilment.pick", Status: 201, Summary: "Start a blind cycle count"}),
		func(ctx context.Context, in *struct {
			Body struct {
				WarehouseID uuid.UUID `json:"warehouseId"`
				Lines       []struct {
					VariantID uuid.UUID `json:"variantId"`
					Condition string    `json:"condition" enum:"new,uk_used,refurbished,open_box"`
					Counted   int32     `json:"counted" minimum:"0"`
				} `json:"lines" minItems:"1"`
			}
		}) (*struct {
			Body struct {
				Count store.InventoryInventoryCount       `json:"count"`
				Lines []store.InventoryInventoryCountLine `json:"lines"`
			}
		}, error) {
			p := httpx.MustPrincipal(ctx)
			out := &struct {
				Body struct {
					Count store.InventoryInventoryCount       `json:"count"`
					Lines []store.InventoryInventoryCountLine `json:"lines"`
				}
			}{}
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				fp, err := tx.Q.SellersGetFirstParty(ctx)
				if err != nil {
					return err
				}
				c, err := tx.Q.InventoryCreateCount(ctx, store.InventoryCreateCountParams{WarehouseID: in.Body.WarehouseID, CountedBy: p.UserID})
				if err != nil {
					return httpx.DB(err, "count")
				}
				out.Body.Count = c
				for _, l := range in.Body.Lines {
					counted := l.Counted
					line, err := tx.Q.InventoryAddCountLine(ctx, store.InventoryAddCountLineParams{CountID: c.ID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: fp.ID, CountedQty: &counted})
					if err != nil {
						return err
					}
					out.Body.Lines = append(out.Body.Lines, line)
				}
				_, err = tx.Q.InventorySetCountStatus(ctx, store.InventorySetCountStatusParams{ID: c.ID, Status: "submitted"})
				return err
			})
			return out, err
		})
	httpx.Register(g, op(httpx.Op{ID: "staffApproveCount", Method: http.MethodPost, Path: "/api/v1/staff/inventory/counts/{id}/approve", Perm: "inventory.adjust", Summary: "Approve variances (different person; posts count_variance movements)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.InventoryInventoryCount }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.InventoryInventoryCount
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				c, err := tx.Q.InventoryGetCount(ctx, in.ID)
				if err != nil {
					return httpx.DB(err, "count")
				}
				if c.CountedBy == p.UserID {
					return httpx.Forbidden("Someone other than the counter must approve the variance.")
				}
				if c.Status != "submitted" {
					return httpx.Conflict("bad_transition", "This count is "+c.Status+".")
				}
				lines, err := tx.Q.InventoryCountLines(ctx, in.ID)
				if err != nil {
					return err
				}
				for _, l := range lines {
					if l.Variance == nil || *l.Variance == 0 {
						continue
					}
					if _, err := s.Apply(ctx, tx, Move{WarehouseID: c.WarehouseID, VariantID: l.VariantID, Condition: l.Condition, OwnerSellerID: l.OwnerSellerID,
						Delta: *l.Variance, Reason: "count_variance", RefType: "count", RefID: &c.ID, Actor: &p.UserID}); err != nil {
						return err
					}
				}
				out, err = tx.Q.InventorySetCountStatus(ctx, store.InventorySetCountStatusParams{ID: in.ID, Status: "approved", ApprovedBy: &p.UserID})
				if err != nil {
					return err
				}
				return tx.Audit("count.approved", "count", in.ID.String(), nil)
			})
			return &struct{ Body store.InventoryInventoryCount }{out}, err
		})
}

var _ kit.Module = (*Service)(nil)
