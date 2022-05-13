package scheduler

import (
	"gotest.tools/assert"
	"testing"
	"time"
)

func Test_getStartsAt(t *testing.T) {
	timeNow, err := time.Parse(time.RFC3339, "2022-05-08T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow     time.Time
		startTime   time.Time
		intervalSec time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "start time is 10 seconds from now - next scan starts in 10 seconds from now",
			args: args{
				timeNow:     timeNow,
				startTime:   timeNow.Add(10 * time.Second),
				intervalSec: 5,
			},
			want: 10,
		},
		{
			name: "start time is 10 seconds before now, next scan should be now (5 + 5 = 10)",
			args: args{
				timeNow:     timeNow,
				startTime:   timeNow.Add(-10 * time.Second),
				intervalSec: 5,
			},
			want: 0,
		},
		{
			name: "start time is 13 seconds before now, next scan should be in 2 seconds (5 + 5 + 5 = 15 -> 15 -13 = 2)",
			args: args{
				timeNow:     timeNow,
				startTime:   timeNow.Add(-13 * time.Second),
				intervalSec: 5,
			},
			want: 2,
		},
		{
			name: "start time is 0.5 second before now, next scan should be now",
			args: args{
				timeNow:     timeNow,
				startTime:   timeNow.Add(500 * time.Millisecond),
				intervalSec: 5,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStartsAtSec(tt.args.timeNow, tt.args.startTime, tt.args.intervalSec); got != tt.want {
				t.Errorf("getStartsAtSec() = %v, want %v", got, tt.want)
			}
		})
	}
}
