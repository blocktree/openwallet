package timer

import (
	"time"
)

type TaskTimer struct {
	f func() //传入方法
	timer *time.Ticker //定时器
	stop bool //停止标记
	pause bool //暂停标记
}

//新建定时器
func NewTask(duration time.Duration,function func()) *TaskTimer{
	t := &TaskTimer{
		timer: time.NewTicker(duration),
		f:function,
		stop: false,
		pause:false,
	}
	return t
}

//启动定时器
func(t *TaskTimer) Start(){
	t.stop = false
	t.pause = false
	go func(innerT *TaskTimer) {
		defer innerT.timer.Stop()
		for {
			select {
			case <-innerT.timer.C:
				if innerT.stop{
					return
				}
				if innerT.pause{
					continue
				}
				innerT.f() //执行我们想要的操作
			}
		}
	}(t)
}

//停止定时器
func (t *TaskTimer) Stop()  {
	t.stop = true
}

//暂停定时器
func (t *TaskTimer) Pause()  {
	t.pause = true
}

//继续定时器
func (t *TaskTimer) Restart()  {
	t.pause = false
}

func (t *TaskTimer) Running() bool {
	if t.stop || t.pause {
		return false
	} else {
		return  true
	}
}