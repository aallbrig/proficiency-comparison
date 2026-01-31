package database

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"
)

// RawFile represents a downloaded file
type RawFile struct {
	ID          int64
	SourceName  string
	FileURL     string
	FilePath    string
	FileType    string
	ContentHash string
	DownloadedAt time.Time
	FileSize    int64
	Parsed      bool
	ParsedAt    *time.Time
	ParseError  *string
}

// SaveRawFile saves metadata about a downloaded file
func SaveRawFile(db *sql.DB, sourceName, fileURL, filePath, fileType string, fileSize int64, contentHash string) error {
	_, err := db.Exec(`
		INSERT INTO raw_files (source_name, file_url, file_path, file_type, content_hash, file_size, downloaded_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(source_name, file_url) DO UPDATE SET
			file_path = excluded.file_path,
			content_hash = excluded.content_hash,
			file_size = excluded.file_size,
			downloaded_at = CURRENT_TIMESTAMP,
			parsed = 0,
			parsed_at = NULL,
			parse_error = NULL
	`, sourceName, fileURL, filePath, fileType, contentHash, fileSize)
	
	return err
}

// GetUnparsedFiles returns files that haven't been parsed yet
func GetUnparsedFiles(db *sql.DB, sourceName string) ([]RawFile, error) {
	rows, err := db.Query(`
		SELECT id, source_name, file_url, file_path, file_type, content_hash, 
		       downloaded_at, file_size, parsed
		FROM raw_files
		WHERE source_name = ? AND parsed = 0
		ORDER BY downloaded_at DESC
	`, sourceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []RawFile
	for rows.Next() {
		var f RawFile
		err := rows.Scan(&f.ID, &f.SourceName, &f.FileURL, &f.FilePath, &f.FileType, 
			&f.ContentHash, &f.DownloadedAt, &f.FileSize, &f.Parsed)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, rows.Err()
}

// MarkFileParsed marks a file as successfully parsed
func MarkFileParsed(db *sql.DB, fileID int64) error {
	_, err := db.Exec(`
		UPDATE raw_files 
		SET parsed = 1, parsed_at = CURRENT_TIMESTAMP, parse_error = NULL
		WHERE id = ?
	`, fileID)
	return err
}

// MarkFileParseError records a parse error
func MarkFileParseError(db *sql.DB, fileID int64, parseError string) error {
	_, err := db.Exec(`
		UPDATE raw_files 
		SET parse_error = ?
		WHERE id = ?
	`, parseError, fileID)
	return err
}

// FileExists checks if a file with the same hash already exists
func FileExists(db *sql.DB, sourceName, fileURL, contentHash string) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM raw_files 
		WHERE source_name = ? AND file_url = ? AND content_hash = ?
	`, sourceName, fileURL, contentHash).Scan(&count)
	
	return count > 0, err
}

// ComputeFileHash computes SHA256 hash of a file
func ComputeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
