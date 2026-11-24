package mico

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/control/def/ctype"
	"adaptivesystemscourse/internal/control/def/info"
	"adaptivesystemscourse/internal/control/mico/actuators"
	"fmt"
)

type MICO struct {
	MOM        int // MOM
	Controller info.Controller
	Actuator   actuators.Actuator
}

func (m *MICO) Initialise(conf config.Config) error {

	// create controller
	controller, err := ctype.NewController(conf)
	if err != nil {
		return fmt.Errorf("Error in creating controller\n")
	}

	// initialise controller
	err = controller.Initialise(conf)
	if err != nil {
		return fmt.Errorf("Error in initialising controller\n")
	}

	m.Controller = controller

	return nil
}

func (m *MICO) Update(p ...float64) float64 {
	return m.Controller.Update(p...)
}
