package main

import (
	"adaptivesystemscourse/internal/control/mpc"
	"fmt"
)

func main() {

	controller := mpc.MPC{
		A:         0.592,
		B:         234.97,
		Horizon:   5,
		GridSteps: 5,
		UMin:      0,
		UMax:      10,
		URef:      0,
		Lambda:    0.02,
	}

	x := 0.0
	setpoint := 8.0

	for k := 0; k < 20; k++ {
		u := controller.Update(x, setpoint)

		// Simulate one plant step. In a real system, read the next
		// measurement from the middleware instead.
		x = controller.A*x + controller.B*u

		fmt.Printf("step=%2d  control=%.2f  output=%.2f\n", k, u, x)
	}
}
