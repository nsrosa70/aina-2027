package ctype

import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/control/aimd"
	"adaptivesystemscourse/internal/control/def/info"
	"adaptivesystemscourse/internal/control/hpa"
	"adaptivesystemscourse/internal/control/lqr"
	"adaptivesystemscourse/internal/control/mpc"
	"adaptivesystemscourse/internal/control/native"
	"adaptivesystemscourse/internal/control/onoff/basic"
	"adaptivesystemscourse/internal/control/pi2"
	"adaptivesystemscourse/internal/control/pid"
	"adaptivesystemscourse/internal/control/pidhysteresis"
	filterpi "adaptivesystemscourse/internal/control/pifilter"
	"adaptivesystemscourse/internal/shared"
	"fmt"
)

func NewController(conf config.Config) (info.Controller, error) {
	switch conf.Experiment.ControllerType {
	case shared.PIDController:
		return &pid.PIDController{}, nil
	case shared.PI2Controller:
		return &pi2.PI2Controller{}, nil
	case shared.HysteresisPIDController:
		return &pidhysteresis.HysteresisPIDController{}, nil
	case shared.HPAController:
		return &hpa.HPAController{}, nil
	case shared.AIMDController:
		return &aimd.AIMDController{}, nil
	case shared.NativeController:
		return &native.Native{}, nil
	case shared.MPCController:
		return &mpc.MPCController{}, nil
	case shared.FilterPI:
		return &filterpi.FilterPI{}, nil
	case shared.LQRController:
		return &lqr.LQRController{}, nil
	case shared.BasicOnoff:
		return &basic.OnOffController{}, nil
	default:
		return nil, fmt.Errorf("unknown controller type: %d", conf.Experiment.ControllerType)
	}
}
