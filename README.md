[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/treepeck/glicko.svg)](https://pkg.go.dev/github.com/treepeck/glicko)

Package glicko implements the instant Glicko-2 rating system.

The stock Glicko-2 system expects a collection of games within a `rating period`<br/>
to apply the rating algorithm.

This implementation allows the rating to be estimated sequentially based on<br/>
individual match outcomes rather than in batches. All system parameters are<br/>
bounded and user-defined, making the system both flexible and safe.

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
