package wms

import (
	"github.com/dprince-03/techshop/backend/internal/store"
)

// PickListView is a pick list with its items (ordered by bin for a short walk).
type PickListView struct {
	store.FulfilmentPickList
	Items []store.FulfilmentPickItemsRow `json:"items"`
}

// ManifestView is a manifest with its parcels.
type ManifestView struct {
	store.FulfilmentManifest
	Parcels []store.FulfilmentManifestParcelsRow `json:"parcels"`
}
