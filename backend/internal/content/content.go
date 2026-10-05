// Package content serves CMS pages, banners and help articles. Editors save drafts; publishing
// needs content.publish and a different person from the last editor (docs/backend.md §3).
package content

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

// Service is the content module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "content" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	httpx.Register(g, httpx.Op{ID: "getPage", Method: http.MethodGet, Path: "/api/v1/content/pages/{site}/{slug}", Tag: "Content", Summary: "A published CMS page"},
		func(ctx context.Context, in *struct {
			Site string `path:"site" enum:"corporate,market,wholesale,seller"`
			Slug string `path:"slug"`
		}) (*struct{ Body store.ContentCmsPage }, error) {
			p, err := s.d.Q.ContentGetPublishedPage(ctx, store.ContentGetPublishedPageParams{Site: in.Site, Slug: in.Slug})
			return &struct{ Body store.ContentCmsPage }{p}, httpx.DB(err, "page")
		})
	httpx.Register(g, httpx.Op{ID: "listBanners", Method: http.MethodGet, Path: "/api/v1/content/banners", Tag: "Content", Summary: "Live banners for a site"},
		func(ctx context.Context, in *struct {
			Site      string `query:"site" enum:"corporate,market,wholesale,seller" required:"true"`
			Placement string `query:"placement"`
		}) (*struct{ Body []store.ContentBanner }, error) {
			var pl *string
			if in.Placement != "" {
				pl = &in.Placement
			}
			rows, err := s.d.Q.ContentLiveBanners(ctx, store.ContentLiveBannersParams{Site: in.Site, Placement: pl})
			return &struct{ Body []store.ContentBanner }{rows}, err
		})
	httpx.Register(g, httpx.Op{ID: "listHelp", Method: http.MethodGet, Path: "/api/v1/content/help/{site}", Tag: "Content", Summary: "Help centre index"},
		func(ctx context.Context, in *struct {
			Site string `path:"site" enum:"market,wholesale,seller"`
		}) (*struct{ Body []store.ContentListHelpRow }, error) {
			rows, err := s.d.Q.ContentListHelp(ctx, in.Site)
			return &struct{ Body []store.ContentListHelpRow }{rows}, err
		})
	httpx.Register(g, httpx.Op{ID: "getHelp", Method: http.MethodGet, Path: "/api/v1/content/help/{site}/{slug}", Tag: "Content", Summary: "A help article"},
		func(ctx context.Context, in *struct {
			Site string `path:"site" enum:"market,wholesale,seller"`
			Slug string `path:"slug"`
		}) (*struct{ Body store.ContentHelpArticle }, error) {
			a, err := s.d.Q.ContentGetHelp(ctx, store.ContentGetHelpParams{Site: in.Site, Slug: in.Slug})
			return &struct{ Body store.ContentHelpArticle }{a}, httpx.DB(err, "article")
		})

	st := func(o httpx.Op) httpx.Op { o.Tag = "Staff · Content"; o.Auth = []string{auth.AudStaff}; return o }
	httpx.Register(g, st(httpx.Op{ID: "staffListPages", Method: http.MethodGet, Path: "/api/v1/staff/content/pages", Perm: "content.edit", Summary: "All pages and their status"}),
		func(ctx context.Context, in *struct {
			Site string `query:"site"`
		}) (*struct{ Body []store.ContentCmsPage }, error) {
			var site *string
			if in.Site != "" {
				site = &in.Site
			}
			rows, err := s.d.Q.ContentListPages(ctx, site)
			return &struct{ Body []store.ContentCmsPage }{rows}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffSavePage", Method: http.MethodPut, Path: "/api/v1/staff/content/pages/{site}/{slug}", Perm: "content.edit", Summary: "Save a draft (blocks)"}),
		func(ctx context.Context, in *struct {
			Site string `path:"site" enum:"corporate,market,wholesale,seller"`
			Slug string `path:"slug" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
			Body struct {
				Title  string          `json:"title" minLength:"2"`
				Blocks json.RawMessage `json:"blocks" doc:"Content blocks (array)"`
			}
		}) (*struct{ Body store.ContentCmsPage }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.ContentCmsPage
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.ContentUpsertPage(ctx, store.ContentUpsertPageParams{Site: in.Site, Slug: in.Slug, Title: in.Body.Title, Body: in.Body.Blocks, UpdatedBy: &p.UserID})
				if err != nil {
					return httpx.DB(err, "page")
				}
				return tx.Audit("page.saved", "cms_page", out.ID.String(), map[string]string{"site": in.Site, "slug": in.Slug})
			})
			return &struct{ Body store.ContentCmsPage }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffPublishPage", Method: http.MethodPost, Path: "/api/v1/staff/content/pages/{id}/publish", Perm: "content.publish", Summary: "Publish (someone other than the last editor)"}),
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct{ Body store.ContentCmsPage }, error) {
			p := httpx.MustPrincipal(ctx)
			var out store.ContentCmsPage
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				pg, err := tx.Q.ContentGetPage(ctx, in.ID)
				if err != nil {
					return httpx.DB(err, "page")
				}
				if pg.UpdatedBy != nil && *pg.UpdatedBy == p.UserID {
					return httpx.Forbidden("Someone other than the last editor must publish.")
				}
				out, err = tx.Q.ContentSetPageStatus(ctx, store.ContentSetPageStatusParams{ID: in.ID, Status: "published", ApprovedBy: &p.UserID})
				if err != nil {
					return err
				}
				return tx.Audit("page.published", "cms_page", in.ID.String(), nil)
			})
			return &struct{ Body store.ContentCmsPage }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffCreateBanner", Method: http.MethodPost, Path: "/api/v1/staff/content/banners", Perm: "content.publish", Status: 201, Summary: "Create a banner"}),
		func(ctx context.Context, in *struct {
			Body struct {
				Site        string     `json:"site" enum:"corporate,market,wholesale,seller"`
				Placement   string     `json:"placement"`
				Title       string     `json:"title"`
				ImageFileID *uuid.UUID `json:"imageFileId,omitempty"`
				LinkURL     string     `json:"linkUrl" pattern:"^(/|https://)"`
				StartsAt    *time.Time `json:"startsAt,omitempty"`
				EndsAt      *time.Time `json:"endsAt,omitempty"`
				Position    int32      `json:"position,omitempty"`
				Status      string     `json:"status" enum:"draft,live"`
			}
		}) (*struct{ Body store.ContentBanner }, error) {
			var out store.ContentBanner
			err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error {
				var err error
				out, err = tx.Q.ContentUpsertBanner(ctx, store.ContentUpsertBannerParams{Site: in.Body.Site, Placement: in.Body.Placement, Title: in.Body.Title, ImageFileID: in.Body.ImageFileID,
					LinkUrl: in.Body.LinkURL, StartsAt: in.Body.StartsAt, EndsAt: in.Body.EndsAt, Position: in.Body.Position, Status: in.Body.Status})
				if err != nil {
					return httpx.DB(err, "banner")
				}
				return tx.Audit("banner.created", "banner", out.ID.String(), in.Body)
			})
			return &struct{ Body store.ContentBanner }{out}, err
		})
	httpx.Register(g, st(httpx.Op{ID: "staffSaveHelp", Method: http.MethodPut, Path: "/api/v1/staff/content/help/{site}/{slug}", Perm: "content.publish", Summary: "Create or update a help article"}),
		func(ctx context.Context, in *struct {
			Site string `path:"site" enum:"market,wholesale,seller"`
			Slug string `path:"slug" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
			Body struct {
				Topic  string          `json:"topic"`
				Title  string          `json:"title"`
				Blocks json.RawMessage `json:"blocks"`
				Status string          `json:"status" enum:"draft,in_review,published"`
			}
		}) (*struct{ Body store.ContentHelpArticle }, error) {
			row, err := s.d.Q.ContentUpsertHelp(ctx, store.ContentUpsertHelpParams{Site: in.Site, Topic: in.Body.Topic, Slug: in.Slug, Title: in.Body.Title, Body: in.Body.Blocks, Status: in.Body.Status})
			return &struct{ Body store.ContentHelpArticle }{row}, httpx.DB(err, "article")
		})
}
