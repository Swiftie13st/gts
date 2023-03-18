/**
  @author: Bruce
  @since: 2023/3/17
  @desc:
**/

package main

import (
	"gts/server"
	"gts/utils"
)

func main() {
	utils.InitSettings()

	s := server.NewServer()
	s.Serve()
}
