/**
  @author: Bruce
  @since: 2023/5/19
  @desc: //TODO
**/

package utils

import (
	"fmt"
	"testing"
)

func TestSnowflakeGenerator_NextVal(t *testing.T) {
	sf := NewSnowflakeGenerator(1, 1)
	for i := 0; i < 10; i++ {
		id, err := sf.NextVal()
		if err != nil {
			fmt.Println("error: ", err)
		}

		fmt.Println("Id: ", id)
	}

}
