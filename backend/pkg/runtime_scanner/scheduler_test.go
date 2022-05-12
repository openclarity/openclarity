package runtime_scanner

import (
	"gotest.tools/assert"
	"testing"
	"time"

	"github.com/openclarity/kubeclarity/backend/pkg/database"
)

func Test_calculateNextScanTime(t *testing.T) {
	timeNow, err := time.Parse(time.RFC3339, "2022-05-08T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow time.Time
		s       *database.Scheduler
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "interval has not been pass since last scan time",
			args: args{
				timeNow: timeNow,
				s: &database.Scheduler{
					LastScanTime: timeNow.Add(-5 * time.Second).Format(time.RFC3339),
					StartTime:    "",
					Interval:     6,
				},
			},
			want:    timeNow.Add(1 * time.Second),
			wantErr: false,
		},
		{
			name: "interval has been pass since last scan time",
			args: args{
				timeNow: timeNow,
				s: &database.Scheduler{
					LastScanTime: timeNow.Add(-10 * time.Second).Format(time.RFC3339),
					StartTime:    "",
					Interval:     6,
				},
			},
			want:    timeNow.Add(2 * time.Second),
			wantErr: false,
		},
		{
			name: "there wasn't any scan, and start time has not been reached yet",
			args: args{
				timeNow: timeNow,
				s: &database.Scheduler{
					LastScanTime: "",
					StartTime:    timeNow.Add(2 * time.Second).Format(time.RFC3339),
					Interval:     6,
				},
			},
			want:    timeNow.Add(2 * time.Second),
			wantErr: false,
		},
		{
			name: "there wasn't any scan, and start time has been pass",
			args: args{
				timeNow: timeNow,
				s: &database.Scheduler{
					LastScanTime: "",
					StartTime:    timeNow.Add(-20 * time.Second).Format(time.RFC3339),
					Interval:     6,
				},
			},
			want:    timeNow.Add(4 * time.Second),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateNextScanTimeOnStart(tt.args.timeNow, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateNextScanTimeOnStart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("calculateNextScanTimeOnStart() got = %v, want %v", got, tt.want)
			}
		})
	}
}
