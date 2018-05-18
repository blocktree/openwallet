package timer

import (
	"fmt"
	"testing"
	"time"
)


func now() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05:000"))
}

func TestTimer(t *testing.T) {
	timer := newTask(time.Second*1, now)
	timer1 := newTask(time.Second*2, now)
	timer.Start()
	timer1.Start()

	time.Sleep(time.Second * 10)

	timer.Pause()
	time.Sleep(time.Second * 5)
	timer.Restart()
	time.Sleep(time.Second * 5)
	timer.Stop()

}
