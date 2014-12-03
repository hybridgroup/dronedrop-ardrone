package main

import (
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/digispark"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

const CLOSE uint8 = 40
const LOAD uint8 = 50
const DROP uint8 = 135

func main() {
	gbot := gobot.NewGobot()
	api.NewAPI(gbot).Start()

	digi := digispark.NewDigisparkAdaptor("digispark")
	servo := gpio.NewServoDriver(digi, "servo", "0")

	gbot.AddCommand("close", func(params map[string]interface{}) interface{} {
		servo.Move(CLOSE)
		return true
	})

	gbot.AddCommand("load", func(params map[string]interface{}) interface{} {
		servo.Move(LOAD)
		return true
	})

	gbot.AddCommand("drop", func(params map[string]interface{}) interface{} {
		servo.Move(DROP)
		return true
	})

	gbot.AddRobot(gobot.NewRobot("dropper", []gobot.Connection{digi}, []gobot.Device{servo}))

	gbot.Start()
}
