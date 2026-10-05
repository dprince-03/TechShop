// Package files issues presigned uploads and downloads (bytes never pass through the API).
// Allowed types, sizes and visibility depend on the purpose; private downloads expire after 5
// minutes and KYC downloads are audited (docs/backend.md §3 Files).
package files

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dprince-03/techshop/backend/internal/auth"
	"github.com/dprince-03/techshop/backend/internal/httpx"
	"github.com/dprince-03/techshop/backend/internal/jobs"
	"github.com/dprince-03/techshop/backend/internal/kit"
	"github.com/dprince-03/techshop/backend/internal/outbox"
	"github.com/dprince-03/techshop/backend/internal/store"
	"github.com/dprince-03/techshop/backend/internal/uow"
)

type rule struct {
	types   []string
	maxSize int64
	private bool
}

var rules = map[string]rule{
	"product_image":     {[]string{"image/jpeg", "image/png", "image/webp"}, 8 << 20, false},
	"car_image":         {[]string{"image/jpeg", "image/png", "image/webp"}, 10 << 20, false},
	"review_image":      {[]string{"image/jpeg", "image/png", "image/webp"}, 8 << 20, false},
	"banner":            {[]string{"image/jpeg", "image/png", "image/webp"}, 8 << 20, false},
	"brand_logo":        {[]string{"image/png", "image/svg+xml", "image/webp"}, 2 << 20, false},
	"kyc_document":      {[]string{"image/jpeg", "image/png", "application/pdf"}, 10 << 20, true},
	"delivery_proof":    {[]string{"image/jpeg", "image/png", "image/webp"}, 8 << 20, true},
	"return_photo":      {[]string{"image/jpeg", "image/png", "image/webp"}, 8 << 20, true},
	"ticket_attachment": {[]string{"image/jpeg", "image/png", "application/pdf"}, 10 << 20, true},
	"car_document":      {[]string{"image/jpeg", "image/png", "application/pdf"}, 10 << 20, true},
	"dispute_evidence":  {[]string{"image/jpeg", "image/png", "application/pdf"}, 10 << 20, true},
}

// Service is the files module.
type Service struct{ d *kit.Deps }

// New creates the module.
func New(d *kit.Deps) *Service { return &Service{d: d} }

// Name implements kit.Module.
func (s *Service) Name() string { return "files" }

// Jobs implements kit.Module.
func (s *Service) Jobs(*jobs.Registry) {}

// Events implements kit.Module.
func (s *Service) Events(*outbox.Relay) {}

var anyAud = []string{auth.AudMarket, auth.AudWholesale, auth.AudCustomerApp, auth.AudSeller, auth.AudStaff, auth.AudLogistics}

// Routes implements kit.Module.
func (s *Service) Routes(r *kit.Routes) {
	g := r.Public
	httpx.Register(g, httpx.Op{ID: "createUpload", Method: http.MethodPost, Path: "/api/v1/files/uploads", Tag: "Files", Auth: anyAud, Status: 201, Summary: "Get a presigned upload URL"},
		func(ctx context.Context, in *struct {
			Body struct {
				Purpose     string `json:"purpose" enum:"product_image,car_image,review_image,banner,brand_logo,kyc_document,delivery_proof,return_photo,ticket_attachment,car_document,dispute_evidence"`
				ContentType string `json:"contentType"`
				SizeBytes   int64  `json:"sizeBytes" minimum:"1"`
				Name        string `json:"name,omitempty" maxLength:"200"`
			}
		}) (*struct {
			Body struct {
				FileID    uuid.UUID `json:"fileId"`
				UploadURL string    `json:"uploadUrl"`
				ExpiresAt time.Time `json:"expiresAt"`
			}
		}, error) {
			rl := rules[in.Body.Purpose]
			ok := false
			for _, t := range rl.types {
				if t == in.Body.ContentType {
					ok = true
				}
			}
			if !ok {
				return nil, httpx.Invalid("file_type_not_allowed", fmt.Sprintf("%s files must be one of %s.", in.Body.Purpose, strings.Join(rl.types, ", ")))
			}
			if in.Body.SizeBytes > rl.maxSize {
				return nil, httpx.Invalid("file_too_large", fmt.Sprintf("Maximum size is %d MB.", rl.maxSize>>20))
			}
			p := httpx.MustPrincipal(ctx)
			vis := "public"
			if rl.private {
				vis = "private"
			}
			key := fmt.Sprintf("%s/%s/%s", in.Body.Purpose, time.Now().Format("2006/01"), uuid.NewString())
			var name *string
			if in.Body.Name != "" {
				name = &in.Body.Name
			}
			f, err := s.d.Q.PlatformCreateFile(ctx, store.PlatformCreateFileParams{StorageKey: key, Purpose: in.Body.Purpose, OriginalName: name, ContentType: in.Body.ContentType,
				SizeBytes: in.Body.SizeBytes, Visibility: vis, UploadedBy: &p.UserID})
			if err != nil {
				return nil, err
			}
			ttl := 15 * time.Minute
			u, err := s.d.Storage.PresignPut(ctx, key, in.Body.ContentType, rl.private, ttl)
			out := &struct {
				Body struct {
					FileID    uuid.UUID `json:"fileId"`
					UploadURL string    `json:"uploadUrl"`
					ExpiresAt time.Time `json:"expiresAt"`
				}
			}{}
			out.Body.FileID, out.Body.UploadURL, out.Body.ExpiresAt = f.ID, u, time.Now().Add(ttl)
			return out, err
		})

	httpx.Register(g, httpx.Op{ID: "completeUpload", Method: http.MethodPost, Path: "/api/v1/files/{id}/complete", Tag: "Files", Auth: anyAud, Summary: "Mark an upload as finished"},
		func(ctx context.Context, in *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Checksum string `json:"checksum,omitempty"`
				Width    *int32 `json:"width,omitempty"`
				Height   *int32 `json:"height,omitempty"`
			}
		}) (*struct{ Body store.PlatformFile }, error) {
			p := httpx.MustPrincipal(ctx)
			f, err := s.d.Q.PlatformGetFile(ctx, in.ID)
			if err != nil || f.UploadedBy == nil || *f.UploadedBy != p.UserID {
				return nil, httpx.NotFound("file")
			}
			var cs *string
			if in.Body.Checksum != "" {
				cs = &in.Body.Checksum
			}
			row, err := s.d.Q.PlatformMarkFileReady(ctx, store.PlatformMarkFileReadyParams{ID: in.ID, Checksum: cs, Width: in.Body.Width, Height: in.Body.Height})
			return &struct{ Body store.PlatformFile }{row}, httpx.DB(err, "pending file")
		})

	httpx.Register(g, httpx.Op{ID: "fileUrl", Method: http.MethodGet, Path: "/api/v1/files/{id}/url", Tag: "Files", Auth: anyAud, Summary: "Download URL (private files: owner or staff; 5 minutes; KYC audited)"},
		func(ctx context.Context, in *struct {
			ID uuid.UUID `path:"id"`
		}) (*struct {
			Body struct {
				URL       string    `json:"url"`
				ExpiresAt time.Time `json:"expiresAt"`
			}
		}, error) {
			p := httpx.MustPrincipal(ctx)
			f, err := s.d.Q.PlatformGetFile(ctx, in.ID)
			if err != nil {
				return nil, httpx.NotFound("file")
			}
			out := &struct {
				Body struct {
					URL       string    `json:"url"`
					ExpiresAt time.Time `json:"expiresAt"`
				}
			}{}
			if f.Visibility == "public" {
				out.Body.URL, out.Body.ExpiresAt = s.d.Storage.PublicURL(f.StorageKey), time.Now().Add(365*24*time.Hour)
				return out, nil
			}
			owner := f.UploadedBy != nil && *f.UploadedBy == p.UserID
			if !owner {
				perm := "support.view"
				if f.Purpose == "kyc_document" {
					perm = "vendors.kyc_review"
				}
				okPerm, err := s.d.RBAC.HasPermission(ctx, p.UserID, perm)
				if err != nil || !okPerm || p.Audience != auth.AudStaff {
					return nil, httpx.NotFound("file")
				}
			}
			if f.Purpose == "kyc_document" && !owner {
				if err := s.d.Runner.Run(ctx, func(tx *uow.Tx) error { return tx.Audit("file.kyc_viewed", "file", f.ID.String(), nil) }); err != nil {
					return nil, err
				}
			}
			u, err := s.d.Storage.PresignGet(ctx, f.StorageKey, true, 5*time.Minute)
			out.Body.URL, out.Body.ExpiresAt = u, time.Now().Add(5*time.Minute)
			return out, err
		})

	// Development: the fake storage provider accepts uploads and serves placeholders.
	if s.d.Cfg.IsDevelopment() {
		r.Gin.PUT("/dev/uploads/*key", func(c *gin.Context) { c.Status(http.StatusOK) })
		r.Gin.GET("/dev/files/*key", func(c *gin.Context) { c.String(http.StatusOK, "placeholder for "+c.Param("key")) })
	}
}
