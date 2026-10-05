package seed

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/store"
)

func init() { Hooks = append(Hooks, seedLogistics) }

// riderDeviceID is the dev rider's pre-approved phone (send it as deviceId when signing in).
const riderDeviceID = "dev-rider-phone-0001"

// seedLogistics adds delivery zones with rate bands (Lagos and Abuja; everywhere else ships by
// carrier), the fake carrier, and a rider (staff member + rider + approved phone). Dev only.
func seedLogistics(ctx context.Context, d *kit.Deps, out *Output, ids map[string]uuid.UUID) error {
	q := d.Q
	acct := map[string]Account{}
	for _, a := range out.Accounts {
		acct[a.Role] = a
	}
	dispatcher, ok := acct["dispatcher"]
	if !ok {
		return errors.New("seed m3: dispatcher account missing")
	}

	zones, err := q.LogisticsListZones(ctx)
	if err != nil {
		return err
	}
	if len(zones) == 0 {
		// Bands: ≤2 kg, ≤10 kg, ≤30 kg. Seller-shipped deliveries cost a little more (pickup leg).
		bands := []struct {
			max        int32
			ts, seller int64
		}{{2000, 250000, 300000}, {10000, 400000, 450000}, {30000, 800000, 900000}}
		for _, z := range []struct{ name, state string }{{"Lagos", "LA"}, {"Abuja FCT", "FC"}} {
			zone, err := q.LogisticsCreateZone(ctx, store.LogisticsCreateZoneParams{Name: z.name, StateCode: z.state, Lgas: []string{}, IsActive: true})
			if err != nil {
				return err
			}
			for _, b := range bands {
				for fb, fee := range map[string]int64{"techshop": b.ts, "seller": b.seller} {
					if _, err := q.LogisticsCreateRate(ctx, store.LogisticsCreateRateParams{ZoneID: zone.ID, FulfilledBy: fb, MaxWeightGrams: b.max, FeeKobo: fee,
						EtaMinDays: 0, EtaMaxDays: 1, ValidFrom: time.Now().AddDate(0, 0, -1)}); err != nil {
						return err
					}
				}
			}
			out.Extra["zone_"+z.state] = zone.ID.String()
		}
	} else {
		for _, z := range zones {
			out.Extra["zone_"+z.StateCode] = z.ID.String()
		}
	}
	if _, err := q.FulfilmentUpsertCarrier(ctx, store.FulfilmentUpsertCarrierParams{Code: "fake", Name: "Test Courier (fake)", IsActive: true}); err != nil {
		return err
	}

	// Rider: signs in with phone OTP on the logistics app.
	email, phone := "rider@dev.techshop.ng", "+2348031234580"
	u, err := q.IdentityGetUserByEmail(ctx, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		now := time.Now()
		u, err = q.IdentityCreateUser(ctx, store.IdentityCreateUserParams{Email: &email, Phone: &phone, FirstName: "Musa", LastName: "Danjuma", PhoneVerifiedAt: &now})
	}
	if err != nil {
		return err
	}
	if _, err := q.IdentityGetStaffMember(ctx, u.ID); err != nil {
		if _, err := q.IdentityCreateStaffMember(ctx, store.IdentityCreateStaffMemberParams{UserID: u.ID, EmployeeNo: "RDR-001", Department: "Logistics", JobTitle: "Rider"}); err != nil {
			return err
		}
	}
	if _, err := q.LogisticsGetRider(ctx, u.ID); err != nil {
		plate := "LSD-123-XY"
		if _, err := q.LogisticsCreateRider(ctx, store.LogisticsCreateRiderParams{UserID: u.ID, VehicleType: "bike", PlateNumber: &plate, HomeWarehouseID: ids["wh-ikj"]}); err != nil {
			return err
		}
		if zid, err := uuid.Parse(out.Extra["zone_LA"]); err == nil {
			if err := q.LogisticsAddRiderZone(ctx, store.LogisticsAddRiderZoneParams{RiderID: u.ID, ZoneID: zid}); err != nil {
				return err
			}
		}
	}
	dev, err := q.LogisticsRegisterDevice(ctx, store.LogisticsRegisterDeviceParams{RiderID: u.ID, DeviceID: riderDeviceID, Platform: "android"})
	if err != nil {
		return err
	}
	if dev.ApprovedAt == nil {
		if _, err := q.LogisticsApproveDevice(ctx, store.LogisticsApproveDeviceParams{ID: dev.ID, ApprovedBy: &dispatcher.UserID}); err != nil {
			return err
		}
	}
	out.Accounts = append(out.Accounts, Account{Role: "rider", UserID: u.ID, Email: email, Phone: phone, Audience: auth.AudLogistics})
	out.Extra["riderDeviceId"] = riderDeviceID
	return nil
}
