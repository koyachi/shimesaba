package shimesaba_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Songmu/flextime"
	"github.com/mashiike/shimesaba"
	"github.com/stretchr/testify/require"
)

func TestAlertObjective(t *testing.T) {
	restore := flextime.Fix(time.Date(2021, time.October, 1, 0, 6, 0, 0, time.UTC))
	defer restore()
	alerts := shimesaba.Alerts{
		shimesaba.NewAlert(
			&shimesaba.Monitor{
				ID:   "hogera",
				Name: "SLO hoge",
				Type: "expression",
			},
			time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC),
			ptrTime(time.Date(2021, time.October, 1, 0, 3, 0, 0, time.UTC)),
		),
		shimesaba.NewAlert(
			&shimesaba.Monitor{
				ID:   "fugara",
				Name: "SLO fuga",
				Type: "service",
			},
			time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC),
			ptrTime(time.Date(2021, time.October, 1, 0, 4, 0, 0, time.UTC)),
		),
		shimesaba.NewAlert(
			&shimesaba.Monitor{
				ID:   "fugara",
				Name: "SLO fuga",
				Type: "service",
			},
			time.Date(2021, time.October, 1, 0, 3, 0, 0, time.UTC),
			ptrTime(time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC)),
		),
		shimesaba.NewAlert(
			&shimesaba.Monitor{
				ID:   "hogera",
				Name: "SLO hoge",
				Type: "expression",
			},
			time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC),
			nil,
		),
	}
	cases := []struct {
		cfg      *shimesaba.AlertObjectiveConfig
		expected map[time.Time]bool
	}{
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorID: "hogera",
			},
			expected: map[time.Time]bool{
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC): false,
			},
		},
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorNameSuffix: "hoge",
			},
			expected: map[time.Time]bool{
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC): false,
			},
		},
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorID: "fugara",
			},
			expected: map[time.Time]bool{
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC): true,
				time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC): true,
				time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 3, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 4, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC): true,
			},
		},
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorNamePrefix: "SLO",
			},
			expected: map[time.Time]bool{
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 3, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 4, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC): false,
			},
		},
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorNamePrefix: "SLO",
				MonitorType:       "Expression",
			},
			expected: map[time.Time]bool{
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC): false,
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC): false,
			},
		},
		{
			cfg: &shimesaba.AlertObjectiveConfig{
				MonitorNameSuffix: "hoge",
				MonitorType:       "service",
			},
			expected: map[time.Time]bool{},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case.%d", i), func(t *testing.T) {
			obj := shimesaba.NewAlertObjective(c.cfg)
			actual, err := obj.NewReliabilityCollection(
				time.Minute,
				alerts,
				time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC),
			)
			require.NoError(t, err)
			expected, _ := shimesaba.NewReliabilityCollection([]*shimesaba.Reliability{
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 0, 0, 0, time.UTC), time.Minute, c.expected),
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 1, 0, 0, time.UTC), time.Minute, c.expected),
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 2, 0, 0, time.UTC), time.Minute, c.expected),
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 3, 0, 0, time.UTC), time.Minute, c.expected),
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 4, 0, 0, time.UTC), time.Minute, c.expected),
				shimesaba.NewReliability(time.Date(2021, time.October, 1, 0, 5, 0, 0, time.UTC), time.Minute, c.expected),
			})
			require.EqualValues(t, expected, actual)
		})
	}

}

func ptrTime(t time.Time) *time.Time {
	return &t
}