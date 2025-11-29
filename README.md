[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/treepeck/chego.svg)](https://pkg.go.dev/github.com/treepeck/chego)

Package glicko implements the player's strength estimate based on the Glicko-2<br/>
rating system.

Originally designed as an improvement over existing rating systems, Glicko-2 has<br/>
become a popular choice among game developers.

However, one potential caveat when using the stock Glicko-2 system is that it<br/>
operates on a collection of games within a `rating period` before applying the<br/>
rating algorithm.  This is not suitable for most online games, since players<br/>
often expect their ratings to update immediately after each match.

This implementation aims to address that limitation by allowing the player's<br/>
rating to be estimated sequentially based on individual match outcomes rather<br/>
than in batches.  To achieve this, the periodFraction parameter was introduced<br/>
into the Estimate function.  As a result, rating, deviation, and volatility<br/>
evolve smoothly when matches occur at arbitrary moments instead of being<br/>
grouped into fixed rating periods.

Another notable modification is the introduction of bounds on the player’s rating,<br/>deviation, and volatility.  All system parameters are user-defined, making the<br/>
system both flexible and safe.

## Usage

To install `glicko`, run `go get`:

```
go get github.com/treepeck/glicko
```

Here is a simple example (based on the original Glicko-2 paper):

```go
package main

import (
	"fmt"

	"github.com/treepeck/glicko"
)

func main() {
	c := glicko.Converter{
		Rating:    glicko.DefaultRating,
		Deviation: glicko.DefaultDeviation,
		Factor:    glicko.DefaultFactor,
	}

	// Initial player's strength.
	s := glicko.Strength{
		Mu:    c.Rating2Mu(glicko.DefaultRating),
		Phi:   c.Deviation2Phi(200),
		Sigma: glicko.DefaultVolatility,
	}

	// Match outcomes.
	outcomes := []glicko.Outcome{
		{Mu: c.Rating2Mu(1400), Phi: c.Deviation2Phi(30), Score: 1},
		{Mu: c.Rating2Mu(1550), Phi: c.Deviation2Phi(100), Score: 0},
		{Mu: c.Rating2Mu(1700), Phi: c.Deviation2Phi(300), Score: 0},
	}

	// Initialize the strength estimator.
	e := glicko.Estimator{
		MinMu: c.Rating2Mu(10), MaxMu: c.Rating2Mu(5000),
		MinPhi:   c.Deviation2Phi(30),
		MaxPhi:   c.Deviation2Phi(glicko.DefaultDeviation),
		MinSigma: 0.04, MaxSigma: 0.08,
		Tau: glicko.DefaultTau, Epsilon: glicko.DefaultEpsilon,
	}

	// Estimate player's rating based on game outcomes sequantially.
	// New rating can be calculated immediately after each match.
	for i := range outcomes {
		e.Estimate(&s, outcomes[i], 1)
	}

	// Those values slighly (+- 1 rating point) differ from the original
	// Glicko-2 example.  This is because the implementation processes each
	// game sequentially and the original algorithm is not linear.  The
	// difference is negligable is most cases.

	// Prints "1463.7884049111285"
	fmt.Println(int(c.Mu2Rating(s.Mu)))
	// Prints "151.87316169372073"
	fmt.Println(c.Phi2Deviation(s.Phi))
	// Prints "0.05999440707992557"
	fmt.Println(s.Sigma)
}
```

## Acknowledgments

https://www.glicko.net/glicko/glicko2.pdf

https://blog.hypersect.com/the-online-skill-ranking-of-inversus-deluxe/