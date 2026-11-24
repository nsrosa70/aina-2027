/*********************************************************************************
Author: Nelson S Rosa
Description: This program implements a simple PID controller as defined
			in "Feedback Control for Computer Systems: Introducing Control Theory to
			Enterprise Programmers", Philipp K. Janert, 2014.
Date: 04/02/2023
*********************************************************************************/

package pi2

import (
	"adaptivesystemscourse/internal/config"
)

// See page 44: The factor DeltaTime is the length of time between
// successive  control actions, expressed in the units in which
// we measure time. (If we measure time in seconds and make 100
// control actions per second, then Delta time = 0.01; if we measure time
// in days and make one update per day, then Delta time = 1.)

type PI2Controller struct {
	Direction              float64
	Min                    float64
	Max                    float64
	Kp                     float64
	Ki                     float64
	Kd                     float64
	Beta                   float64
	DeltaTime              float64
	Integrator             float64
	PreviousError          float64
	PreviousPreviousError  float64
	SumPrevErrors          float64
	Out                    float64
	PreviousDifferentiator float64
}

func (c *PI2Controller) Initialise(conf config.Config) error {

	c.Direction = conf.Experiment.Direction
	c.Min = conf.Experiment.Min
	c.Max = conf.Experiment.Max

	c.Kp = conf.Experiment.Kp
	c.Ki = conf.Experiment.Ki
	c.Beta = conf.Experiment.Beta
	c.DeltaTime = float64(conf.Experiment.DeltaTime)

	c.Integrator = 0.0
	c.PreviousError = 0.0
	c.PreviousPreviousError = 0.0
	c.SumPrevErrors = 0.0
	c.Out = c.Min
	c.PreviousDifferentiator = 0.0

	return nil
}

func (c *PI2Controller) Update(p ...float64) float64 {

	r := p[0] // goal
	y := p[1] // plant output

	// errors
	err := c.Direction * (r - y)

	// Proportional
	proportional := c.Kp * (c.Beta*r - y)

	// Integrator (page 49)
	integrator := c.Ki * (c.SumPrevErrors + err) * c.DeltaTime

	// Differentiator (page 49)
	differentiator := c.Kd * (err - c.PreviousError) / c.DeltaTime

	// control law
	c.Out = proportional + integrator + differentiator

	if c.Out > c.Max {
		c.Out = c.Max
	} else if c.Out < c.Min {
		c.Out = c.Min
	}

	c.PreviousError = err
	c.SumPrevErrors += err

	//fmt.Println("HERE:: PID:: ", c.Kp, r, y, err, c.Out)
	return c.Out
}
