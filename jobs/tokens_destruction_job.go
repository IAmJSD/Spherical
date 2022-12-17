package jobs

import (
	"context"

	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/scheduler"
)

var TokensDestructionJob = scheduler.NewJob("tokens_destruction_job", func(ctx context.Context, userId uint64) error {
	tokens, err := db.InternallyDeleteAllUserIDTokens(ctx, userId)
	if err != nil {
		return err
	}
	_, err = db.DeleteTasks(ctx, tokens...)
	return err
})
