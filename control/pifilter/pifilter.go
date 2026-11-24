/*********************************************************************************
Author: Nelson S Rosa
Description: This program implements a simple PID controller as defined
			in "Feedback Control for Computer Systems: Introducing Control Theory to
			Enterprise Programmers", Philipp K. Janert, 2014.
Date: 04/02/2023
*********************************************************************************/

package filterpi

import (
	"adaptivesystemscourse/internal/config"
	"math"
)

type FilterPI struct {
	Direction              float64
	Kp                     float64
	Ki                     float64
	Kd                     float64
	Max                    float64
	Min                    float64
	DeltaTime              int32
	Tau                    float64
	FilteredSetpoint       float64
	Integrator             float64
	PreviousError          float64
	PreviousPreviousError  float64
	SumPrevErrors          float64
	Out                    float64
	PreviousDifferentiator float64
	FiltePI                float64
}

func (c *FilterPI) Initialise(conf config.Config) error {

	c.Direction = conf.Experiment.Direction
	c.Min = conf.Experiment.Min
	c.Max = conf.Experiment.Max

	c.Kp = conf.Experiment.Kp
	c.Ki = conf.Experiment.Ki
	c.Kd = conf.Experiment.Kd
	c.DeltaTime = conf.Experiment.DeltaTime
	c.Tau = conf.Experiment.Tau

	c.Integrator = 0.0
	c.PreviousError = 0.0
	c.PreviousPreviousError = 0.0
	c.SumPrevErrors = 0.0
	c.Out = c.Min
	c.PreviousDifferentiator = 0.0
	c.FilteredSetpoint = 0.0

	return nil
}

func (c *FilterPI) Update(p ...float64) float64 {

	r := p[0] // goal
	y := p[1] // plant output

	// error
	//err = c.Direction * (r - y)
	a := math.Exp(-float64(c.DeltaTime) / c.Tau)
	c.FilteredSetpoint = a*c.FilteredSetpoint + (1-a)*r
	err := c.FilteredSetpoint - y

	//	fmt.Println("HERE FILTERPI", c.FilteredSetpoint, err)

	// Proportional
	proportional := c.Kp * err

	// Integrator (page 49)
	integrator := (c.SumPrevErrors + err) * c.Ki * float64(c.DeltaTime)

	// Differentiator (page 49)
	differentiator := c.Kd * (err - c.PreviousError) / float64(c.DeltaTime)

	// control law
	c.Out = proportional + integrator + differentiator

	if c.Out > c.Max {
		c.Out = c.Max
	} else if c.Out < c.Min {
		c.Out = c.Min
	}

	c.PreviousError = err
	c.SumPrevErrors += err

	//fmt.Println("HERE:: PID:: ", c.Info.Kp, r, y, err, c.Info.Out)
	return c.Out
}
