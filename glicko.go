// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Author: Artem Bielikov artem.bielikov@treepeck.com

// Package glicko implements the player-strength evaluation based on the
// Glicko-2 rating system.
//
// This code names variables and functions according to the conventions used in
// Professor Mark E. Glickman's paper:
//   - **mu**: the player's strength estimate (rating converted to the Glicko-2
//     scale).
//   - **phi**: the rating deviation converted to the Glicko-2 scale.  Phi
//     defines the bounds of the 95% confidence interval, where the lower bound
//     is mu-2*phi, and the upper bound is mu+2*phi.
//   - **sigma**: the degree of expected fluctuation in a player's rating.  This
//     value is high when a player has erratic performances, and low when the
//     player performs at a consistent level.
//   - **tau**: the volatility change constraint.
//   - **g**: a weighting function that reduces the influence of enemies with
//     high phi values.
//   - **e**: the expected score against an enemy with the specified mu and phi.
//   - **v**: the updated player's mu based only on expected game outcomes.
//   - **delta**: the estimated value of the updated player's mu based on actual
//     game outcomes.
//
// See https://www.glicko.net/glicko/glicko2.pdf for more details.
package glicko

import "math"

// Recommended values based on the original Glicko-2 paper.
const (
	DefaultRating = 1500
	// This value is also an upper bound, since the system cannot be less
	// certain about a player's rating than it is for an unrated player.
	DefaultDeviation  = 350
	DefaultVolatility = 0.06
	DefaultTau        = 0.5
	DefaultFactor     = 173.7178
	DefaultEpsilon    = 0.000001
)

// Converter performs conversions between the Glicko-2 and traditional
// "Elo-style" rating scales.  Internally all calculations to estimate the
// player's [Strength] are performed using the Glicko-2 scaled values.
type Converter struct {
	// Default rating of the unrated player.  Use [DefaultRating] constant for
	// the recommended value.
	Rating float64
	// Default deviation of the unrated player.  Use [DefaultDeviation] constant
	// for the recommended value.
	Deviation float64
	// Scaling factor.  Use [DefaultFactor] constant for the recommended value.
	Factor float64
}

// Rating2Mu converts rating to the Glicko-2 scale.
func (c Converter) Rating2Mu(rating float64) float64 {
	return (rating - c.Rating) / c.Factor
}

// Deviation2Phi converts deviation to the Glicko-2 scale.
func (c Converter) Deviation2Phi(deviation float64) float64 {
	return deviation / c.Factor
}

// Mu2Rating converts mu to the "Elo-style" rating scale.
func (c Converter) Mu2Rating(mu float64) float64 {
	return mu*c.Factor + c.Rating
}

// Phi2Deviation converts phi to the "Elo-style" rating deviation scale.
func (c Converter) Phi2Deviation(phi float64) float64 {
	return phi * c.Factor
}

// Strength represents a player's strength estimate.
type Strength struct {
	Mu    float64
	Phi   float64
	Sigma float64
}

// Outcome represents a game information used to calculate the new player's
// rating.  It stores only Glicko-2 scaled values which are used for internal
// calculations, hence all fields are unexported.
type Outcome struct {
	mu    float64
	phi   float64
	g     float64
	e     float64
	score float64
}

// NewOutcome returns an [Outcome] with the specified parameters.  The score
// parameter has to be 0 if the player lost, 0.5 for a draw, and 1 if the
// player won.
func NewOutcome(mu, enemyMu, enemyPhi, score float64) Outcome {
	g := 1 / math.Sqrt(1+(3*pow2(enemyPhi)/pow2(math.Pi)))
	e := 1 / (1 + math.Exp(-g*(mu-enemyMu)))

	return Outcome{
		mu:    enemyMu,
		phi:   enemyPhi,
		g:     g,
		e:     e,
		score: score,
	}
}

// Estimator performs calculations of the player's strength based on the
// provided initial [Strength] and match [Outcome].
type Estimator struct {
	// Lower bound of the possible mu value.
	MinPhi float64
	// Upper bound of the possible mu value.
	MaxPhi float64
	// Lower bound of the possible phi value.
	MinMu float64
	// Upper bound of the possible phi value.
	MaxMu float64
	// Lower bound of the possible sigma value.
	MinSigma float64
	// Upper bound of the possible sigma value.
	MaxSigma float64
	// System variable.  Use [DefaultTau] constant for the recommended value.
	Tau float64
	// System variable.  Use [DefaultEpsilon] constant for the recommended value.
	Epsilon float64
}

// Estimate updates the player's [Strength] by analyzing:
//   - s: player's [Strength] at the onset of the rating period.
//   - outcomes: match [Outcome] withing a single rating period.
//
// It's a caller responsilibty to call [Validate] to validate the results.
func (e Estimator) Estimate(s *Strength, outcomes []Outcome) {
	// If a player doesn't compete during the rating period, the mu and sigma
	// remains the same, but the phi increases.
	if len(outcomes) == 0 {
		s.Phi = math.Sqrt(pow2(s.Phi) + pow2(s.Sigma))
		return
	}

	// Calculate v.
	v := 0.0
	for i := range outcomes {
		v += pow2(outcomes[i].g) * outcomes[i].e * (1 - outcomes[i].e)
	}
	v = 1 / v

	// Calculate delta.
	tmp := 0.0
	for i := range outcomes {
		tmp += outcomes[i].g * (outcomes[i].score - outcomes[i].e)
	}
	delta := tmp * v

	// Calculate new sigma.
	s.Sigma = e.sigmaPrime(*s, delta, v)

	// Calculate new phi.
	phiStar := math.Sqrt(pow2(s.Phi) + pow2(s.Sigma))
	s.Phi = 1 / math.Sqrt(1/pow2(phiStar)+1/v)

	// Calculate new mu.
	s.Mu = s.Mu + pow2(s.Phi)*tmp
}

// Validate validates the [Srength] by checking if the values satisfy the
// established limits.
func (e Estimator) Validate(s *Strength) {
	if s.Mu < e.MinMu {
		s.Mu = e.MinMu
	} else if s.Mu > e.MaxMu {
		s.Mu = e.MaxMu
	}

	if s.Phi < e.MinPhi {
		s.Phi = e.MinPhi
	} else if s.Phi > e.MaxPhi {
		s.Phi = e.MaxPhi
	}

	if s.Sigma < e.MinSigma {
		s.Sigma = e.MinSigma
	} else if s.Sigma > e.MaxSigma {
		s.Sigma = e.MaxSigma
	}
}

// Internal helper function.
func (e Estimator) sigmaPrime(s Strength, delta, v float64) float64 {
	a := math.Log(pow2(s.Sigma))

	B := 0.0
	if pow2(delta) > pow2(s.Phi)+v {
		B = math.Log(pow2(delta) - pow2(s.Phi) - v)
	} else {
		for k := 1.0; ; k++ {
			B = a - k*e.Tau

			if e.f(delta, s.Phi, v, a, a-k*e.Tau) > 0 {
				break
			}
		}
	}

	A := a
	fA := e.f(delta, s.Phi, v, a, A)
	fB := e.f(delta, s.Phi, v, a, B)

	for math.Abs(B-A) > e.Epsilon {
		C := A + (A-B)*fA/(fB-fA)
		fC := e.f(delta, s.Phi, v, a, C)

		if fC*fB <= 0 {
			A = B
			fA = fB
		} else {
			fA /= 2
		}

		B = C
		fB = fC
	}
	return math.Exp(a / 2)
}

// Internal helper function.
func (e Estimator) f(phi, v, delta, a, x float64) float64 {
	exp := math.Exp(x)
	tmp := pow2(phi) - v - exp
	return exp*(pow2(delta)-tmp)/(2*pow2(tmp)) - (x-a)/pow2(e.Tau)
}

func pow2(val float64) float64 { return val * val }
