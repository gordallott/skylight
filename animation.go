package main

import (
	"sync/atomic"
	"time"
)

var handlerCount uint64

type Animation struct {
	Handler   uint64
	Repeating bool
	Length    time.Duration

	callback func(pos float64) bool
	easing   func(pos float64) float64
	start    time.Time
}

func defaultEasing(pos float64) float64 {
	return pos
}

func NewAnimationWithTickrate(callback func(pos float64) bool, tickRate time.Duration) *Animation {
	animation := &Animation{
		Handler:  atomic.AddUint64(&handlerCount, 1),
		easing:   defaultEasing,
		start:    time.Now(),
		callback: callback,
	}

	ticker := time.NewTicker(tickRate)
	go func() {
		for now := range ticker.C {
			var pos float64 = float64(now.Sub(animation.start))

			if animation.Repeating {
				pos = float64(time.Duration(pos) % animation.Length)
			}
			if pos > 1.0 {
				ticker.Stop()
				break
			}

			pos = pos / float64(animation.Length)
			cont := animation.callback(animation.easing(pos))
			if !cont {
				ticker.Stop()
				break
			}
		}
	}()

	return animation
}

// NewAnimation ...
func NewAnimation(callback func(pos float64) bool) *Animation {
	return NewAnimationWithTickrate(callback, time.Millisecond*100)
}
