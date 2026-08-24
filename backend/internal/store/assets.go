package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"colorpixel/internal/timeutil"
)

func (d *DB) InsertAsset(ctx context.Context, a *Asset) error {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = timeutil.Now()
	}
	a.UpdatedAt = a.CreatedAt
	if a.Tags == nil {
		a.Tags = []string{}
	}
	if len(a.ExifRaw) == 0 {
		a.ExifRaw = []byte("{}")
	}
	return d.Pool.QueryRow(ctx, `
INSERT INTO assets (
  filename, format, size_bytes, storage_path, preview_path, extraction_mode,
  camera_make, camera_model, lens_model, lens_spec, aperture, shutter_text, shutter_seconds,
  iso, focal_length, focal_length_35mm, datetime_original, orientation, white_balance, exposure_bias,
  rating, tags, sharpness, noise, clip_shadow, clip_highlight, ev_deviation,
  tile_status, tile_max_z, width, height, exif_raw, created_at, updated_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
  $21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34
) RETURNING id`,
		a.Filename, a.Format, a.SizeBytes, a.StoragePath, nullStr(a.PreviewPath), a.ExtractionMode,
		a.CameraMake, a.CameraModel, a.LensModel, a.LensSpec, a.Aperture, a.ShutterText, a.ShutterSeconds,
		a.ISO, a.FocalLength, a.FocalLength35mm, nullTime(a.DateTimeOriginal), a.Orientation, a.WhiteBalance, a.ExposureBias,
		a.Rating, a.Tags, a.Sharpness, a.Noise, a.ClipShadow, a.ClipHighlight, a.EVDeviation,
		a.TileStatus, a.TileMaxZ, a.Width, a.Height, a.ExifRaw, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
}

func (d *DB) GetAsset(ctx context.Context, id int64) (*Asset, error) {
	row := d.Pool.QueryRow(ctx, assetSelect+` WHERE id=$1 AND deleted_at IS NULL`, id)
	a, err := scanAsset(row)
	if errorsIsNoRows(err) {
		return nil, ErrNotFound
	}
	return a, err
}

func (d *DB) SoftDelete(ctx context.Context, id int64) error {
	tag, err := d.Pool.Exec(ctx, `UPDATE assets SET deleted_at=$2, updated_at=$2 WHERE id=$1 AND deleted_at IS NULL`, id, timeutil.Now())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *DB) UpdateRating(ctx context.Context, id int64, rating int, tags []string) error {
	if tags == nil {
		tags = []string{}
	}
	tag, err := d.Pool.Exec(ctx, `UPDATE assets SET rating=$2, tags=$3, updated_at=$4 WHERE id=$1 AND deleted_at IS NULL`,
		id, rating, tags, timeutil.Now())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (d *DB) UpdateTiles(ctx context.Context, id int64, status string, maxZ, w, h int, sharpness, noise, clipS, clipH, ev *float64) error {
	_, err := d.Pool.Exec(ctx, `
UPDATE assets SET tile_status=$2, tile_max_z=$3, width=COALESCE(NULLIF($4,0), width),
 height=COALESCE(NULLIF($5,0), height), sharpness=$6, noise=$7, clip_shadow=$8, clip_highlight=$9,
 ev_deviation=$10, updated_at=$11 WHERE id=$1`,
		id, status, maxZ, w, h, sharpness, noise, clipS, clipH, ev, timeutil.Now())
	return err
}

func (d *DB) CountAlive(ctx context.Context) (int, error) {
	var n int
	err := d.Pool.QueryRow(ctx, `SELECT count(*) FROM assets WHERE deleted_at IS NULL`).Scan(&n)
	return n, err
}

const assetSelect = `
SELECT id, filename, format, size_bytes, storage_path, COALESCE(preview_path,''), extraction_mode,
  COALESCE(camera_make,''), COALESCE(camera_model,''), COALESCE(lens_model,''), COALESCE(lens_spec,''),
  COALESCE(aperture,0), COALESCE(shutter_text,''), COALESCE(shutter_seconds,0),
  COALESCE(iso,0), COALESCE(focal_length,0), COALESCE(focal_length_35mm,0),
  datetime_original, COALESCE(orientation,0), COALESCE(white_balance,''), COALESCE(exposure_bias,0),
  rating, tags, sharpness, noise, clip_shadow, clip_highlight, ev_deviation,
  tile_status, tile_max_z, COALESCE(width,0), COALESCE(height,0), exif_raw, created_at, updated_at
FROM assets`

func (d *DB) ListAssets(ctx context.Context, f AssetFilter) ([]Asset, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 40
	}
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	add := func(clause string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}
	if f.Q != "" {
		args = append(args, "%"+f.Q+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(filename ILIKE $%d OR camera_model ILIKE $%d OR lens_model ILIKE $%d)", n, n, n))
	}
	if f.Camera != "" {
		add("camera_model ILIKE $%d", "%"+f.Camera+"%")
	}
	if f.Lens != "" {
		add("lens_model ILIKE $%d", "%"+f.Lens+"%")
	}
	if f.ISOMin > 0 {
		add("iso >= $%d", f.ISOMin)
	}
	if f.ISOMax > 0 {
		add("iso <= $%d", f.ISOMax)
	}
	if f.FocalMin > 0 {
		add("focal_length_35mm >= $%d", f.FocalMin)
	}
	if f.FocalMax > 0 {
		add("focal_length_35mm <= $%d", f.FocalMax)
	}
	if f.ApertureMin > 0 {
		add("aperture >= $%d", f.ApertureMin)
	}
	if f.ApertureMax > 0 {
		add("aperture <= $%d", f.ApertureMax)
	}
	if !f.From.IsZero() {
		add("datetime_original >= $%d", f.From)
	}
	if !f.To.IsZero() {
		add("datetime_original <= $%d", f.To)
	}
	w := strings.Join(where, " AND ")
	order := "datetime_original DESC NULLS LAST, id DESC"
	switch f.Sort {
	case "iso":
		order = "iso ASC NULLS LAST, id DESC"
	case "focal":
		order = "focal_length_35mm ASC NULLS LAST, id DESC"
	case "created":
		order = "created_at DESC, id DESC"
	}
	var total int
	if err := d.Pool.QueryRow(ctx, `SELECT count(*) FROM assets WHERE `+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	q := assetSelect + ` WHERE ` + w + ` ORDER BY ` + order + fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)-1, len(args))
	rows, err := d.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *a)
	}
	if list == nil {
		list = []Asset{}
	}
	return list, total, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAsset(row scannable) (*Asset, error) {
	var a Asset
	var dt *time.Time
	var tags []string
	var exif []byte
	err := row.Scan(
		&a.ID, &a.Filename, &a.Format, &a.SizeBytes, &a.StoragePath, &a.PreviewPath, &a.ExtractionMode,
		&a.CameraMake, &a.CameraModel, &a.LensModel, &a.LensSpec,
		&a.Aperture, &a.ShutterText, &a.ShutterSeconds,
		&a.ISO, &a.FocalLength, &a.FocalLength35mm,
		&dt, &a.Orientation, &a.WhiteBalance, &a.ExposureBias,
		&a.Rating, &tags, &a.Sharpness, &a.Noise, &a.ClipShadow, &a.ClipHighlight, &a.EVDeviation,
		&a.TileStatus, &a.TileMaxZ, &a.Width, &a.Height, &exif, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if dt != nil {
		a.DateTimeOriginal = *dt
	}
	if tags == nil {
		tags = []string{}
	}
	a.Tags = tags
	if len(exif) == 0 {
		exif = []byte("{}")
	}
	a.ExifRaw = json.RawMessage(exif)
	return &a, nil
}

func errorsIsNoRows(err error) bool {
	return err != nil && (err == pgx.ErrNoRows)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (d *DB) LensRows(ctx context.Context, from, to time.Time) ([]Asset, error) {
	rows, err := d.Pool.Query(ctx, assetSelect+`
 WHERE deleted_at IS NULL AND datetime_original >= $1 AND datetime_original < $2`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Asset
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *a)
	}
	return list, rows.Err()
}
