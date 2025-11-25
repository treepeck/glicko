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
//   - Mu: the player's rating converted to the Glicko-2 scale.
//   - Phi: the player's rating deviation converted to the Glicko-2 scale.
//   - Sigma: the player's rating volatility.
//   - Tau: the volatility change constraint.
//   - G: a weighting function that reduces the influence of enemies with high
//     rating deviations.
//   - E: the expected score against an enemy with the specified rating and
//     rating deviation.
//   - V: the player's performance rating based only on game outcomes.
//   - Delta: the estimated value of the new player's rating.
//
// See https://www.glicko.net/glicko/glicko2.pdf for more details.
package glicko

import "math"

// Evaluation represents a player's strength estimate.
type Evaluation struct {
	Rating     float64
	Deviation  float64
	Volatility float64
}

// Outcome represents a game information used to calculate the new player's
// rating.  It stores values needed only for internal calculations, hence all
// fields are unexported.
type Outcome struct {
	mu  float64
	phi float64
	g   float64
	e   float64
	// 0 - loss; 0.5 - draw; 1 - win.
	score float64
}

// NewOutcome converts the provided values to the Glicko-2 scale and returns
// an [Outcome].
func NewOutcome(rating, enemyRating, enemyDeviation, score float64) Outcome {
	mu := calcMu(rating)
	eMu := calcMu(enemyRating)
	phi := calcPhi(enemyDeviation)
	g := calcG(phi)

	return Outcome{
		mu:    eMu,
		phi:   phi,
		g:     g,
		e:     calcE(mu, eMu, g),
		score: score,
	}
}

const (
	tau     = 0.5
	scale   = 173.7178
	epsilon = 0.000001
)

// EvaluatePlayer returns the updated player rating, deviation, and
// volatility based on the rating period outcomes.
func EvaluatePlayer(curr Evaluation, outcomes []Outcome) Evaluation {
	// Step 1.
	mu := calcMu(curr.Rating)
	phi := calcPhi(curr.Deviation)
	sigma := curr.Volatility

	// Step 2.
	v := calcV(outcomes)

	// Step 3.
	delta := calcDelta(v, outcomes)

	// Step 4.
	curr.Volatility = sigmaPrime(sigma, delta, phi, v)

	// Step 5.
	phiStar := math.Sqrt(pow2(phi) + pow2(curr.Volatility))

	// Step 6.
	curr.Deviation = 1 / math.Sqrt((1/pow2(phiStar))+(1/v))

	sum := 0.0
	for i := range outcomes {
		sum += outcomes[i].g * (outcomes[i].score - outcomes[i].e)
	}
	curr.Rating = mu + pow2(curr.Deviation)*sum

	// Step 7.
	return Evaluation{
		Rating:     scale*curr.Rating + 1500,
		Deviation:  curr.Deviation * scale,
		Volatility: curr.Volatility,
	}
}

func calcPhi(deviation float64) float64 { return deviation / scale }

func calcMu(rating float64) float64 { return (rating - 1500) / scale }

func calcG(phi float64) float64 {
	return 1 / math.Sqrt(1+(3*pow2(phi)/pow2(math.Pi)))
}

func calcE(mu, eMu, g float64) float64 {
	return 1 / (1 + math.Exp(-g*(mu-eMu)))
}

func calcV(outcomes []Outcome) float64 {
	v := 0.0
	for i := range outcomes {
		v += pow2(outcomes[i].g) * outcomes[i].e * (1 - outcomes[i].e)
	}
	return 1 / v
}

func calcDelta(v float64, outcomes []Outcome) float64 {
	delta := 0.0
	for i := range outcomes {
		delta += outcomes[i].g * (outcomes[i].score - outcomes[i].e)
	}
	return v * delta
}

// f is a helper function used in to compute the sigma prime.
func f(delta, phi, v, a, x float64) float64 {
	exp := math.Exp(x)
	left := (exp * (pow2(delta) - pow2(phi) - v - exp) / (2 * pow2(pow2(phi)+v+exp)))
	right := ((x - a) / pow2(tau))
	return left - right
}

// sigmaPrime computes the new player's rating volatility.
func sigmaPrime(sigma, delta, phi, v float64) float64 {
	a := math.Log(pow2(sigma))

	B := 0.0
	if pow2(delta) > pow2(phi)+v {
		B = math.Log(pow2(delta) - pow2(phi) - v)
	} else {
		for k := 1.0; ; k++ {
			B = a - k*tau

			if f(delta, phi, v, a, a-k*tau) > 0 {
				break
			}
		}
	}

	A := a
	fA := f(delta, phi, v, a, A)
	fB := f(delta, phi, v, a, B)

	for math.Abs(B-A) > epsilon {
		C := A + (A-B)*fA/(fB-fA)
		fC := f(delta, phi, v, a, C)

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

func pow2(val float64) float64 { return val * val }
