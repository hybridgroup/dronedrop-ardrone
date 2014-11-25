package main

import (
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/digispark"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()
	api.NewAPI(gbot).Start()

	digi := digispark.NewDigisparkAdaptor("digispark")
	servo := gpio.NewServoDriver(digi, "servo", "0")

	gbot.AddCommand("reset", func(params map[string]interface{}) interface{} {
		servo.Move(10)
		return true
	})

	gbot.AddCommand("drop", func(params map[string]interface{}) interface{} {
		servo.Move(150)
		return true
	})

	gbot.AddRobot(gobot.NewRobot("dropper", []gobot.Connection{digi}, []gobot.Device{servo}))

	gbot.Start()
}
