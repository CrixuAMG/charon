package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type ProjectAccess struct {
	ProjectName string
	AccessedAt  time.Time
	AccessCount int
}

type DB struct {
	conn *sql.DB
}

func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	dbDir := filepath.Join(homeDir, ".config", "charon")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create db directory: %w", err)
	}
	return filepath.Join(dbDir, "charon.db"), nil
}

func Open() (*DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.init(); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) init() error {
	query := `
		CREATE TABLE IF NOT EXISTS project_access (
			project_name TEXT PRIMARY KEY,
			accessed_at DATETIME NOT NULL,
			access_count INTEGER NOT NULL DEFAULT 1
		);
		CREATE INDEX IF NOT EXISTS idx_accessed_at ON project_access(accessed_at DESC);
	`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	return nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

func (db *DB) RecordAccess(projectName string) error {
	query := `
		INSERT INTO project_access (project_name, accessed_at, access_count)
		VALUES (?, ?, 1)
		ON CONFLICT(project_name) DO UPDATE SET
			accessed_at = excluded.accessed_at,
			access_count = access_count + 1
	`
	_, err := db.conn.Exec(query, projectName, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record access: %w", err)
	}
	return nil
}

func (db *DB) GetRecentProjects(limit int) ([]ProjectAccess, error) {
	query := `
		SELECT project_name, accessed_at, access_count
		FROM project_access
		ORDER BY accessed_at DESC
		LIMIT ?
	`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent projects: %w", err)
	}
	defer rows.Close()

	var projects []ProjectAccess
	for rows.Next() {
		var p ProjectAccess
		if err := rows.Scan(&p.ProjectName, &p.AccessedAt, &p.AccessCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, rows.Err()
}

func (db *DB) GetProjectAccessTime(projectName string) (*time.Time, error) {
	query := `SELECT accessed_at FROM project_access WHERE project_name = ?`
	var accessedAt time.Time
	err := db.conn.QueryRow(query, projectName).Scan(&accessedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get access time: %w", err)
	}
	return &accessedAt, nil
}

func (db *DB) GetProjectAccessCount(projectName string) (int, error) {
	query := `SELECT access_count FROM project_access WHERE project_name = ?`
	var count int
	err := db.conn.QueryRow(query, projectName).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get access count: %w", err)
	}
	return count, nil
}

