package waitgroup

import (
	"sync"
	"testing"
	"time"
)

func TestWaitGroup_WaitTimeout(t *testing.T) {
	const (
		delayFn = time.Millisecond * 10
	)

	fastFn := func(wg *sync.WaitGroup) {
		wg.Done()
	}

	slowFn := func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(delayFn)
	}

	// hopefully we'll never see this in prod :)
	/*	neverStopFn := func(wg *sync.WaitGroup) {
		forever := make(chan struct{})
		<-forever
	}*/

	type fields struct {
		wg     *WaitGroup
		worker func(wg *sync.WaitGroup)
	}
	type args struct {
		timeout time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "fast worker finishes in time",
			fields: fields{
				wg:     &WaitGroup{},
				worker: fastFn,
			},
			args: args{
				timeout: delayFn / 2,
			},
			wantErr: nil,
		},
		{
			name: "slow worker takes too long",
			fields: fields{
				wg:     &WaitGroup{},
				worker: slowFn,
			},
			args: args{
				timeout: delayFn / 2,
			},
			wantErr: ErrTimeout,
		},
		/*		{
				name: "buggy worker never returns",
				fields: fields{
					wg:     &WaitGroup{},
					worker: neverStopFn,
				},
				args: args{
					timeout: delayFn / 2,
				},
				wantErr: ErrTimeout,
			},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := tt.fields.wg
			wg.Add(1)
			// spawn worker
			go tt.fields.worker(&wg.WaitGroup)

			if err := tt.fields.wg.WaitTimeout(tt.args.timeout); err != tt.wantErr {
				t.Errorf("WaitTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
