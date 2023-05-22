/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //
**/

package pool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func demoFunc() {

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Hello World!")
	wg.Done()
}

func demoFunc2(i interface{}) {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!", i.(int))
	wg.Done()
}

func TestPool(t *testing.T) {
	p, err := NewPool(50)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)

		err := p.Submit(demoFunc)
		if err != nil {
			t.Error(err)
			return
		}
	}

	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks.\n")
}

func TestPoolWithFunc(t *testing.T) {
	p, err := NewPoolWithFunc(10, func(i interface{}) {
		demoFunc2(i)
	})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		err := p.Invoke(i)
		if err != nil {
			t.Error(err)
			return
		}
		wg.Add(1)

	}
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks.\n")
}

func TestPurge(t *testing.T) {
	p, err := NewPool(50)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)

		err := p.Submit(demoFunc)
		if err != nil {
			t.Error(err)
			return
		}
	}
	wg.Add(1)
	wg.Wait()
	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks.\n")
}
