package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marsmay/golib/logger"
)

type Point struct {
	signal bool
	t      time.Time
}

type Worker struct {
	provider  IProvider
	logger    *logger.Logger
	ctx       context.Context
	loopTimer chan *Point
	endSign   chan bool
}

func (w *Worker) SendSign(signal string) (err error) {
	signTime, e := time.ParseInLocation(SignalFormat, strings.TrimSpace(signal), time.Local)

	if e != nil {
		return fmt.Errorf("[%s] signal fotmat error: '%s'", w.provider.GetName(), signal)
	}

	select {
	case <-w.ctx.Done():
		err = fmt.Errorf("[%s] worker is closed", w.provider.GetName())
	case w.loopTimer <- &Point{true, signTime}:
		w.logger.Infof("[%s] receive signal: %s", w.provider.GetName(), signal)
	default:
		err = fmt.Errorf("[%s] worker is busy", w.provider.GetName())
	}

	return
}

func (w *Worker) startLoop() {
	if second := time.Now().Second(); second > 0 {
		time.Sleep(time.Minute - time.Duration(second)*time.Second)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case t := <-ticker.C:
			if w.provider.CheckInterval(t) {
				w.loopTimer <- &Point{false, t}
			}
		}
	}
}

func (w *Worker) Run() {
	defer func() {
		w.logger.Infof("[%s] worker end", w.provider.GetName())
		w.endSign <- true
	}()

	w.logger.Infof("[%s] worker start", w.provider.GetName())
	go w.startLoop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case p := <-w.loopTimer:
			if p.signal {
				w.logger.Infof("[%s] run by signal", w.provider.GetName())
			}

			w.provider.Run(p.t)
		}
	}
}

func (w *Worker) Done() {
	<-w.endSign
}

func NewWorker(p IProvider, l *logger.Logger, ctx context.Context) (worker *Worker, err error) {
	if err = p.Init(); err != nil {
		return
	}

	worker = &Worker{
		provider:  p,
		logger:    l,
		ctx:       ctx,
		loopTimer: make(chan *Point, 1),
		endSign:   make(chan bool, 1),
	}
	return
}
