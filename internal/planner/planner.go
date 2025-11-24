package planner

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

// Task represents a single unit of work
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"` // "pending", "completed", "in_progress"
}

// Planner manages a list of tasks using SQLite
type Planner struct {
	db *sql.DB
}

// NewPlanner creates a new Planner instance
func NewPlanner(dbPath string) (*Planner, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		status TEXT DEFAULT 'pending'
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Planner{db: db}, nil
}

// AddTask adds a new task to the planner
func (p *Planner) AddTask(title, description string, start, end time.Time) (Task, error) {
	query := `INSERT INTO tasks (title, description, start_time, end_time, status) VALUES (?, ?, ?, ?, ?)`
	res, err := p.db.Exec(query, title, description, start, end, "pending")
	if err != nil {
		return Task{}, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return Task{
		ID:          int(id),
		Title:       title,
		Description: description,
		StartTime:   start,
		EndTime:     end,
		Status:      "pending",
	}, nil
}

// ListTasks returns all tasks
func (p *Planner) ListTasks() ([]Task, error) {
	query := `SELECT id, title, description, start_time, end_time, status FROM tasks ORDER BY start_time ASC`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.StartTime, &t.EndTime, &t.Status); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetTask finds a task by ID
func (p *Planner) GetTask(id int) (Task, error) {
	query := `SELECT id, title, description, start_time, end_time, status FROM tasks WHERE id = ?`
	row := p.db.QueryRow(query, id)

	var t Task
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &t.StartTime, &t.EndTime, &t.Status); err != nil {
		if err == sql.ErrNoRows {
			return Task{}, fmt.Errorf("task with ID %d not found", id)
		}
		return Task{}, fmt.Errorf("failed to scan task: %w", err)
	}
	return t, nil
}

// UpdateTask updates an existing task
func (p *Planner) UpdateTask(t Task) error {
	query := `UPDATE tasks SET title = ?, description = ?, start_time = ?, end_time = ?, status = ? WHERE id = ?`
	res, err := p.db.Exec(query, t.Title, t.Description, t.StartTime, t.EndTime, t.Status, t.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task with ID %d not found", t.ID)
	}
	return nil
}

// DeleteTask deletes a task by ID
func (p *Planner) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	res, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task with ID %d not found", id)
	}
	return nil
}

// ExportToMarkdown exports all tasks to a markdown file
func (p *Planner) ExportToMarkdown(filename string) error {
	tasks, err := p.ListTasks()
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Gomentum Plan\n\n")
	fmt.Fprintf(f, "Generated at: %s\n\n", time.Now().Format(time.RFC1123))

	for _, t := range tasks {
		fmt.Fprintf(f, "## %s\n", t.Title)
		fmt.Fprintf(f, "- **ID**: %d\n", t.ID)
		fmt.Fprintf(f, "- **Time**: %s - %s\n", t.StartTime.Format("15:04"), t.EndTime.Format("15:04"))
		fmt.Fprintf(f, "- **Status**: %s\n", t.Status)
		if t.Description != "" {
			fmt.Fprintf(f, "- **Description**: %s\n", t.Description)
		}
		fmt.Fprintln(f)
	}
	return nil
}

// Close closes the database connection
func (p *Planner) Close() error {
	return p.db.Close()
}
