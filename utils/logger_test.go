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
	logger, err := InitLogger("DEBUG")
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

func TestNewLogDefault(t *testing.T) {
	InitSettings("../conf/config.yaml")
	logger, err := InitLoggerDefault()
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
