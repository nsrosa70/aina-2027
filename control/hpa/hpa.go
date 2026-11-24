package hpa

import "C"
import (
	"adaptivesystemscourse/internal/config"
	"math"
)

type HPAController struct {
	Direction float64
	Max       float64
	Min       float64
	Out       float64
}

func (c *HPAController) Initialise(conf config.Config) error {

	c.Direction = conf.Experiment.Direction
	c.Min = conf.Experiment.Min
	c.Max = conf.Experiment.Max
	c.Out = 0.0

	return nil
}

func (c *HPAController) Update(p ...float64) float64 {
	r := p[0] // desired metric value (goal)
	y := p[1] // plant output

	// control law
	if c.Direction < 0 {
		c.Out = math.Round(c.Out * (y / r))
	} else {
		c.Out = math.Round(c.Out * (r / y))
	}

	// min-max checking
	if c.Out > c.Max {
		c.Out = c.Max
	} else if c.Out < c.Min {
		c.Out = c.Min
	}

	// fmt.Println("HERE: HPA: ", c.Direction, c.Min, c.Max, c.Out)
	return c.Out
}
