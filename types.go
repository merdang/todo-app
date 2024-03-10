package main

import "time"

type CreateTaskReqest struct {
	TaskDesc string `json:"taskDescription"`
}

type UpdateTaskReq struct {
	Update string `json:"update"`
}

type Task struct {
	ID        int       `json:"id"`
	TaskDesc  string    `json:"taskDescription"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewTask(taskdesc string) *Task {
	return &Task{
		TaskDesc:  taskdesc,
		Status:    false,
		CreatedAt: time.Now().UTC(),
	}
}
