package jobs

import (
	"context"
	"time"

	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/scheduler"
)

type TokenExtendEvent struct {
	Token string
	In    time.Duration
}

var TokenExtendJob = scheduler.NewJob("token_extend_job", func(ctx context.Context, e TokenExtendEvent) error {
	taskId, err := db.InternallyGetTokenJobID(ctx, e.Token)
	if err != nil {
		return err
	}
	_, err = db.ExtendTask(ctx, taskId, e.In)
	return err
})
