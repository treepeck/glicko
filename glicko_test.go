package glicko

import "testing"

func TestEvaluatePlayer(t *testing.T) {
	testcases := []struct {
		outcomes []Outcome
		current  Evaluation
		expected Evaluation
	}{
		{
			[]Outcome{
				NewOutcome(1500, 1400, 30, 1),
				NewOutcome(1500, 1550, 100, 0),
				NewOutcome(1500, 1700, 300, 0),
			},
			Evaluation{Rating: 1500, Deviation: 200, Volatility: 0.06},
			Evaluation{Rating: 1464.050663079054, Deviation: 151.51653984530088,
				Volatility: 0.06},
		},
	}

	for _, tc := range testcases {
		got := EvaluatePlayer(tc.current, tc.outcomes)

		if got != tc.expected {
			t.Fatalf("expected: %v, got: %v", tc.expected, got)
		}
	}
}

func BenchmarkEvaluatePlayer(b *testing.B) {
	curr := Evaluation{Rating: 1500, Deviation: 200, Volatility: 0.06}
	outcomes := []Outcome{
		NewOutcome(1500, 1400, 30, 1),
		NewOutcome(1500, 1550, 100, 0),
		NewOutcome(1500, 1700, 300, 0),
	}

	for b.Loop() {
		EvaluatePlayer(curr, outcomes)
	}
}
