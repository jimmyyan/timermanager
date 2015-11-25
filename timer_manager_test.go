package timermanager

import (
	"fmt"
	"testing"
	"time"
)

func timer_func() {
	fmt.Println("I am in timer")
}

func timer_func1() {
	fmt.Println("I am in timer1")
}

func TestTimerManager_AddPeriodTimer(t *testing.T) {
	tm := NewTimerManager()

	timer_id, err := tm.AddPeriodTimer(1000, timer_func)

	if timer_id != 1 {
		t.Errorf("timer id is incorrect %d ", timer_id)
	}

	timer_id, err = tm.AddPeriodTimer(500, timer_func1)
	time.Sleep(5 * time.Second)

	fmt.Println("len: ", tm.TimerNum())
	err = tm.RemoveTimer(timer_id)
	if err != nil {
		t.Fatalf("err : %v", err)
	}

	tm.Stop()

	_, err = tm.AddPeriodTimer(1000, timer_func)
	if err == nil {
		t.Fatalf("not expect to start a period timer when timer manager is stoppped")
	}
}

func TestTimerManager_AddOneshotTimer(t *testing.T) {
	tm := NewTimerManager()
	timer_id, err := tm.AddOneshotTimer(2000, timer_func)

	if err != nil || timer_id != 1 {
		t.Fatalf("Failed to add oneshot timer")
	}

	if tm.TimerNum() != 1 {
		t.Fatalf("Timer id expect %d actual %d ", 1, tm.TimerNum())
	}

	time.Sleep(4 * time.Second)
	if tm.TimerNum() != 0 {
		t.Fatalf("Timer id expect %d actual %d ", 0, tm.TimerNum())
	}

	timer_id, err = tm.AddPeriodTimer(1000, timer_func)
	if err != nil || timer_id != 2 {
		t.Fatalf("Failed to add period timer")
	}

	time.Sleep(4 * time.Second)
	tm.Clear()
	if tm.TimerNum() != 0 {
		t.Errorf("Timer id expect %d actual %d ", 0, tm.TimerNum())
	}

	fmt.Println("xxxxxxx")
	tm.Stop()
}
