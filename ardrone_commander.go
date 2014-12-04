package main

import (
	"math"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/ardrone"
	"github.com/hybridgroup/gobot/platforms/digispark"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

type pair struct {
	x float64
	y float64
}

const CLOSE uint8 = 40
const LOAD uint8 = 50
const DROP uint8 = 135
const VERSION string = "0.1"

func main() {
	gbot := gobot.NewGobot()

	server := api.NewAPI(gbot)
	server.Port = "8080"
	server.Start()

	ardroneAdaptor := ardrone.NewArdroneAdaptor("Drone", "127.0.0.1")
	drone := ardrone.NewArdroneDriver(ardroneAdaptor, "drone")

	digi := digispark.NewDigisparkAdaptor("digispark")
	servo := gpio.NewServoDriver(digi, "servo", "0")

	rightStick := pair{x: 0, y: 0}
	leftStick := pair{x: 0, y: 0}
	land := false

	close := func() {
		servo.Move(CLOSE)
	}
	load := func() {
		servo.Move(LOAD)
	}
	drop := func() {
		servo.Move(DROP)
	}

	work := func() {
		go func() {
			for {
				pair := leftStick
				if pair.y > 0.1 {
					drone.Forward(validatePitch(pair.y))
				} else if pair.y < -0.1 {
					drone.Backward(validatePitch(pair.y))
				} else {
					drone.Forward(0)
				}

				if pair.x > 0.1 {
					drone.Right(validatePitch(pair.x))
				} else if pair.x < -0.1 {
					drone.Left(validatePitch(pair.x))
				} else {
					drone.Right(0)
				}
				<-time.After(10 * time.Millisecond)
			}
		}()

		go func() {
			for {
				pair := rightStick
				if pair.y > 0.1 {
					drone.Up(validatePitch(pair.y))
				} else if pair.y < -0.1 {
					drone.Down(validatePitch(pair.y))
				} else {
					drone.Up(0)
				}

				if pair.x > 0.3 {
					drone.Clockwise(validatePitch(pair.x))
				} else if pair.x < -0.3 {
					drone.CounterClockwise(validatePitch(pair.x))
				} else {
					drone.Clockwise(0)
				}
				<-time.After(10 * time.Millisecond)
			}
		}()
	}

	robot := gobot.NewRobot("joystick",
		[]gobot.Connection{ardroneAdaptor, digi},
		[]gobot.Device{drone, servo},
		work,
	)

	robot.AddCommand("joystick_event", func(params map[string]interface{}) interface{} {
		name := params["name"].(string)
		if name == "left" {
			leftStick.y = params["position"].(map[string]interface{})["y"].(float64)
			leftStick.x = params["position"].(map[string]interface{})["x"].(float64)
		} else if name == "right" {
			rightStick.y = params["position"].(map[string]interface{})["y"].(float64)
			rightStick.x = params["position"].(map[string]interface{})["x"].(float64)
		}

		return nil
	})

	robot.AddCommand("button_event", func(params map[string]interface{}) interface{} {
		name := params["name"].(string)
		action := params["action"].(string)

		if name == "A" && action == "press" {
			if land {
				drone.Land()
				land = false
			} else {
				drone.TakeOff()
				land = true
			}
		}

		if name == "B" && action == "press" {
			close()
		} else if name == "X" && action == "press" {
			load()
		} else if name == "Y" && action == "press" {
			drop()
		}

		return nil
	})

	gbot.AddCommand("close", func(params map[string]interface{}) interface{} {
		close()
		return true
	})

	gbot.AddCommand("load", func(params map[string]interface{}) interface{} {
		load()
		return true
	})

	gbot.AddCommand("drop", func(params map[string]interface{}) interface{} {
		drop()
		return true
	})

	gbot.AddCommand("version", func(params map[string]interface{}) interface{} {
		return VERSION
	})

	gbot.AddRobot(robot)
	gbot.Start()
}

func validatePitch(value float64) float64 {
	value = math.Abs(value)
	if value >= 0.1 {
		if value <= 1.0 {
			return float64(int(value*100)) / 100
		}
		return 1.0
	}
	return 0.0
}
