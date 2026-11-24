package old_version

import (
	"adaptivesystemscourse/internal/config"
	"fmt"
	"math"
	"os"
)

type LQRController struct {
	A    float64
	B    float64
	Bias float64
	Q    float64
	R    float64
	K    float64
	UMin float64
	UMax float64
	Out  float64
}

func (c *LQRController) Initialise(conf config.Config) error {
	if conf.Experiment.B == 0 {
		return fmt.Errorf("B must be nonzero")
	}
	if conf.Experiment.Q <= 0 || conf.Experiment.R <= 0 {
		return fmt.Errorf("Q and R must be positive")
	}
	if conf.Experiment.UMin > conf.Experiment.UMax {
		return fmt.Errorf("UMin must be less than or equal to UMax")
	}

	p := conf.Experiment.Q
	const tolerance = 1e-12
	const maxIterations = 10000

	for i := 0; i < maxIterations; i++ {
		denominator := conf.Experiment.R + conf.Experiment.B*conf.Experiment.B*p
		pNext := conf.Experiment.Q + conf.Experiment.A*conf.Experiment.A*p - (conf.Experiment.A*conf.Experiment.B*p)*(conf.Experiment.A*conf.Experiment.B*p)/denominator

		if math.Abs(pNext-p) < tolerance {
			p = pNext
			break
		}
		p = pNext
	}

	k := (conf.Experiment.B * p * conf.Experiment.A) / (conf.Experiment.R + conf.Experiment.B*conf.Experiment.B*p)

	c.A = conf.Experiment.A
	c.B = conf.Experiment.B
	c.Bias = conf.Experiment.Bias
	c.Q = conf.Experiment.Q
	c.R = conf.Experiment.R
	c.UMax = conf.Experiment.Max
	c.UMin = conf.Experiment.Min
	c.K = k

	return nil
}

func (c *LQRController) Update(p ...float64) float64 {

	r := p[0] // goal
	y := p[1] // plant output

	if math.IsNaN(y) || math.IsInf(y, 0) ||
		math.IsNaN(r) || math.IsInf(r, 0) {
		fmt.Errorf("measurement and reference must be finite")
		os.Exit(1)
	}

	// At steady state, reference = A*reference + B*uEquilibrium.
	//uEquilibrium := ((1 - c.A) / c.B) * r
	uEquilibrium := ((1-c.A)*r - c.Bias) / c.B

	// Feedback around the reference and its equilibrium actuator value.
	u := uEquilibrium - c.K*(y-r)

	c.Out = u

	// Enforce actuator limits.
	if c.Out < c.UMin {
		c.Out = c.UMin
	}
	if c.Out > c.UMax {
		c.Out = c.UMax
	}

	return c.Out
}
