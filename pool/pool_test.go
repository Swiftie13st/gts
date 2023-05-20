/**
  @author: Bruce
  @since: 2023/5/21
  @desc: //
**/

package pool

import (
	"fmt"
	"testing"
	"time"
)

func demoFunc() {
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Hello World!")
}

func TestPool(t *testing.T) {
	p, err := NewPool(50)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100; i++ {
		err := p.Submit(demoFunc)
		if err != nil {
			t.Error(err)
			return
		}
	}

	fmt.Printf("running goroutines: %d\n", p.Running())
	fmt.Printf("finish all tasks.\n")
}
