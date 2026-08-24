package store

import (
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Asset struct {
	ID               int64      `json:"id"`
	Filename         string     `json:"filename"`
	Format           string     `json:"format"`
	SizeBytes        int64      `json:"size_bytes"`
	StoragePath      string     `json:"-"`
	PreviewPath      string     `json:"-"`
	ExtractionMode   string     `json:"extraction_mode"`
	CameraMake       string     `json:"camera_make"`
	CameraModel      string     `json:"camera_model"`
	LensModel        string     `json:"lens_model"`
	LensSpec         string     `json:"lens_spec"`
	Aperture         float64    `json:"aperture"`
	ShutterText      string     `json:"shutter_text"`
	ShutterSeconds   float64    `json:"shutter_seconds"`
	ISO              int        `json:"iso"`
	FocalLength      float64    `json:"focal_length"`
	FocalLength35mm  float64    `json:"focal_length_35mm"`
	DateTimeOriginal time.Time  `json:"datetime_original"`
	Orientation      int        `json:"orientation"`
	WhiteBalance     string     `json:"white_balance"`
	ExposureBias     float64    `json:"exposure_bias"`
	Rating           int        `json:"rating"`
	Tags             []string   `json:"tags"`
	Sharpness        *float64   `json:"sharpness"`
	Noise            *float64   `json:"noise"`
	ClipShadow       *float64   `json:"clip_shadow"`
	ClipHighlight    *float64   `json:"clip_highlight"`
	EVDeviation      *float64   `json:"ev_deviation"`
	TileStatus       string     `json:"tile_status"`
	TileMaxZ         int        `json:"tile_max_z"`
	Width            int        `json:"width"`
	Height           int        `json:"height"`
	ExifRaw          []byte     `json:"exif_raw"`
	DeletedAt        *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Job struct {
	ID        int64
	AssetID   int64
	Kind      string
	Status    string
	Attempts  int
	LastError string
}

type AssetFilter struct {
	Q           string
	Camera      string
	Lens        string
	ISOMin      int
	ISOMax      int
	FocalMin    float64
	FocalMax    float64
	ApertureMin float64
	ApertureMax float64
	From        time.Time
	To          time.Time
	Sort        string
	Page        int
	PageSize    int
}

type JobStats struct {
	Queued    int
	Running   int
	Succeeded int
	Failed    int
}
