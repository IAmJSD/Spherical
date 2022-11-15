package tlru

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testObject struct {
	present bool
}

func TestCache_Set_Get(t *testing.T) {
	c := Cache[testObject]{}

	// Get without value existing.
	val := c.Get("hello", time.Millisecond)
	assert.False(t, val.present)

	// Set value.
	c.Set("hello", testObject{present: true}, time.Millisecond)
	val = c.Get("hello", time.Millisecond*5)
	assert.True(t, val.present)

	// Sleep a couple ms and make sure we can still get it to make sure the TTL was reset.
	time.Sleep(time.Millisecond * 2)
	val = c.Get("hello", time.Millisecond)
	assert.True(t, val.present)

	// Sleep a couple milliseconds and make sure it is no longer there.
	time.Sleep(time.Millisecond * 2)
	val = c.Get("hello", time.Millisecond)
	assert.False(t, val.present)
}

func TestCache_Get_Race(t *testing.T) {
	c := Cache[testObject]{}
	c.Set("testing", testObject{present: true}, time.Hour)
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			c.Get("testing", time.Hour)
		}()
	}
	wg.Wait()
}

func TestCache_Set_Race(t *testing.T) {
	c := Cache[testObject]{}
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func(i int) {
			defer wg.Done()
			c.Set(strconv.Itoa(i), testObject{present: true}, time.Hour)
		}(i)
	}
	wg.Wait()
}
