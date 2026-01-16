/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package tinys3cli

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gammazero/workerpool"
)

type baseWorker struct {
	client    *s3.Client
	wp        *workerpool.WorkerPool
	mux       sync.Mutex
	lasterror error
}

func newBaseWorker(n_jobs int) (*baseWorker, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}
	return &baseWorker{
		client: client,
		wp:     workerpool.New(n_jobs),
		mux:    sync.Mutex{},
	}, nil
}

func (w *baseWorker) GetLastErr() error {
	return w.lasterror
}

func (w *baseWorker) SetLastErr(err error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.lasterror = err
}

func (w *baseWorker) Wait() {
	w.wp.StopWait()
}
