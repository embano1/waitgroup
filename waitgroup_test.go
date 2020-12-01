package waitgroup

import (
	"errors"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// timeout used in waiters
	timeout = time.Millisecond * 10
)

var (
	fastFn = func(wg *sync.WaitGroup) {
		wg.Done()
	}

	slowFn = func(wg *sync.WaitGroup) {
		defer wg.Done()
		// simulate work
		time.Sleep(timeout)
	}

	neverStopFn = func(wg *sync.WaitGroup) {
		// hopefully we'll never see this in prod :)
		// we intentionally don't call wg.Done() to simulate a bug
		forever := make(chan struct{})
		<-forever
	}
)

func TestWaitGroup_WaitTimeout(t *testing.T) {
	type fields struct {
		wg     *WaitGroup
		worker []func(wg *sync.WaitGroup)
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
				worker: []func(wg *sync.WaitGroup){fastFn},
			},
			args: args{
				timeout: timeout / 2,
			},
			wantErr: nil,
		},
		{
			name: "slow worker exceeds timeout",
			fields: fields{
				wg:     &WaitGroup{},
				worker: []func(wg *sync.WaitGroup){slowFn},
			},
			args: args{
				timeout: timeout / 2,
			},
			wantErr: ErrTimeout,
		},
		{
			name: "buggy worker never returns",
			fields: fields{
				wg:     &WaitGroup{},
				worker: []func(wg *sync.WaitGroup){neverStopFn},
			},
			args: args{
				timeout: timeout / 2,
			},
			wantErr: ErrTimeout,
		},
		{
			name: "one fast one slow worker exceeding timeout",
			fields: fields{
				wg:     &WaitGroup{},
				worker: []func(wg *sync.WaitGroup){fastFn, slowFn},
			},
			args: args{
				timeout: timeout / 2,
			},
			wantErr: ErrTimeout,
		},
		{
			name: "one fast one slow worker which never returns",
			fields: fields{
				wg:     &WaitGroup{},
				worker: []func(wg *sync.WaitGroup){fastFn, neverStopFn},
			},
			args: args{
				timeout: timeout / 2,
			},
			wantErr: ErrTimeout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := tt.fields.wg
			wg.Add(len(tt.fields.worker))
			// spawn worker(s)
			for _, w := range tt.fields.worker {
				go w(&wg.WaitGroup)
			}

			if err := tt.fields.wg.WaitTimeout(tt.args.timeout); err != tt.wantErr {
				t.Errorf("WaitTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAwait(t *testing.T) {
	type args struct {
		wf      Waiter
		timeout time.Duration
	}
	tests := []struct {
		name    string
		worker  []func(wg *sync.WaitGroup)
		args    args
		wantErr error
	}{
		{
			name:    "sync.WaitGroup with one fast worker",
			worker:  []func(wg *sync.WaitGroup){fastFn},
			args:    args{wf: &sync.WaitGroup{}, timeout: timeout / 2},
			wantErr: nil,
		},
		{
			name:    "sync.WaitGroup with one slow worker exceeding timeout",
			worker:  []func(wg *sync.WaitGroup){slowFn},
			args:    args{wf: &sync.WaitGroup{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "sync.WaitGroup with one worker which never returns",
			worker:  []func(wg *sync.WaitGroup){neverStopFn},
			args:    args{wf: &sync.WaitGroup{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "sync.WaitGroup with one fast and one slow worker exceeding timeout",
			worker:  []func(wg *sync.WaitGroup){fastFn, slowFn},
			args:    args{wf: &sync.WaitGroup{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "sync.WaitGroup with one fast and one worker which never returns",
			worker:  []func(wg *sync.WaitGroup){fastFn, neverStopFn},
			args:    args{wf: &sync.WaitGroup{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := tt.args.wf.(*sync.WaitGroup)
			wg.Add(len(tt.worker))
			// spawn worker(s)
			for _, w := range tt.worker {
				go w(wg)
			}

			if err := Await(tt.args.wf, tt.args.timeout); err != tt.wantErr {
				t.Errorf("WaitTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAwaitWithError(t *testing.T) {
	var (
		testErr = errors.New("error")

		fastFn = func() error {
			return nil
		}

		slowFn = func() error {
			// simulate work
			time.Sleep(timeout)
			return nil
		}

		errFn = func() error {
			return testErr
		}

		slowErrFn = func() error {
			// simulate work
			time.Sleep(timeout)
			return testErr
		}
	)

	type args struct {
		wf      WaitErrorer
		timeout time.Duration
	}
	tests := []struct {
		name    string
		worker  []func() error
		args    args
		wantErr error
	}{
		{
			name:    "ErrorGroup with one fast worker",
			worker:  []func() error{fastFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: nil,
		},
		{
			name:    "ErrorGroup with one slow worker exceeding timeout",
			worker:  []func() error{slowFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "ErrorGroup with one worker returning error",
			worker:  []func() error{errFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: testErr,
		},
		{
			name:    "ErrorGroup with one slow worker returning error",
			worker:  []func() error{slowErrFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "ErrorGroup with one fast and one slow worker exceeding timeout",
			worker:  []func() error{fastFn, slowFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
		{
			name:    "ErrorGroup with one fast and one worker returning error",
			worker:  []func() error{fastFn, errFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: testErr,
		},
		{
			name:    "ErrorGroup with one fast and one slow worker returning error",
			worker:  []func() error{fastFn, slowErrFn},
			args:    args{wf: &errgroup.Group{}, timeout: timeout / 2},
			wantErr: ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eg := tt.args.wf.(*errgroup.Group)

			// spawn worker(s)
			for _, w := range tt.worker {
				eg.Go(w)
			}

			if err := AwaitWithError(tt.args.wf, tt.args.timeout); err != tt.wantErr {
				t.Errorf("WaitTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
