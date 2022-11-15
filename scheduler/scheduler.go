package scheduler

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/errhandler"
	"github.com/vmihailenco/msgpack/v5"
)

type handlerData struct {
	fn             reflect.Value
	typePtrFactory func() any
}

func (h handlerData) run(ctx context.Context, b []byte) error {
	valPtr := h.typePtrFactory()
	err := msgpack.Unmarshal(b, valPtr)
	if err != nil {
		return err
	}
	values := h.fn.Call([]reflect.Value{
		reflect.ValueOf(ctx), reflect.ValueOf(valPtr).Elem()})
	iFace := values[0].Interface()
	if iFace == nil {
		// Special case because it is nil and not (error)(nil).
		return nil
	}
	return iFace.(error)
}

var (
	jobs     = map[string]handlerData{}
	jobsLock = sync.RWMutex{}
)

// Job is used to define a schedule job handler. Should not be created manually,
// use NewJob instead.
type Job[T any] struct {
	jobName string
}

// NewJob is used to create a new scheduler job type. The returned struct
// can be used to interact with it.
func NewJob[T any](name string, hn func(context.Context, T) error) Job[T] {
	typePtrFactory := func() any {
		var val T
		return &val
	}
	reflectFn := reflect.ValueOf(hn)
	jobsLock.Lock()
	jobs[name] = handlerData{
		fn:             reflectFn,
		typePtrFactory: typePtrFactory,
	}
	jobsLock.Unlock()
	return Job[T]{jobName: name}
}

type doNotRerun struct {
	err error
}

func (d doNotRerun) Error() string {
	return d.err.Error()
}

// DoNotRerun is a special wrapper for the error so it knows not to be reran.
func DoNotRerun(err error) error {
	return doNotRerun{err}
}

func isNoRerun(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	{
		x, ok := err.(doNotRerun)
		if ok {
			return true, x.err
		}
	}
	x, ok := err.(*doNotRerun)
	if ok {
		return true, x.err
	}
	return false, err
}

// RunAndBlock is used to run the job right away and block until it is done.
// This does not put it into the scheduler queue and is mainly used for config testing.
func (j Job[T]) RunAndBlock(ctx context.Context, body T) error {
	jobsLock.RLock()
	hn := jobs[j.jobName]
	jobsLock.RUnlock()

	b, err := msgpack.Marshal(body)
	if err != nil {
		return err
	}

	_, err = isNoRerun(hn.run(ctx, b))
	return err
}

func recoverCondom(hn func() error) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			// No panic.
			return
		}
		var ok bool
		if err, ok = r.(error); !ok {
			// Turn it into a fmt type.
			err = fmt.Errorf("%v", r)
		}
	}()
	return hn()
}

// Metadata is used to define the metadata relating to a scheduler.
type Metadata struct {
	// Retries is the number of times this task will be retried with each run decrementing 1.
	// A start of 0 here means unlimited.
	Retries uint

	// Timeout is used to define the timeout of the job.
	Timeout time.Duration

	// RefireDuration is the duration that should be slept if this is refired.
	RefireDuration time.Duration
}

// A job dispatching worker to handle dispatching the job and any additional processing
// that comes from this.
func dispatchWorker(task *db.Task) {
	// Get the handler.
	jobsLock.RLock()
	hn := jobs[task.JobHandler]
	jobsLock.RUnlock()

	// Unmarshal the scheduler metadata.
	var metadata Metadata
	err := msgpack.Unmarshal(task.SchedulerMetadata, &metadata)
	if err == nil {
		// Run the function.
		err = recoverCondom(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), metadata.Timeout)
			defer cancel()
			return hn.run(ctx, task.JobBody)
		})
	}

	// Handle reprocessing if error is not nil.
	if err != nil {
		// Handle checking if we should run this again.
		var noRerun bool
		noRerun, err = isNoRerun(err)
		noRerun = noRerun || metadata.Retries == 1

		// Process the error.
		noRerunStr := "false"
		if noRerun {
			noRerunStr = "true"
		}
		errhandler.Process(err, "scheduler", map[string]string{
			"task_id":     task.TaskID,
			"job_handler": task.JobHandler,
			"no_rerun":    noRerunStr,
		})

		if noRerun {
			// We have exhausted all retries. Return here.
			return
		}

		// Subtract 1 from the retries if not 0.
		if metadata.Retries != 0 {
			metadata.Retries--
		}

		// Schedule this again in the DB.
		metadataPacked, err := msgpack.Marshal(metadata)
		if err != nil {
			// Give up now.
			return
		}
		newTask := &db.Task{
			TaskID:            task.TaskID,
			JobHandler:        task.JobHandler,
			JobBody:           task.JobBody,
			SchedulerMetadata: metadataPacked,
		}
		insert := func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			_, err := db.InternallyAddTask(ctx, newTask, metadata.RefireDuration)
			if err != nil {
				// This isn't good. Blow up.
				_, _ = fmt.Fprintln(os.Stderr, "[scheduler] Failed to reschedule task ID:", newTask.TaskID,
					"job handler:", newTask.JobHandler, "because of the error", err.Error(), "- will keep retrying!")
				return false
			}
			return true
		}
		go func() {
			for {
				// Try inserting.
				if ok := insert(); ok {
					return
				}

				// Sleep for a second.
				time.Sleep(time.Second)
			}
		}()
	}
}

func tasksTimeout(supportedJobs []string, n uint64) ([]*db.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return db.DeleteAndReturnNTasks(ctx, supportedJobs, n)
}

// A queue worker goroutine to work through the scheduler queue and process everything.
func queueWorker() {
	// Get the available jobs.
	jobsLock.RLock()
	jobTypes := make([]string, len(jobs))
	i := 0
	for k := range jobs {
		jobTypes[i] = k
		i++
	}
	jobsLock.RUnlock()

	// Start a job processing loop.
	for {
		// Get the config.
		c := config.Config()

		// Get the number of ms we should sleep for and go to sleep.
		ms := c.SchedulerSleepTime
		if ms == 0 {
			// Default to 1000.
			ms = 1000
		}
		timeDuration := time.Duration(ms) * time.Millisecond
		time.Sleep(timeDuration)

		// Get all the tasks we should process.
		n := uint64(c.JobCount)
		if n == 0 {
			// Default to 120 jobs.
			n = 120
		}
		queue, err := tasksTimeout(jobTypes, n)
		if err != nil {
			// Log that the job fetching failed and then continue.
			_, _ = fmt.Fprintln(os.Stderr, "[scheduler] Failed to get queue:", err.Error())
			continue
		}
		if len(queue) != 0 {
			// Log that we are doing something.
			fmt.Println("[scheduler] Performing", len(queue), "task(s)")
		}

		// Spawn a goroutine to process each one.
		for _, v := range queue {
			go dispatchWorker(v)
		}
	}
}

// InitScheduler is used to initialise the scheduler. This should only be called once.
func InitScheduler() {
	// Start the queue worker in a goroutine.
	go queueWorker()

	// Log that we did that.
	fmt.Println("[scheduler] Scheduling goroutine started.")
}

// ScheduleImmediately is used to add a job to the scheduler with the first fire to be as soon
// as it is picked up by the scheduler. Returns the task ID and error.
func (j Job[T]) ScheduleImmediately(ctx context.Context, body T, metadata Metadata) (string, error) {
	bodyPacked, err := msgpack.Marshal(body)
	if err != nil {
		return "", err
	}

	metadataPacked, err := msgpack.Marshal(metadata)
	if err != nil {
		return "", err
	}

	task := &db.Task{
		JobHandler:        j.jobName,
		JobBody:           bodyPacked,
		SchedulerMetadata: metadataPacked,
	}
	return db.InternallyAddTask(ctx, task, 0)
}

// Schedule is used to add a job to the scheduler with the first fire either in the duration specified or if that is
// zero when the refire duration is.
func (j Job[T]) Schedule(ctx context.Context, body T, metadata Metadata, in time.Duration) (string, error) {
	if in == 0 {
		in = metadata.RefireDuration
	}

	bodyPacked, err := msgpack.Marshal(body)
	if err != nil {
		return "", err
	}

	metadataPacked, err := msgpack.Marshal(metadata)
	if err != nil {
		return "", err
	}

	task := &db.Task{
		JobHandler:        j.jobName,
		JobBody:           bodyPacked,
		SchedulerMetadata: metadataPacked,
	}
	return db.InternallyAddTask(ctx, task, in)
}
