// Response shapes returned by the Go API (backend/internal/handler).

export type HealthResponse = {
  status: "ok" | "degraded";
  database: "up" | "down";
};
