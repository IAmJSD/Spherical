package jobs

import (
	"context"

	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/scheduler"
)

var TokenDeleteJob = scheduler.NewJob("token_delete_job", func(ctx context.Context, token string) error {
	taskId, err := db.InternallyDeleteToken(ctx, token)
	if err != nil {
		return err
	}
	_, err = db.DeleteTasks(ctx, taskId)
	return err
})
