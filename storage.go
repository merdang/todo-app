package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateTask(*Task) error
	DeleteTask(int) error
	UpdateTask(int) error
	GetTasks() ([]*Task, error)
	GetTaskByID(int) (*Task, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	//connStr := "user=postgres dbname=postgres password=todopass sslmode=disable"
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateTaskTable()
}

func (s *PostgresStore) CreateTaskTable() error {
	query := `create table if not exists task (
		id serial primary key,
		task varchar(255),
		status boolean default false,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateTask(t *Task) error {
	query := `insert into task (task, created_at)
	values ($1,$2)`

	resp, err := s.db.Query(query, t.TaskDesc, t.CreatedAt)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgresStore) DeleteTask(id int) error {
	query := `delete from task where id = $1`

	_, err := s.db.Query(query, id)
	return err
}

func (s *PostgresStore) UpdateTask(id int) error {
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	v := !task.Status
	query := `update task
	set status = $2
	where id = $1`

	_, err = s.db.Query(query, id, v)
	return err
}

func (s *PostgresStore) GetTasks() ([]*Task, error) {
	rows, err := s.db.Query("select * from task")
	if err != nil {
		return nil, err
	}

	tasks := []*Task{}
	for rows.Next() {
		task, err := scanIntoTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *PostgresStore) GetTaskByID(id int) (*Task, error) {
	rows, err := s.db.Query("select * from task where id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoTask(rows)
	}

	return nil, fmt.Errorf("task %d not found", id)
}

func scanIntoTask(rows *sql.Rows) (*Task, error) {
	task := new(Task)
	err := rows.Scan(
		&task.ID,
		&task.TaskDesc,
		&task.Status,
		&task.CreatedAt)
	return task, err
}

//func (s *PostgresStore) GetTask(id int) (*Task, error) {
//	return nil, nil
//}
