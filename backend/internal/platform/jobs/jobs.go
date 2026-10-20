// Package jobs configures River (Postgres-backed job queue, schema "river"). Modules register
// their workers and periodic jobs; the worker process (or the API in local dev) runs them.
package jobs

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// Queues: critical messages never wait behind marketing (docs/messaging-marketing.md §1).
const (
	QueueDefault       = river.QueueDefault
	QueueCritical      = "critical"
	QueueTransactional = "transactional"
	QueueMarketing     = "marketing"
	QueueFinance       = "finance"
)

// Registry collects workers and periodic jobs from modules.
type Registry struct {
	Workers  *river.Workers
	Periodic []*river.PeriodicJob
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry { return &Registry{Workers: river.NewWorkers()} }

// Every adds a periodic job that runs at a fixed interval.
func (r *Registry) Every(d time.Duration, build func() (river.JobArgs, *river.InsertOpts)) {
	r.Periodic = append(r.Periodic, river.NewPeriodicJob(river.PeriodicInterval(d), build, &river.PeriodicJobOpts{RunOnStart: false}))
}

// NewClient builds a River client. With work=false it only inserts jobs (the API in production).
func NewClient(pool *pgxpool.Pool, reg *Registry, work bool, concurrency int, logger *slog.Logger) (*river.Client[pgx.Tx], error) {
	cfg := &river.Config{Schema: "river", Logger: logger, Workers: reg.Workers}
	if work {
		cfg.Queues = map[string]river.QueueConfig{
			QueueDefault:       {MaxWorkers: concurrency},
			QueueCritical:      {MaxWorkers: 5},
			QueueTransactional: {MaxWorkers: 5},
			QueueMarketing:     {MaxWorkers: 2},
			QueueFinance:       {MaxWorkers: 2},
		}
		cfg.PeriodicJobs = reg.Periodic
	}
	return river.NewClient(riverpgxv5.New(pool), cfg)
}
