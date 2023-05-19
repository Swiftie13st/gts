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

func TestNewLog(t *testing.T) {
	InitSettings("../conf/config.yaml")
	logger, err := InitLogger()
	if err != nil {
		fmt.Println("InitLogger err: ", err)
	}
	logger.Info("Info")
	logger.Debug("Debug")
	logger.Error("Error")
	logger.Fatal("Fatal")
	logger.Warning("Warning")
	logger.TRACE("trace")
}
