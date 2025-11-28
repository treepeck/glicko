package glicko

import "testing"

func TestEstimate(t *testing.T) {
	c := Converter{
		Rating:    DefaultRating,
		Deviation: DefaultDeviation,
		Factor:    DefaultFactor,
	}

	testcases := []struct {
		outcomes []Outcome
		input    Strength
		expected Strength
	}{
		{
			[]Outcome{
				{Mu: c.Rating2Mu(1400), Phi: c.Deviation2Phi(30), Score: 1},
				{Mu: c.Rating2Mu(1550), Phi: c.Deviation2Phi(100), Score: 0},
				{Mu: c.Rating2Mu(1700), Phi: c.Deviation2Phi(300), Score: 0},
			},
			Strength{
				Mu:    c.Rating2Mu(1500),
				Phi:   c.Deviation2Phi(200),
				Sigma: 0.06,
			},
			// Those values slighly (+- 1 rating point) differs from the original
			// Glicko-2 example.  This is because implementation processes each
			// game sequentially and the original algorithm is not linear.
			// The difference is negligable.
			Strength{
				Mu:    -0.20845068892693425,
				Phi:   0.8742521589251114,
				Sigma: 0.05999440707992557,
			},
		},
	}

	for _, tc := range testcases {
		e := Estimator{
			MinMu: c.Rating2Mu(10), MaxMu: c.Rating2Mu(5000),
			MinPhi: c.Deviation2Phi(50), MaxPhi: c.Deviation2Phi(DefaultDeviation),
			MinSigma: 0.04, MaxSigma: 0.08, Tau: DefaultTau, Epsilon: DefaultEpsilon,
		}

		for i := range tc.outcomes {
			e.Estimate(&tc.input, tc.outcomes[i], 1)
		}

		if tc.input != tc.expected {
			t.Fatalf("expected: %v, got: %v", tc.expected, tc.input)
		}
	}
}

func BenchmarkEstimate(b *testing.B) {
	c := Converter{
		Rating:    DefaultRating,
		Deviation: DefaultDeviation,
		Factor:    DefaultFactor,
	}

	s := Strength{
		Mu:    c.Rating2Mu(1500),
		Phi:   c.Deviation2Phi(350),
		Sigma: 0.06,
	}

	e := Estimator{
		MinMu: c.Rating2Mu(10), MaxMu: c.Rating2Mu(5000),
		MinPhi: c.Deviation2Phi(30), MaxPhi: c.Deviation2Phi(DefaultDeviation),
		MinSigma: 0.04, MaxSigma: 0.08, Tau: DefaultTau, Epsilon: DefaultEpsilon,
	}

	for b.Loop() {
		e.Estimate(&s, Outcome{
			Mu:    c.Rating2Mu(1400),
			Phi:   c.Deviation2Phi(30),
			Score: 1,
		}, 1)
	}
}
