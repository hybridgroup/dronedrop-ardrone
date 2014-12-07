package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/ardrone"
	"github.com/hybridgroup/gobot/platforms/digispark"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

const CONFIG_FILE = "/data/video/dronedrop.json"
const VERSION string = "0.3"

var GRAB uint8 = 40
var LOAD uint8 = 50
var DROP uint8 = 135

type config struct {
	Commander bool  `json:"commander"`
	Grab      uint8 `json:"grab"`
	Drop      uint8 `json:"drop"`
	Load      uint8 `json:"load"`
}

type pair struct {
	x float64
	y float64
}

func main() {
	c, err := readConfig()
	if err != nil {
		log.Println(err)
	}
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

	grab := func() {
		servo.Move(GRAB)
	}
	load := func() {
		servo.Move(LOAD)
	}
	drop := func() {
		servo.Move(DROP)
	}

	work := func() {
		if !c.Commander {
			for {
				if errs := ardroneAdaptor.Disconnect(); len(errs) > 0 {
					log.Println("disconnecting...")
					break
				}
			}
		}

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
			if !ardroneAdaptor.Connected() {
				if errs := ardroneAdaptor.Connect(); len(errs) > 0 {
					return errs
				}
			}

			if land {
				drone.Land()
				land = false
			} else {
				drone.TakeOff()
				land = true
			}
		}

		if name == "B" && action == "press" {
			grab()
		} else if name == "X" && action == "press" {
			load()
		} else if name == "Y" && action == "press" {
			drop()
		}

		return nil
	})

	gbot.AddCommand("grab", func(params map[string]interface{}) interface{} {
		grab()
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

	gbot.AddCommand("config", func(params map[string]interface{}) interface{} {
		if value, ok := params["grab"]; ok {
			GRAB = uint8(value.(float64))
			c.Grab = GRAB
		}
		if value, ok := params["load"]; ok {
			LOAD = uint8(value.(float64))
			c.Load = LOAD
		}
		if value, ok := params["drop"]; ok {
			DROP = uint8(value.(float64))
			c.Drop = DROP
		}
		writeConfig(c)
		return nil
	})

	gbot.AddCommand("commander", func(params map[string]interface{}) interface{} {
		if value, ok := params["enable"]; ok {
			if !value.(bool) {
				c.Commander = false
				writeConfig(c)
				for {
					if errs := ardroneAdaptor.Disconnect(); len(errs) > 0 {
						return nil
					}
				}
			}
		}

		c.Commander = true
		writeConfig(c)
		return nil
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

func readConfig() (c config, err error) {
	c = config{Commander: true, Grab: 40, Load: 50, Drop: 153}

	if _, err := os.Stat(CONFIG_FILE); err != nil {
		if err = writeConfig(c); err != nil {
			return c, err
		}
	} else if f, err := os.OpenFile(CONFIG_FILE, os.O_RDWR|os.O_CREATE, 0666); err != nil {
		return c, err
	} else {
		if err := json.NewDecoder(f).Decode(&c); err != nil {
			return c, err
		}
	}
	return
}

func writeConfig(c config) (err error) {
	if buf, err := json.Marshal(c); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(CONFIG_FILE, buf, 0666); err != nil {
			return err
		}
	}
	return
}
