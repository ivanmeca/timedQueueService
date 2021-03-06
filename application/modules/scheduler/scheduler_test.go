package scheduler

import (
	"context"
	"fmt"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestSchedulerPoll(test *testing.T) {
	RegisterTestingT(test)
	fmt.Println("Trying to poll database")
	scheduler := NewScheduler(2, 1, 1800)
	scheduler.Run(context.Background())
	time.Sleep(20000 * time.Second)
}
