package marketing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// sendBatch is how many recipients one send job handles before re-queueing itself.
const sendBatch = 500

// VariantInput is one content variant of a campaign (per channel, A/B by label).
type VariantInput struct {
	Label             string     `json:"label" pattern:"^[A-Z]$" doc:"A, B, …"`
	Channel           string     `json:"channel" enum:"sms,email,push"`
	Body              string     `json:"body,omitempty" maxLength:"1000" doc:"Plain text with {{first_name}}, {{coupon}} and {{link}}"`
	Subject           string     `json:"subject,omitempty" maxLength:"150"`
	TemplateVersionID *uuid.UUID `json:"templateVersionId,omitempty"`
	SharePct          float64    `json:"sharePct" minimum:"1" maximum:"100"`
}

// CampaignInput creates a campaign.
type CampaignInput struct {
	Name       string         `json:"name" minLength:"3" maxLength:"120"`
	SegmentID  uuid.UUID      `json:"segmentId"`
	Channels   []string       `json:"channels" minItems:"1" uniqueItems:"true"`
	HoldoutPct float64        `json:"holdoutPct,omitempty" minimum:"0" maximum:"50" default:"5"`
	Exclusions Exclusions     `json:"exclusions"`
	LinkURL    string         `json:"linkUrl,omitempty" pattern:"^https://([a-z0-9-]+\\.)*techshop\\.ng(/.*)?$" doc:"Tracked click target (techshop.ng only)"`
	Variants   []VariantInput `json:"variants" minItems:"1" maxItems:"8"`
}

// Exclusions remove people from a campaign at send time.
type Exclusions struct {
	OrderedWithinDays *int `json:"orderedWithinDays,omitempty" minimum:"1" maximum:"90" doc:"Skip people who bought recently"`
}

// CreateCampaign stores a draft campaign with its variants (text bodies become template versions).
func (s *Service) CreateCampaign(ctx context.Context, in CampaignInput, actor uuid.UUID) (store.MarketingCampaign, error) {
	var c store.MarketingCampaign
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		if _, err := tx.Q.MarketingGetSegment(ctx, in.SegmentID); err != nil {
			return httpx.Invalid("bad_segment", "Unknown audience.")
		}
		excl, _ := json.Marshal(in.Exclusions)
		holdout := in.HoldoutPct
		var err error
		c, err = tx.Q.MarketingCreateCampaign(ctx, store.MarketingCreateCampaignParams{Name: in.Name, SegmentID: in.SegmentID, Exclusions: excl, Channels: in.Channels,
			HoldoutPct: holdout, AbTest: abTest(in), CreatedBy: actor})
		if err != nil {
			return httpx.DB(err, "campaign")
		}
		for _, v := range in.Variants {
			if !containsStr(in.Channels, v.Channel) {
				return httpx.Invalid("bad_variant", "Variant "+v.Label+" uses channel "+v.Channel+", which the campaign doesn't send on.")
			}
			tv := v.TemplateVersionID
			if tv == nil {
				if v.Body == "" {
					return httpx.Invalid("bad_variant", "Variant "+v.Label+" needs a body or a template version.")
				}
				id, err := s.textTemplate(ctx, tx, "campaign:"+c.ID.String()+":"+v.Label, v.Channel, v.Subject, v.Body, actor)
				if err != nil {
					return err
				}
				tv = &id
			}
			if _, err := tx.Q.MarketingAddVariant(ctx, store.MarketingAddVariantParams{CampaignID: c.ID, Label: v.Label, Channel: v.Channel, TemplateVersionID: tv, SharePct: v.SharePct}); err != nil {
				return httpx.DB(err, "variant")
			}
		}
		if in.LinkURL != "" {
			if _, err := tx.Q.MarketingCreateLink(ctx, store.MarketingCreateLinkParams{Url: in.LinkURL, CampaignID: &c.ID}); err != nil {
				return httpx.DB(err, "link")
			}
		}
		return tx.Audit("campaign.created", "campaign", c.ID.String(), map[string]any{"name": in.Name, "segment": in.SegmentID})
	})
	return c, err
}

func abTest(in CampaignInput) []byte {
	labels := map[string]bool{}
	for _, v := range in.Variants {
		labels[v.Label] = true
	}
	if len(labels) < 2 {
		return nil
	}
	b, _ := json.Marshal(map[string]any{"metric": "clicks", "variants": len(labels)})
	return b
}

// textTemplate stores a plain-text body as a one-block template version.
func (s *Service) textTemplate(ctx context.Context, tx *uow.Tx, key, channel, subject, body string, actor uuid.UUID) (uuid.UUID, error) {
	t, err := tx.Q.MessagingCreateTemplate(ctx, store.MessagingCreateTemplateParams{Key: key, Channel: channel, Category: "deals", Locale: "en"})
	if err != nil {
		return uuid.Nil, httpx.DB(err, "template")
	}
	blocks, _ := json.Marshal([]map[string]string{{"type": "text", "text": body}})
	var subj *string
	if subject != "" {
		subj = &subject
	}
	v, err := tx.Q.MessagingAddTemplateVersion(ctx, store.MessagingAddTemplateVersionParams{TemplateID: t.ID, Subject: subj, Blocks: blocks, CreatedBy: actor})
	return v.ID, err
}

// Transition moves a campaign through review, approval and scheduling. Approval must come
// from someone other than the creator (also a database check) and needs marketing.approve.
func (s *Service) Transition(ctx context.Context, id uuid.UUID, action string, actor uuid.UUID, at *time.Time) (store.MarketingCampaign, error) {
	var out store.MarketingCampaign
	err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		c, err := tx.Q.MarketingGetCampaignForUpdate(ctx, id)
		if err != nil {
			return httpx.NotFound("Campaign")
		}
		next, p := "", store.MarketingSetCampaignStatusParams{ID: id}
		switch {
		case action == "submit" && c.Status == "draft":
			next = "in_review"
		case action == "approve" && c.Status == "in_review":
			if c.CreatedBy == actor {
				return httpx.E(403, "separation_of_duties", "You created this campaign, so someone else must approve it.")
			}
			next, p.ApprovedBy = "scheduled", &actor
			when := time.Now()
			if at != nil && at.After(when) {
				when = *at
			}
			p.ScheduledAt = &when
		case action == "pause" && (c.Status == "scheduled" || c.Status == "sending"):
			next = "paused"
		case action == "resume" && c.Status == "paused":
			next = "sending"
			if c.ApprovedBy == nil {
				return httpx.Conflict("not_approved", "This campaign was never approved.")
			}
			if n, err := tx.Q.MarketingCampaignStats(ctx, &id); err == nil && n.Recipients == 0 {
				next = "scheduled"
			}
		case action == "cancel" && c.Status != "sent" && c.Status != "cancelled":
			next = "cancelled"
		default:
			return httpx.Conflict("bad_transition", "Can't "+action+" a campaign that is "+c.Status+".")
		}
		p.Status = next
		out, err = tx.Q.MarketingSetCampaignStatus(ctx, p)
		if err != nil {
			return httpx.DB(err, "campaign")
		}
		if next == "sending" {
			tx.Enqueue(SendCampaignArgs{CampaignID: id}, marketingQueue())
		}
		return tx.Audit("campaign."+action, "campaign", id.String(), map[string]any{"from": c.Status, "to": next})
	})
	return out, err
}

// StartDue snapshots recipients for campaigns whose time has come and starts sending.
func (s *Service) StartDue(ctx context.Context) error {
	due, err := s.d.Q.MarketingDueCampaigns(ctx)
	if err != nil {
		return err
	}
	for _, c := range due {
		if err := s.start(ctx, c); err != nil {
			s.d.Logger.Error("campaign start failed", "campaign", c.ID, "err", err)
		}
	}
	return nil
}

// start takes the recipient snapshot (stable and auditable): holdout people get nothing,
// the rest are split between variants by share.
func (s *Service) start(ctx context.Context, c store.MarketingCampaign) error {
	seg, err := s.d.Q.MarketingGetSegment(ctx, c.SegmentID)
	if err != nil {
		return err
	}
	var rules Rules
	if err := json.Unmarshal(seg.Rules, &rules); err != nil {
		return err
	}
	var excl Exclusions
	_ = json.Unmarshal(c.Exclusions, &excl)
	members, err := s.Members(ctx, rules)
	if err != nil {
		return err
	}
	variants, err := s.d.Q.MarketingCampaignVariants(ctx, c.ID)
	if err != nil {
		return err
	}
	labels := labelShares(variants)
	return s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		cur, err := tx.Q.MarketingGetCampaignForUpdate(ctx, c.ID)
		if err != nil || cur.Status != "scheduled" {
			return err
		}
		for _, u := range members {
			if excl.OrderedWithinDays != nil {
				ordered, err := tx.Q.MarketingOrderedSince(ctx, store.MarketingOrderedSinceParams{CustomerUserID: &u, PlacedAt: time.Now().AddDate(0, 0, -*excl.OrderedWithinDays)})
				if err != nil {
					return err
				}
				if ordered {
					continue
				}
			}
			bucket := bucketOf(c.ID, u)
			if bucket < c.HoldoutPct {
				if err := tx.Q.MarketingInsertRecipient(ctx, store.MarketingInsertRecipientParams{CampaignID: c.ID, UserID: u, Holdout: true}); err != nil {
					return err
				}
				continue
			}
			label := pick(labels, (bucket-c.HoldoutPct)/(100-c.HoldoutPct)*100)
			var vid *uuid.UUID
			for _, v := range variants {
				if v.Label == label {
					id := v.ID
					vid = &id
					break
				}
			}
			if err := tx.Q.MarketingInsertRecipient(ctx, store.MarketingInsertRecipientParams{CampaignID: c.ID, UserID: u, VariantID: vid}); err != nil {
				return err
			}
		}
		if _, err := tx.Q.MarketingSetCampaignStatus(ctx, store.MarketingSetCampaignStatusParams{ID: c.ID, Status: "sending"}); err != nil {
			return err
		}
		tx.Enqueue(SendCampaignArgs{CampaignID: c.ID}, marketingQueue())
		return nil
	})
}

type share struct {
	label string
	pct   float64
}

// labelShares: variants with the same label (one per channel) share one split.
func labelShares(vs []store.MarketingCampaignVariant) []share {
	seen := map[string]bool{}
	var out []share
	for _, v := range vs {
		if !seen[v.Label] {
			seen[v.Label] = true
			out = append(out, share{v.Label, v.SharePct})
		}
	}
	return out
}

// pick chooses a label for a position 0–100 by cumulative share.
func pick(shares []share, pos float64) string {
	var total, acc float64
	for _, s := range shares {
		total += s.pct
	}
	for _, s := range shares {
		acc += s.pct / total * 100
		if pos < acc {
			return s.label
		}
	}
	return shares[len(shares)-1].label
}

// bucketOf maps a person to a stable 0–100 position for this campaign (holdout and A/B).
func bucketOf(campaign, user uuid.UUID) float64 {
	h := sha256.Sum256(append(campaign[:], user[:]...))
	return float64(binary.BigEndian.Uint64(h[:8])%10000) / 100
}

// SendBatch sends to the next pending recipients; returns true when more remain.
func (s *Service) SendBatch(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	c, err := s.d.Q.MarketingGetCampaign(ctx, campaignID)
	if err != nil || c.Status != "sending" {
		return false, err // paused or cancelled: stop
	}
	variants, err := s.d.Q.MarketingCampaignVariants(ctx, campaignID)
	if err != nil {
		return false, err
	}
	rs, err := s.d.Q.MarketingPendingRecipients(ctx, store.MarketingPendingRecipientsParams{CampaignID: campaignID, Limit: sendBatch})
	if err != nil {
		return false, err
	}
	link, _ := s.campaignLink(ctx, campaignID)
	err = s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
		for _, r := range rs {
			label := ""
			for _, v := range variants {
				if r.VariantID != nil && v.ID == *r.VariantID {
					label = v.Label
				}
			}
			sent := false
			for _, v := range variants {
				if v.Label != label {
					continue
				}
				to := ""
				switch v.Channel {
				case "sms":
					to = derefStr(r.Phone)
				case "email":
					to = derefStr(r.Email)
				}
				if v.Channel != "push" && to == "" {
					continue
				}
				data := map[string]any{}
				if link != nil {
					data["link"] = s.TrackedURL(link.ID, r.UserID)
				}
				cid := campaignID
				if err := s.Send(ctx, tx, Message{UserID: r.UserID, To: to, Channel: v.Channel, Template: "campaign", Data: data,
					Key: "campaign:" + campaignID.String() + ":" + r.UserID.String() + ":" + v.Channel, CampaignID: &cid, TemplateVersionID: v.TemplateVersionID}); err != nil {
					return err
				}
				sent = true
			}
			status := "sent"
			if !sent {
				status = "skipped"
			}
			if err := tx.Q.MarketingSetRecipientStatus(ctx, store.MarketingSetRecipientStatusParams{CampaignID: campaignID, UserID: r.UserID, Status: status}); err != nil {
				return err
			}
		}
		if len(rs) < sendBatch {
			_, err := tx.Q.MarketingSetCampaignStatus(ctx, store.MarketingSetCampaignStatusParams{ID: campaignID, Status: "sent"})
			return err
		}
		return nil
	})
	return len(rs) == sendBatch, err
}

func (s *Service) campaignLink(ctx context.Context, campaignID uuid.UUID) (*store.MarketingTrackedLink, error) {
	l, err := s.d.Q.MarketingCampaignLink(ctx, &campaignID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &l, err
}

// ---- Signed tokens: tracked links and one-click unsubscribe (TOKEN_HMAC_KEY).

func (s *Service) sign(parts ...string) string {
	m := hmac.New(sha256.New, []byte(s.d.Cfg.TokenHMACKey))
	for _, p := range parts {
		m.Write([]byte(p))
		m.Write([]byte{0})
	}
	return hex.EncodeToString(m.Sum(nil))[:32]
}

// TrackedURL is the click-redirect URL for one person (signed, so ids can't be swapped).
func (s *Service) TrackedURL(linkID, userID uuid.UUID) string {
	return s.d.Cfg.AppBaseURL + "/api/v1/l/" + linkID.String() + "?u=" + userID.String() + "&s=" + s.sign("link", linkID.String(), userID.String())
}

// UnsubscribeURL is the one-click unsubscribe URL for a person and channel.
func (s *Service) UnsubscribeURL(userID uuid.UUID, channel string) string {
	return s.d.Cfg.AppBaseURL + "/api/v1/unsubscribe?u=" + userID.String() + "&c=" + channel + "&s=" + s.sign("unsub", userID.String(), channel)
}

func (s *Service) validSig(sig string, parts ...string) bool {
	return hmac.Equal([]byte(sig), []byte(s.sign(parts...)))
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func containsStr(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
