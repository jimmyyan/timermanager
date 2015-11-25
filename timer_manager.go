package timermanager

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	// "github.com/golang/glog"
)

type TimerManager interface {
	AddPeriodTimer(interval int64, f TimerFunc) (uint64, error)

	AddOneshotTimer(interval int64, f TimerFunc) (uint64, error)

	RemoveTimer(timer_id uint64) error

	Stop()

	IsStopped() bool

	Clear()

	TimerNum() int
}

type timerManager struct {
	m          sync.Mutex
	timers     []*timer
	stopChan   chan bool
	is_stopped bool
	wg         sync.WaitGroup
	atomic_id  uint64
}

type TimerFunc func()

type timer struct {
	t         *time.Timer
	is_period bool
	timer_id  uint64
	interval  time.Duration
	stopChan  chan bool
	task      TimerFunc
}

func (tm *timerManager) RemoveTimer(timer_id uint64) error {
	tm.m.Lock()
	defer tm.m.Unlock()

	if tm.timers == nil {
		return fmt.Errorf("No timer in timer manager")
	}

	if index, err := tm.findTimer(timer_id); err != nil {
		return err
	} else {
		tm.timers[index].stopChan <- true
		tm.timers = append(tm.timers[0:index], tm.timers[(index+1):]...)
		return nil
	}
}

func (tm *timerManager) findTimerWithLock(timer_id uint64) (int, error) {
	tm.m.Lock()
	defer tm.m.Unlock()

	return tm.findTimer(timer_id)
}

func (tm *timerManager) findTimer(timer_id uint64) (int, error) {
	for index, t := range tm.timers {
		if t.timer_id == timer_id {
			return index, nil
		}
	}

	return -1, fmt.Errorf("Not found timer %d", timer_id)
}

func (tm *timerManager) AddPeriodTimer(interval int64, f TimerFunc) (uint64, error) {
	return tm.addTimerInternal(interval, f, true)
}

func (tm *timerManager) AddOneshotTimer(interval int64, f TimerFunc) (uint64, error) {
	return tm.addTimerInternal(interval, f, false)
}

func (tm *timerManager) addTimerInternal(interval int64, f TimerFunc, period bool) (uint64, error) {
	if tm.IsStopped() {
		return 0, fmt.Errorf("Timer Manager is stopped")
	}

	pt := &timer{t: time.NewTimer(time.Duration(interval) * time.Millisecond),
		is_period: period,
		timer_id:  tm.createTimerId(),
		interval:  time.Duration(interval) * time.Millisecond,
		stopChan:  make(chan bool),
		task:      f,
	}
	tm.wg.Add(1)
	go func() {
		for {
			select {
			case <-tm.stopChan:
				tm.wg.Done()
				return
			case <-pt.t.C:
				pt.task()
				if pt.is_period {
					pt.t.Reset(time.Duration(pt.interval))
				} else {
					tm.wg.Done()
					if index, err := tm.findTimerWithLock(pt.timer_id); err != nil {
						fmt.Printf("Failed to remove oneshot timer, %d", pt.timer_id)
					} else {
						tm.m.Lock()
						defer tm.m.Unlock()
						tm.timers = append(tm.timers[0:index], tm.timers[(index+1):]...)
					}
					return
				}
			case <-pt.stopChan:
				tm.wg.Done()
				return
			}
		}
	}()

	if tm.timers == nil {
		tm.timers = make([]*timer, 0, 1000)
	}
	tm.timers = append(tm.timers, pt)
	return pt.timer_id, nil
}

func (tm *timerManager) createTimerId() uint64 {
	return atomic.AddUint64(&tm.atomic_id, 1)
}

func (tm *timerManager) Stop() {
	tm.m.Lock()
	defer tm.m.Unlock()
	if tm.is_stopped {
		return
	}

	tm.stopAllTimers()
	tm.is_stopped = true
}

func (tm *timerManager) stopAllTimers() {
	for _, t := range tm.timers {
		close(t.stopChan)
	}
	tm.wg.Wait()
}

func (tm *timerManager) IsStopped() bool {
	tm.m.Lock()
	defer tm.m.Unlock()
	return tm.is_stopped
}

func NewTimerManager() TimerManager {
	tm := &timerManager{timers: make([]*timer, 0),
		stopChan: make(chan bool)}
	return tm
}

func (tm *timerManager) TimerNum() int {
	tm.m.Lock()
	defer tm.m.Unlock()
	return len(tm.timers)
}

func (tm *timerManager) Clear() {
	tm.m.Lock()
	defer tm.m.Unlock()
	tm.stopAllTimers()
	tm.timers = nil
}
