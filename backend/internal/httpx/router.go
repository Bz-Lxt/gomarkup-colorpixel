package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"colorpixel/internal/config"
	"colorpixel/internal/jobs"
	"colorpixel/internal/store"
)

type Server struct {
	Cfg    config.Config
	DB     *store.DB
	Worker *jobs.Worker
}

func New(cfg config.Config, db *store.DB, worker *jobs.Worker) http.Handler {
	s := &Server{Cfg: cfg, DB: db, Worker: worker}
	r := chi.NewRouter()
	r.Use(requestLog)
	r.Use(cors(cfg.CORSOrigins))
	r.Get("/api/v1/health", s.health)
	r.Post("/api/v1/auth/login", s.login)
	r.Group(func(r chi.Router) {
		r.Use(s.requireAuth)
		r.Post("/api/v1/assets/upload", s.upload)
		r.Get("/api/v1/assets", s.listAssets)
		r.Get("/api/v1/assets/{id}", s.getAsset)
		r.Patch("/api/v1/assets/{id}", s.patchAsset)
		r.Delete("/api/v1/assets/{id}", s.deleteAsset)
		r.Get("/api/v1/assets/{id}/histogram", s.histogram)
		r.Get("/api/v1/reports/golden-lens", s.goldenLens)
	})
	r.Get("/api/v1/assets/{id}/preview", s.preview)
	r.Get("/api/v1/assets/{id}/tiles/{z}/{x}/{y}.jpg", s.tile)
	r.Get("/api/v1/assets/{id}/thumb.jpg", s.thumb)
	return r
}
