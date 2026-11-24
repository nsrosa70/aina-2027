package native

import "C"
import (
	"adaptivesystemscourse/internal/config"
	"adaptivesystemscourse/internal/shared"
)

type Native struct {
	Min float64
	Max float64
	Out float64
}

func (c *Native) Initialise(conf config.Config) error {

	c.Min = conf.Experiment.Min
	c.Max = conf.Experiment.Max
	if conf.Experiment.MOMType == shared.Pulsar {
		c.Out = float64(conf.Experiment.InitialDispatchRate)
	} else {
		c.Out = float64(conf.Experiment.InitialLinkedCredit)
	}

	return nil
}

func (c *Native) Update(p ...float64) float64 {
	//r := p[0] // reference
	//y := p[1] //plant output not used

	if c.Out < c.Min {
		c.Out = c.Min
	} else if c.Out > c.Max {
		c.Out = c.Max
	}
	return c.Out
}
