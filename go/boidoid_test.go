package main

import (
	"testing"
)


func Benchmark250(b *testing.B) {
	cfg := config{
		separationFactor:    0.05,
		alignmentFactor:     0.05,
		cohesionFactor:      0.0008,
		turnImpulse:         0.2,
		margin:              25,
		separationThreshold: 2,
		maxSpeed:            2.5,
		minSpeed:            1,
	}

	g := newGoggle(20, 250, 1000, 1000, cfg)
	for i := 0; i < b.N; i++ {
		g.updateFlocks(1000, 1000)
	}
}

func Benchmark500(b *testing.B) {
	cfg := config{
		separationFactor:    0.05,
		alignmentFactor:     0.05,
		cohesionFactor:      0.0008,
		turnImpulse:         0.2,
		margin:              25,
		separationThreshold: 2,
		maxSpeed:            2.5,
		minSpeed:            1,
	}

	g := newGoggle(20, 500, 1000, 1000, cfg)
	for i := 0; i < b.N; i++ {
		g.updateFlocks(1000, 1000)
	}
}
