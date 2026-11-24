package lqr

import (
	"adaptivesystemscourse/internal/config"
	"fmt"
	"math"
	"os"
)

type LQRController struct {
	A, B, Bias float64 // y[k+1] = A*y[k] + B*u[k] + Bias
	Q, R       float64
	K          float64
	UMin, UMax float64
	UStep      float64 // 0 disables quantization; use 1 for integer link credit
}

// NewLQRController initializes the controller and calculates K from the
// discrete-time algebraic Riccati equation.
func (c *LQRController) Initialise(conf config.Config) error {
	values := []float64{conf.Experiment.A, conf.Experiment.B, conf.Experiment.Bias, conf.Experiment.Q, conf.Experiment.R, conf.Experiment.UMin, conf.Experiment.UMax, conf.Experiment.UStep}
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("LQR parameters must be finite")
		}
	}

	if conf.Experiment.B == 0 {
		return fmt.Errorf("B must be nonzero")
	}
	if conf.Experiment.Q <= 0 || conf.Experiment.R <= 0 {
		return fmt.Errorf("Q and R must be positive")
	}
	if conf.Experiment.UMin > conf.Experiment.UMax {
		return fmt.Errorf("UMin must not exceed UMax")
	}
	if conf.Experiment.UStep < 0 {
		return fmt.Errorf("UStep must be zero or positive")
	}

	k, err := discreteLQRGain(conf.Experiment.A, conf.Experiment.B, conf.Experiment.Q, conf.Experiment.R)
	if err != nil {
		return err
	}

	c.A = conf.Experiment.A
	c.B = conf.Experiment.B
	c.Bias = conf.Experiment.Bias
	c.Q = conf.Experiment.Q
	c.R = conf.Experiment.R
	c.K = conf.Experiment.K
	c.UMin = conf.Experiment.UMin
	c.UMax = conf.Experiment.UMax
	c.UStep = conf.Experiment.UStep
	c.K = k

	//fmt.Printf("%v %v %v %v %v %v %v %v %v %v\n", c.A, c.B, c.Bias, c.Q, c.R, c.K, c.UMin, c.UMax, c.UStep, c.K)
	return nil
}

// Update computes the actuator command for the current reference and output.
func (c *LQRController) Update(p ...float64) float64 {
	r := p[0] // goal
	y := p[1] // plant output

	if math.IsNaN(r) || math.IsInf(r, 0) ||
		math.IsNaN(y) || math.IsInf(y, 0) {
		fmt.Printf("LQR:: reference and measurement must be finite")
		os.Exit(1)
	}

	// Steady-state input for the current reference:
	// reference = A*reference + B*uEquilibrium + Bias.
	uEquilibrium := ((1-c.A)*r - c.Bias) / c.B

	// LQR feedback around the current reference and its equilibrium input.
	u := uEquilibrium - c.K*(y-r)
	if math.IsNaN(u) || math.IsInf(u, 0) {
		fmt.Printf("LQR:: computed actuator command is not finite")
		os.Exit(1)
	}

	// Apply actuator limits.
	u = clamp(u, c.UMin, c.UMax)

	//fmt.Println("HERE: LQR: ", u, c.UMax, c.UMin)
	// Quantize, if requested. The grid is anchored at UMin.
	if c.UStep > 0 {
		u = c.UMin + math.Round((u-c.UMin)/c.UStep)*c.UStep
		u = clamp(u, c.UMin, c.UMax)
	}
	return u
}

func discreteLQRGain(a, b, q, r float64) (float64, error) {
	// Scalar discrete-time Riccati iteration.
	p := q
	converged := false

	for i := 0; i < 10000; i++ {
		denominator := r + b*b*p
		if denominator <= 0 || math.IsNaN(denominator) || math.IsInf(denominator, 0) {
			return 0, fmt.Errorf("invalid Riccati denominator")
		}

		nextP := q + a*a*p - (a*b*p)*(a*b*p)/denominator
		if math.IsNaN(nextP) || math.IsInf(nextP, 0) {
			return 0, fmt.Errorf("Riccati iteration diverged")
		}

		if math.Abs(nextP-p) <= 1e-12*(1+math.Abs(nextP)) {
			p = nextP
			converged = true
			break
		}
		p = nextP
	}

	if !converged {
		return 0, fmt.Errorf("Riccati iteration did not converge")
	}

	k := (a * b * p) / (r + b*b*p)
	if math.Abs(a-b*k) >= 1 {
		return 0, fmt.Errorf("computed LQR gain does not stabilize the model")
	}
	return k, nil
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
