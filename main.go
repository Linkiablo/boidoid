package main

import (
	"log"
	"math"

	"github.com/rthornton128/goncurses"
	"golang.org/x/exp/rand"
)

// TODO: config interactive
// TODO: find a more general configuration, independent of terminal size
const SEPARATIONFACTOR = 0.05
const ALIGNMENTFACTOR = 0.05
const COHESIONFACTOR = 0.0005
const TURNIMPULSE = 0.2
const MARGIN = 50
const SEPARATIONTHRESHOLD = 2
const MAXSPEED = 3
const MINSPEED = 1

func clamp(val, min, max float64) float64 {
	return math.Min(math.Max(val, min), max)
}

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

type birdoid struct {
	pos     vec2
	moveDir vec2
	flock   []*birdoid
}

func (b *birdoid) update(maxY, maxX int) {
	invClose := vec2{0, 0}
	avgHeading := vec2{0, 0}
	avgPosition := vec2{0, 0}

	for i := 0; i < len(b.flock); i++ {
		if (b.pos.distance(b.flock[i].pos)) < SEPARATIONTHRESHOLD {
			// neighbor direction for separation
			invClose.x += b.pos.x - b.flock[i].pos.x
			invClose.y += b.pos.y - b.flock[i].pos.y
		}

		// neighbor position to calculate cohesion
		avgPosition.x += b.flock[i].pos.x
		avgPosition.y += b.flock[i].pos.y

		// avg neighbor heading for alignment
		avgHeading.x += b.flock[i].moveDir.x
		avgHeading.y += b.flock[i].moveDir.y

	}

	invClose.multiplyScalar(SEPARATIONFACTOR)

	if len(b.flock) > 0 {
		avgHeading.divideScalar(float64(len(b.flock)))
		avgHeading.multiplyScalar(ALIGNMENTFACTOR)

		avgPosition.divideScalar(float64(len(b.flock)))
		// get movement vector by subtracting current position from avg neighbor position
		avgPosition.subtract(b.pos)
		avgPosition.multiplyScalar(COHESIONFACTOR)
	}

	b.moveDir.add(avgHeading)
	b.moveDir.add(avgPosition)
	b.moveDir.add(invClose)

	// logFd, _ := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// defer logFd.Close()
	// fmt.Fprintln(logFd, b)

	if b.pos.x < MARGIN {
		b.moveDir.x += TURNIMPULSE
	}
	if b.pos.x > float64(maxX-MARGIN) {
		b.moveDir.x -= TURNIMPULSE
	}
	if b.pos.y < MARGIN {
		b.moveDir.y += TURNIMPULSE
	}
	if b.pos.y > float64(maxY-MARGIN) {
		b.moveDir.y -= TURNIMPULSE
	}

	speed := b.moveDir.distance(vec2{0, 0})
	if speed > MAXSPEED {
		b.moveDir.divideScalar(speed)
		b.moveDir.multiplyScalar(MAXSPEED)
	} else if speed < MINSPEED {
		b.moveDir.divideScalar(speed)
		b.moveDir.multiplyScalar(MINSPEED)
	}

	// finally update position
	b.pos.add(b.moveDir)
}

type goggle struct {
	radius   float64
	birdoids []birdoid
}

func newGoggle(radius float64, numBirdoids, maxY, maxX int) goggle {
	var bs []birdoid
	for i := 0; i < numBirdoids; i++ {
		pos := vec2{rand.Float64() * float64(maxX), rand.Float64() * float64(maxY)}
		moveDir := vec2{rand.Float64() * MINSPEED, rand.Float64() * MINSPEED}

		bs = append(bs, birdoid{
			pos:     pos,
			moveDir: moveDir,
		})
	}

	return goggle{
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
		g.birdoids[i].update(maxY, maxX)
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
	g := newGoggle(20, 250, maxY, maxX)
	for {
		maxY, maxX = win.MaxYX()
		draw(&g, win)
		g.updateFlocks(maxY, maxX)

		if win.GetChar() == 'q' {
			break
		}
	}
}
