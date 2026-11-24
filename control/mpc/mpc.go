package mpc

import (
	"adaptivesystemscourse/internal/config"
	"math"
)

type MPCController struct {
	A                float64
	B                float64
	Bias             float64
	Horizon          float64
	GridSteps        float64
	UMin             float64
	UMax             float64
	URef             float64 // initial/legacy actuator value
	Uscale           float64
	NormalizedLambda float64
	Lambda           float64
}

func (c *MPCController) Initialise(conf config.Config) error {
	c.A = conf.Experiment.A
	c.B = conf.Experiment.B
	c.Bias = conf.Experiment.Bias
	c.Horizon = conf.Experiment.Horizon
	c.GridSteps = conf.Experiment.GridSteps
	c.UMin = conf.Experiment.UMin
	c.UMax = conf.Experiment.UMax
	c.URef = conf.Experiment.URef // initial/legacy actuator value
	c.Uscale = conf.Experiment.UScale
	c.NormalizedLambda = conf.Experiment.NormalizedLambda
	c.Lambda = conf.Experiment.Lambda

	return nil
}

// Update returns the first control action from the lowest-cost actuator
// sequence. Arguments are target and measured output, followed optionally by
// the actuator value actually applied during the previous control interval.
//
// Prefer passing all three values:
//
//	u := controller.Update(target, measuredOutput, lastAppliedU)
//
// The two-argument form is retained for compatibility and uses configured
// URef as the previous actuator value.

func (c *MPCController) Update(p ...float64) float64 {
	target, y := p[0], p[1]
	uPrev := c.URef

	c.URef = c.updateForReferences([]float64{target}, y, uPrev)
	return c.URef
}

// UpdateForReferences plans against a reference trajectory. References are
// consumed one per prediction step. If fewer than Horizon are provided, the
// final reference is held; if more are provided, only the first Horizon are
// used. uPrev must be the actuator value applied during the previous interval.
func (c *MPCController) UpdateForReferences(
	refs []float64,
	y, uPrev float64,
) float64 {
	if len(refs) == 0 {
		return uPrev
	}
	return c.updateForReferences(refs, y, uPrev)
}

func (c *MPCController) updateForReferences(
	refs []float64,
	y, uPrev float64,
) float64 {
	if c.Horizon <= 0 || c.GridSteps <= 0 {
		return uPrev
	}

	// Scale the normalized move penalty using the first reference.
	// The penalty is on actuator changes, not deviation from URef.
	c.Lambda = c.NormalizedLambda
	if c.Uscale != 0 {
		c.Lambda *= math.Pow(refs[0]/c.Uscale, 2)
	}

	bestCost := math.Inf(1)
	bestFirstU := uPrev

	var search func(
		depth int,
		state, previousU, cost, firstU float64,
	)

	search = func(
		depth int,
		state, previousU, cost, firstU float64,
	) {
		if depth == int(c.Horizon) {
			if cost < bestCost {
				bestCost = cost
				bestFirstU = firstU
			}
			return
		}

		ref := refs[len(refs)-1]
		if depth < len(refs) {
			ref = refs[depth]
		}

		for i := 0; i < int(c.GridSteps); i++ {
			u := c.candidate(i)

			// Affine plant model:
			// y[k+1] = A*y[k] + B*u[k] + Bias
			nextState := c.A*state +
				c.B*u +
				c.Bias

			trackingError := ref - nextState
			move := u - previousU

			nextCost := cost +
				trackingError*trackingError +
				c.Lambda*move*move

			nextFirstU := firstU
			if depth == 0 {
				nextFirstU = u
			}

			search(
				depth+1,
				nextState,
				u,
				nextCost,
				nextFirstU,
			)
		}
	}

	search(0, y, uPrev, 0, uPrev)
	return bestFirstU
}

func (c *MPCController) candidate(index int) float64 {
	if c.GridSteps == 1 {
		return c.UMin +
			(c.UMax-c.UMin)/2
	}

	fraction := float64(index) /
		float64(int(c.GridSteps)-1)

	return c.UMin +
		fraction*(c.UMax-c.UMin)
}
