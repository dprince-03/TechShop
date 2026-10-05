// Package database groups reusable database clients, one sub-package per engine:
//
//	postgres/ – PostgreSQL pools (pgx), used today
//	redis/    – planned: caching and fast counters (allowed since 2026-10-06)
//	mongo/    – planned
//
// Each client takes plain Options and returns the driver's handle, so services wire them
// from their own config. Nothing here imports TechShop's internal packages.
package database
