package main

import (
	"log"
	"math"

	"github.com/rthornton128/goncurses"
	"golang.org/x/exp/rand"
)

type vec2 struct {
	x, y float64
}

func (v *vec2) add(rhs vec2) {
	v.x += rhs.x
	v.y += rhs.y
}

func (v *vec2) subtract(rhs vec2) {
	v.x -= rhs.x
	v.y -= rhs.y
}

func (v *vec2) divideScalar(scalar float64) {
	v.x /= scalar
	v.y /= scalar
}

func (v *vec2) multiplyScalar(scalar float64) {
	v.x *= scalar
	v.y *= scalar
}

func (lhs vec2) distance(rhs vec2) float64 {
	return math.Sqrt((lhs.x-rhs.x)*(lhs.x-rhs.x) + (lhs.y-rhs.y)*(lhs.y-rhs.y))
}

type config struct {
	separationFactor    float64
	alignmentFactor     float64
	cohesionFactor      float64
	turnImpulse         float64
	margin              float64
	separationThreshold float64
	maxSpeed            float64
	minSpeed            float64
}

type birdoid struct {
	pos     vec2
	moveDir vec2
	flock   []*birdoid
}

type goggle struct {
	config   config
	radius   float64
	birdoids []birdoid
}

func newGoggle(radius float64, numBirdoids, maxY, maxX int, conf config) goggle {
	var bs []birdoid
	for i := 0; i < numBirdoids; i++ {
		pos := vec2{rand.Float64() * float64(maxY),
			float64(maxX)/2 - float64(maxY)/2 + rand.Float64()*float64(maxY)}
		moveDir := vec2{rand.Float64() * conf.minSpeed, rand.Float64() * conf.minSpeed}

		bs = append(bs, birdoid{
			pos:     pos,
			moveDir: moveDir,
		})
	}

	return goggle{
		config:   conf,
		radius:   radius,
		birdoids: bs,
	}
}

func (g *goggle) updateFlocks(maxY, maxX int) {
	for i := 0; i < len(g.birdoids); i++ {
		g.birdoids[i].flock = []*birdoid{}
	}

	for i := 0; i < len(g.birdoids); i++ {
		for j := i + 1; j < len(g.birdoids); j++ {
			if g.birdoids[i].pos.distance(g.birdoids[j].pos) < g.radius {
				g.birdoids[i].flock =
					append(g.birdoids[i].flock, &g.birdoids[j])
				g.birdoids[j].flock =
					append(g.birdoids[j].flock, &g.birdoids[i])
			}
		}
	}

	for i := 0; i < len(g.birdoids); i++ {
		b := &g.birdoids[i]
		invClose := vec2{0, 0}
		avgHeading := vec2{0, 0}
		avgPosition := vec2{0, 0}

		for i := 0; i < len(b.flock); i++ {
			if (b.pos.distance(b.flock[i].pos)) < g.config.separationThreshold {
				// neighbor direction for separation
				invClose.x += b.pos.x - b.flock[i].pos.x
				invClose.y += b.pos.y - b.flock[i].pos.y
			}

			// neighbor position to calculate cohesion
			avgPosition.add(b.flock[i].pos)

			// avg neighbor heading for alignment
			avgHeading.add(b.flock[i].moveDir)
		}

		invClose.multiplyScalar(g.config.separationFactor)

		if len(b.flock) > 0 {
			avgHeading.divideScalar(float64(len(b.flock)))
			avgHeading.multiplyScalar(g.config.alignmentFactor)

			avgPosition.divideScalar(float64(len(b.flock)))
			// get movement vector by subtracting current position from avg neighbor position
			avgPosition.subtract(b.pos)
			avgPosition.multiplyScalar(g.config.cohesionFactor)
		}

		b.moveDir.add(avgHeading)
		b.moveDir.add(avgPosition)
		b.moveDir.add(invClose)

		if b.pos.x < g.config.margin {
			b.moveDir.x += g.config.turnImpulse
		}
		if b.pos.x > float64(maxX)-g.config.margin {
			b.moveDir.x -= g.config.turnImpulse
		}
		if b.pos.y < g.config.margin {
			b.moveDir.y += g.config.turnImpulse
		}
		if b.pos.y > float64(maxY)-g.config.margin {
			b.moveDir.y -= g.config.turnImpulse
		}

		speed := b.moveDir.distance(vec2{0, 0})
		if speed > g.config.maxSpeed {
			b.moveDir.divideScalar(speed)
			b.moveDir.multiplyScalar(g.config.maxSpeed)
		} else if speed < g.config.minSpeed {
			b.moveDir.divideScalar(speed)
			b.moveDir.multiplyScalar(g.config.minSpeed)
		}

		// finally update position
		b.pos.add(b.moveDir)
	}
}

func draw(g *goggle, win *goncurses.Window) {
	win.Erase()

	for _, b := range g.birdoids {
		win.MoveAddChar(int(b.pos.y), int(b.pos.x), 'o')
	}
}

func main() {
	win, err := goncurses.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	defer goncurses.End()

	goncurses.Echo(false)
	goncurses.Cursor(0)
	goncurses.HalfDelay(1)

	maxY, maxX := win.MaxYX()
	conf := config{
		separationFactor:    0.05,
		alignmentFactor:     0.05,
		cohesionFactor:      0.0008,
		turnImpulse:         0.2,
		margin:              25,
		separationThreshold: 2,
		maxSpeed:            2.5,
		minSpeed:            1,
	}
	g := newGoggle(20, 250, maxY, maxX, conf)

	for {
		maxY, maxX = win.MaxYX()
		draw(&g, win)
		g.updateFlocks(maxY, maxX)

		if win.GetChar() == 'q' {
			break
		}
	}
}
