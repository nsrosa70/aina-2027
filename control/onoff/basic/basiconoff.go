package basic

import (
	"adaptivesystemscourse/internal/config"
)

type OnOffController struct {
	Max       float64
	Min       float64
	Direction float64
}

func (c *OnOffController) Initialise(conf config.Config) error {

	c.Direction = conf.Experiment.Direction
	c.Min = conf.Experiment.Min
	c.Max = conf.Experiment.Max

	return nil
}

func (c *OnOffController) Update(p ...float64) float64 {
	u := 0.0

	s := p[0] // goal
	y := p[1] // plant output

	// error
	//err := c.cfg.Direction * (s - y)
	err := c.Direction * (s - y)

	// control law
	if err > 0 { // lower than the goal
		//u = c.cfg.Max
		u = c.Max
	} else { // higher than the goal
		//u = c.cfg.Min
		u = c.Min
	}
	return u
}
