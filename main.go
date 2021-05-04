package main

import (
	_ "image/png"
	"weather-pi/cmd"
)

func main() {
	cmd.Execute()

	//e := epd.NewEpd2in13v3(sugaredLogger)
	//defer func(e *epd.Dev2in13v3) {
	//	if err := e.Close(); err != nil {
	//		sugaredLogger.With("err", err).Error("could not close device")
	//	}
	//}(e)
	//defer func(e *epd.Dev2in13v3) {
	//	if err := e.Clear(); err != nil {
	//		sugaredLogger.With("err", err).Error("could not clear the device")
	//	}
	//}(e)
	//err := e.Init()
	//if err != nil {
	//	sugaredLogger.With("err", err).Fatal("error while initializing device")
	//}
	//
	//err = e.Clear()
	//if err != nil {
	//	sugaredLogger.With("err", err).Fatal("error while clearing the device screen")
	//}

	//time.Sleep(5 * time.Second)
}
