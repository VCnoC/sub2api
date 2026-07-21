package service

import "testing"

func TestRequirementSatisfied(t *testing.T) {
	tests := []struct {
		name        string
		recharge    float64
		spend       float64
		requirement TeamLevelRequirement
		want        bool
	}{
		{name: "and both met", recharge: 100, spend: 20, requirement: TeamLevelRequirement{Recharge: 100, Spend7Days: 20, Mode: "and"}, want: true},
		{name: "and only recharge met", recharge: 100, spend: 19, requirement: TeamLevelRequirement{Recharge: 100, Spend7Days: 20, Mode: "and"}, want: false},
		{name: "or recharge met", recharge: 100, spend: 0, requirement: TeamLevelRequirement{Recharge: 100, Spend7Days: 20, Mode: "or"}, want: true},
		{name: "or neither met", recharge: 99, spend: 19, requirement: TeamLevelRequirement{Recharge: 100, Spend7Days: 20, Mode: "or"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := requirementSatisfied(test.recharge, test.spend, test.requirement); got != test.want {
				t.Fatalf("requirementSatisfied() = %v, want %v", got, test.want)
			}
		})
	}
}
