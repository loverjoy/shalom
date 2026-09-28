package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"shalom/internal/models"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, f *models.FileUpload) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO file_uploads (id, uploader_id, file_name, file_type, file_size, mime_type, storage_key, url, is_public, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		 RETURNING id, created_at`,
		f.ID, f.UploaderID, f.FileName, f.FileType, f.FileSize, f.MimeType, f.StorageKey, f.URL, f.IsPublic,
	).Scan(&f.ID, &f.CreatedAt)
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.FileUpload, error) {
	f := &models.FileUpload{}
	err := r.db.QueryRow(ctx,
		`SELECT id, uploader_id, file_name, file_type, file_size, mime_type, storage_key, url, is_public, created_at
		 FROM file_uploads WHERE id = $1`, id,
	).Scan(&f.ID, &f.UploaderID, &f.FileName, &f.FileType, &f.FileSize, &f.MimeType, &f.StorageKey, &f.URL, &f.IsPublic, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FileRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.FileUpload, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, uploader_id, file_name, file_type, file_size, mime_type, storage_key, url, is_public, created_at
		 FROM file_uploads WHERE uploader_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.FileUpload
	for rows.Next() {
		var f models.FileUpload
		rows.Scan(&f.ID, &f.UploaderID, &f.FileName, &f.FileType, &f.FileSize, &f.MimeType, &f.StorageKey, &f.URL, &f.IsPublic, &f.CreatedAt)
		files = append(files, f)
	}
	return files, nil
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM file_uploads WHERE id = $1`, id)
	return err
}

func (r *FileRepository) GetUserStorageUsed(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(file_size), 0) FROM file_uploads WHERE uploader_id = $1`, userID,
	).Scan(&total)
	return total, err
}
