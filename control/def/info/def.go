package info

import "adaptivesystemscourse/internal/config"

type FuzzySet struct {
	Center float64
	Width  float64
}

type Controller interface {
	Initialise(conf config.Config) error
	Update(p ...float64) float64
}

type Config struct {
	TypeName               string // Controller ctype name
	Direction              float64
	Kp                     float64  // kp constant used by PID controllers
	Ki                     float64  // ki constant used by PID controllers
	Kd                     float64  // kd constant used by PID controllers
	Min                    float64  // Minimum value of the controller output
	Max                    float64  // Maximum value of the controller output
	Integrator             float64  // Integrator component
	SumPrevErrors          float64  // Sum of previous errors -- used by some controllers
	PreviousOut            float64  // Previous output -- used by some controllers
	PreviousError          float64  // Last error -- used by some controllers
	PreviousPreviousError  float64  // Penultimate error -- used by some controllers
	PreviousDifferentiator float64  // Ante penultimate error -- used by some controllers
	DeadZone               float64  // Dead zone band used by some controllers
	HysteresisBand         float64  // Hysteresis band used by some controllers
	Out                    float64  // Controller output
	GainTrigger            float64  // Gain trigger based on two level gains
	PC                     float64  // for HPA only
	PreviousRate           float64  // for AsTAR
	Alfa                   float64  // For controllers Setpoint weighting
	Beta                   float64  // for controller with two degrees of freedom/Setpointweighting
	InputSet               FuzzySet // for Fuzzy controllers
	OutputSet              FuzzySet // for Fuzzy controllers
	DeltaTime              float64  // for PID controllers
	A                      float64  // plant model (for MPC/LQR)
	B                      float64  // plant model (for MPC/LQR)
	Bias                   float64  // plant model (for MPC)
	Horizon                float64  // prediction steps (for MPC)
	GridSteps              float64  // candidate control values (for MPC)
	URef                   float64  // preferred control value (for MPC)
	UPrevious              float64  // previous URefe (for MPC)
	Uscale                 float64  // u scale (for MPC)
	Lambda                 float64  // penalty on control effort (for MPC)
	NormalizedLambda       float64  // for MPC
	UMin                   float64
	UMax                   float64
	FilteredSetpoint       float64 // for FilterPI
	Tau                    float64 // for FilterPI
	Q                      float64 // for LQR
	R                      float64 // for LQR
	K                      float64 // for LQR
}
