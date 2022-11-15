package jobs

import (
	"context"
	"fmt"

	"github.com/jakemakesstuff/spherical/scheduler"
)

var HelloJob = scheduler.NewJob("hello_job", func(ctx context.Context, name string) error {
	fmt.Println("Hello", name)
	return nil
})
