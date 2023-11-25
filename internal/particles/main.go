package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
)

var (
	canvas    js.Value
	ctx       js.Value
	document  js.Value
	colors    []*color
	gray      *color
	mouse     cursor
	particles []*particle
	done      chan bool
)

type color struct {
	r int
	g int
	b int
}

func (c *color) rgba(opacity float64) string {
	return fmt.Sprintf("rgba(%d, %d, %d, %f)", c.r, c.g, c.b, opacity)
}

type cursor struct {
	active bool
	x      float64
	y      float64
	radius float64
}

type particle struct {
	x          float64
	y          float64
	directionX float64
	directionY float64
	size       float64
	c          *color
}

func (p *particle) draw() {
	ctx.Call("beginPath")
	ctx.Call("arc", p.x, p.y, p.size, 0, math.Pi*2)
	ctx.Set("fillStyle", p.c.rgba(1))
	ctx.Call("fill")
}

func (p *particle) update() {
	canvW := canvas.Get("width").Float()
	canvH := canvas.Get("height").Float()
	// Reverse direction if particle has hit a window border.
	if (p.x+p.size) > canvW || (p.x-p.size) < 0 {
		p.directionX = -p.directionX
	}
	if (p.y+p.size) > canvH || (p.y-p.size) < 0 {
		p.directionY = -p.directionY
	}
	// Check distance from particle to mouse radius.
	dx := mouse.x - p.x
	dy := mouse.y - p.y
	dist := math.Sqrt(dx*dx + dy*dy)
	if mouse.active && dist < mouse.radius+p.size {
		// Particle is right of mouse.
		if mouse.x < p.x && p.x < canvW-p.size*10 {
			p.x += 10
			// Reverse direction if approaching from right.
			if p.directionX < 0 {
				p.directionX = -p.directionX
			}
		}
		// Particle is left of mouse.
		if mouse.x > p.x && p.x > p.size*10 {
			p.x -= 10
			// Reverse direction if approaching from left.
			if p.directionX > 0 {
				p.directionX = -p.directionX
			}
		}
		// Particle is above mouse.
		if mouse.y > p.y && p.y > p.size*10 {
			p.y -= 10
			// Reverse if approaching from top.
			if p.directionY > 0 {
				p.directionY = -p.directionY
			}
		}
		// Particle is below mouse.
		if mouse.y < p.y && p.y < canvH-p.size*10 {
			p.y += 10
			// Reverse if approaching from bottom.
			if p.directionY < 0 {
				p.directionY = -p.directionY
			}
		}
	}
	// Update particle position & draw.
	p.x += p.directionX
	p.y += p.directionY
	p.draw()
}

func newRandomParticle() *particle {
	size := rand.Float64()*5 + 1
	canvW := canvas.Get("width").Float()
	canvH := canvas.Get("height").Float()
	colorIdx := rand.Intn(len(colors))

	return &particle{
		x:          rand.Float64()*((canvW-size*2)-size*2) + size*2,
		y:          rand.Float64()*((canvH-size*2)-size*2) + size*2,
		directionX: rand.Float64()*5 - 2.5,
		directionY: rand.Float64()*5 - 2.5,
		size:       size,
		c:          colors[colorIdx],
	}
}

func init() {
	// Create "done" channel -- keeps thread from exiting early.
	done = make(chan bool)
	// Grab documment and canvas element.
	document = js.Global().Get("document")
	canvas = document.Call("getElementById", "canvas")
	// Get canvas 2d context.
	ctx = canvas.Call("getContext", "2d")
	// Add particle colors.
	pink := &color{r: 244, g: 143, b: 177}
	blue := &color{r: 0, g: 204, b: 204}
	gray = &color{r: 20, g: 20, b: 20}
	colors = append(colors, pink, blue)
	// Fit canvas to window size, populate particles.
	fitCanvasAndInitParticles()
	bindEventListeners()
}

// Animate sets up the main animation loop.
func animate() {
	var lastTimestamp float64
	var timeStep float64 = float64(1000) / 60
	var renderFrame js.Func
	renderFrame = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		timestamp := args[0].Float()
		if timestamp-lastTimestamp >= timeStep {
			canvW := canvas.Get("width").Int()
			canvH := canvas.Get("height").Int()
			ctx.Call("clearRect", 0, 0, canvW, canvH)
			for _, p := range particles {
				p.update()
			}
		}
		js.Global().Call("requestAnimationFrame", renderFrame)
		connect()
		return nil
	})
	defer renderFrame.Release()
	js.Global().Call("requestAnimationFrame", renderFrame)
	<-done
}

// Draw connections between particles based on distance between particles.
func connect() {
	canvW := canvas.Get("width").Float()
	canvH := canvas.Get("height").Float()
	// For each particle, loop through all other particles.
	for a := 0; a < len(particles)-1; a++ {
		for b := a + 1; b < len(particles); b++ {
			partA := particles[a]
			partB := particles[b]
			dx := partA.x - partB.x
			dy := partA.y - partB.y
			dist := (dx*dx + dy*dy)
			// Connect if within distance threshhold.
			if dist < (canvW / 14 * canvH / 14) {
				var lineColor string
				// Decrease opacity as particles separate.
				opacity := 1 - dist/10000
				// Use particle color if colors match, else gray.
				if partA.c == partB.c {
					lineColor = partA.c.rgba(opacity)
				} else {
					lineColor = gray.rgba(opacity)
				}
				ctx.Set("strokeStyle", lineColor)
				ctx.Set("lineWidth", 1)
				ctx.Call("beginPath")
				ctx.Call("moveTo", partA.x, partA.y)
				ctx.Call("lineTo", partB.x, partB.y)
				ctx.Call("stroke")
			}
		}
	}
}

func bindEventListeners() {
	// Set mouse position on mousemove.
	mouseCb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ev := args[0]
		mouse.active = true
		mouse.x = ev.Get("x").Float()
		mouse.y = ev.Get("y").Float()
		return nil
	})
	js.Global().Call("addEventListener", "mousemove", mouseCb)
	// Remove mouse 'obstacle' on mouseout.
	mouseoutCb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		mouse.active = false
		return nil
	})
	js.Global().Call("addEventListener", "mouseout", mouseoutCb)
	// Re-init on window resize.
	resizeCb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fitCanvasAndInitParticles()
		return nil
	})
	js.Global().Call("addEventListener", "resize", resizeCb)
}

func fitCanvasAndInitParticles() {
	windowW := js.Global().Get("innerWidth").Int()
	windowH := js.Global().Get("innerHeight").Int()
	canvas.Set("width", windowW)
	canvas.Set("height", windowH)
	// Scale mouse radius relative to window size.
	mouse.radius = (float64(windowW) / 120) * float64((windowH)/120)
	// Reset particles array, populate proportional to canvas size.
	particles = []*particle{}
	numParticles := (windowW * windowH) / 9000
	for i := 0; i < numParticles; i++ {
		particles = append(particles, newRandomParticle())
	}
}

func main() {
	animate()
}
