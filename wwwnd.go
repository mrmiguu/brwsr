package wwwnd

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/jsutil"
)

var (
	phaser *js.Object
	window *Window
	ready  = make(chan bool, 1)
)

type Window struct {
	width, height int

	game             *js.Object
	load             *js.Object
	add              *js.Object
	centerx, centery int

	images struct {
		sync.RWMutex
		m map[string]*Image
	}
}

type Image struct {
	js struct {
		sync.RWMutex
		o *js.Object
	}
}

func init() {
	style := js.Global.Get("document").Get("body").Get("style")
	style.Set("background", "#000000")
	style.Set("margin", 0)

	<-jsutil.Load(
		"phaser.min.js",
		"phaser.js",
		"lib/phaser.min.js",
		"lib/phaser.js",
		"js/phaser.min.js",
		"js/phaser.js",
		"https://github.com/photonstorm/phaser-ce/releases/download/v2.8.5/phaser.min.js",
	)
	phaser = js.Global.Get("Phaser")
}

func New(width, height int) *Window {
	if window != nil {
		panic("window already created")
	}
	window = &Window{
		width:  width,
		height: height,
		game:   phaser.Get("Game").New(width, height, phaser.Get("AUTO"), "", js.M{"preload": preload, "create": create}),
	}
	window.images.m = map[string]*Image{}

	<-ready
	return window
}

func preload() {
	window.game.Get("canvas").Set("oncontextmenu", func(e *js.Object) { e.Call("preventDefault") })

	scale := window.game.Get("scale")
	mode := phaser.Get("ScaleManager").Get("SHOW_ALL")
	scale.Set("scaleMode", mode)
	scale.Set("fullScreenScaleMode", mode)
	scale.Set("pageAlignHorizontally", true)
	scale.Set("pageAlignVertically", true)

	window.load = window.game.Get("load")
	// window.load.Call("spritesheet", "loading", "res/loading.png", window.width, window.height)
}

func create() {
	window.add = window.game.Get("add")
	world := window.game.Get("world")
	window.centerx = world.Get("centerX").Int()
	window.centery = world.Get("centerY").Int()

	// loading = window.NewImage("loading")
	// loading.Set("visible", true)
	// loading.Set("alpha", 0)
	// fadeIn := newTween(loading, js.M{"alpha": 1}, 1333)
	// fadeIn.Call("start")

	// bg = "res/bg.png"
	// taptostart = "res/taptostart.png"
	// username = "res/username.png"

	// onLoad, _ := jsutil.Callback()

	window.load.Get("onFileComplete").Call("add", func(_, key *js.Object) {
		window.addImage(key.String())
	})

	// spin = loading.Get("animations").Call("add", "spin")
	// spin.Call("play", 9, true)
	// onSpin, _ := jsutil.Callback()
	// spin.Get("onLoop").Call("add", onSpin)

	// window.load.Call("image", bg, bg)
	// window.load.Call("spritesheet", taptostart, taptostart, w, h)
	// window.load.Call("spritesheet", username, username, 360, 216)

	ready <- true
	close(ready)
}

func (w *Window) addImage(key string) {
	o := w.add.Call("sprite", w.centerx, w.centery, key)
	o.Get("anchor").Call("setTo", 0.5, 0.5)

	w.images.Lock()
	defer w.images.Unlock()

	i := w.images.m[key]
	o.Set("width", i.js.o.Get("width"))
	o.Set("height", i.js.o.Get("height"))

	i.js.Lock()
	i.js.o = o
	i.js.Unlock()
}

// func newTween(o *js.Object, params js.M, ms int) *js.Object {
// 	twn := window.add.Call("tween", o).Call("to", params, ms)
// 	twn.Set("frameBased", true)
// 	return twn
// }

func (w *Window) NewImage(url string, width, height int) *Image {
	var i Image
	i.js.o = phaser.Get("Rectangle").New(
		w.width/2-width/2, w.height/2-height/2,
		width, height,
	)

	w.images.Lock()
	defer w.images.Unlock()

	if _, exists := w.images.m[url]; exists {
		w.addImage(url)
		return &i
	}

	w.images.m[url] = &i

	w.load.Call("image", url, url)
	w.load.Call("start")

	return &i
}

// func newButton(url string) (*js.Object, <-chan bool) {
// 	x, y := world.Get("centerX").Int(), world.Get("centerY").Int()
// 	onHit, hit := jsutil.Callback()
// 	btn := window.game.Get("add").Call("button", x, y, url, onHit, nil, 0, 0, 1, 0)
// 	btn.Set("visible", false)
// 	btn.Get("anchor").Call("setTo", 0.5, 0.5)
// 	btn.Get("onInputDown").Call("add", func() { btn.Set("y", y+util.Min(h-btn.Get("h").Int(), 3)) })
// 	btn.Get("onInputOver").Call("add", func() { btn.Set("y", y) })
// 	btn.Get("onInputOut").Call("add", func() { btn.Set("y", y) })
// 	btn.Get("onInputUp").Call("add", func() { btn.Set("y", y) })
// 	return btn, hit
// }
