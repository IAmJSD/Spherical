package db

import (
	"context"
	"time"
)

// Task is used to define an item in the database tasks table.
type Task struct {
	// TaskID is the ID of the task. This is ignored by InternallyAddTask.
	TaskID string

	// JobHandler is the job handler in charge of running this task.
	JobHandler string

	// JobBody is the msgpack body of the job.
	JobBody []byte

	// SchedulerMetadata is the msgpack metadata for the scheduler.
	SchedulerMetadata []byte
}

// DeleteAndReturnNTasks is used to delete N number of tasks from the DB which
// are ready to be processed with the supported jobs of this worker and then
// deletes them from the database.
func DeleteAndReturnNTasks(ctx context.Context, supportedJobs []string, n uint64) ([]*Task, error) {
	query := `DELETE FROM tasks
		WHERE task_id = ANY(ARRAY(SELECT task_id FROM tasks
		WHERE NOW() >= scheduled_for AND job_handler = ANY($1) LIMIT $2)) RETURNING task_id,
		job_handler, job_body, scheduler_metadata`
	rows, err := dbConn().Query(ctx, query, supportedJobs, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []*Task{}
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.TaskID, &task.JobHandler, &task.JobBody, &task.SchedulerMetadata)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// InternallyAddTask is used to add a task in a very low level way to the database.
// You likely want the scheduler package for this. The task ID is returned.
func InternallyAddTask(ctx context.Context, task *Task, in time.Duration) (string, error) {
	query := "INSERT INTO tasks (job_handler, job_body, scheduler_metadata, scheduled_for)" +
		" VALUES ($1, $2, $3, NOW() + $4) RETURNING task_id"
	var taskId string
	err := dbConn().
		QueryRow(ctx, query, task.JobHandler, task.JobBody, task.SchedulerMetadata, in).
		Scan(&taskId)
	return taskId, err
}

// DeleteTasks is used to delete task(s) from the database. Returns true for deleted if they all were.
func DeleteTasks(ctx context.Context, taskIds ...string) (deleted bool, err error) {
	query := "DELETE FROM tasks WHERE task_id = ANY($1)"
	res, err := dbConn().Exec(ctx, query, taskIds)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() == int64(len(taskIds)), nil
}

// ExtendTask is used to extend the scheduled time of a task.
func ExtendTask(ctx context.Context, taskId string, in time.Duration) (extended bool, err error) {
	query := "UPDATE tasks SET scheduled_for = NOW() + $2 WHERE task_id = $1"
	res, err := dbConn().Exec(ctx, query, taskId, in)
	if err != nil {
		return false, err
	}
	return res.RowsAffected() == 1, nil
}
