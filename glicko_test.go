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
				NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1400), c.Deviation2Phi(30), 1),
				NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1550), c.Deviation2Phi(100), 0),
				NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1700), c.Deviation2Phi(300), 0),
			},
			Strength{
				Mu:    c.Rating2Mu(1500),
				Phi:   c.Deviation2Phi(200),
				Sigma: 0.06,
			},
			Strength{
				Mu:    -0.20694100961988987,
				Phi:   0.8721992786306347,
				Sigma: 0.06,
			},
		},
	}

	for _, tc := range testcases {
		e := Estimator{
			MinMu: c.Rating2Mu(10), MaxMu: c.Rating2Mu(5000),
			MinPhi: c.Deviation2Phi(50), MaxPhi: c.Deviation2Phi(DefaultDeviation),
			MinSigma: 0.04, MaxSigma: 0.08, Tau: DefaultTau, Epsilon: DefaultEpsilon,
		}

		e.Estimate(&tc.input, tc.outcomes)

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

	outcomes := []Outcome{
		NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1400), c.Deviation2Phi(30), 1),
		NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1550), c.Deviation2Phi(100), 0),
		NewOutcome(c.Rating2Mu(1500), c.Rating2Mu(1700), c.Deviation2Phi(300), 0),
	}

	s := Strength{
		Mu:    c.Rating2Mu(1500),
		Phi:   c.Deviation2Phi(200),
		Sigma: 0.06,
	}

	e := Estimator{
		MinMu: c.Rating2Mu(10), MaxMu: c.Rating2Mu(5000),
		MinPhi: c.Deviation2Phi(50), MaxPhi: c.Deviation2Phi(DefaultDeviation),
		MinSigma: 0.04, MaxSigma: 0.08, Tau: DefaultTau, Epsilon: DefaultEpsilon,
	}

	for b.Loop() {
		e.Estimate(&s, outcomes)
	}
}
