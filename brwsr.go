package brwsr

import (
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/jsutil"
)

var (
	phaser  *js.Object
	browser *Browser
	ready   = make(chan bool, 1)
	fps     = time.Tick(250 * time.Millisecond)
)

type Browser struct {
	width, height int

	game             *js.Object
	load             *js.Object
	add              *js.Object
	centerx, centery int

	images struct {
		sync.RWMutex
		m map[string]struct {
			i     *Image
			donut *js.Object
			anim  *js.Object
		}
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

func New(width, height int) *Browser {
	if browser != nil {
		panic("browser already created")
	}
	browser = &Browser{
		width:  width,
		height: height,
		game: phaser.Get("Game").New(width, height, phaser.Get("AUTO"), "", js.M{
			"preload": preload,
			"create":  create,
			"render":  render,
		}),
	}
	browser.images.m = make(map[string]struct {
		i     *Image
		donut *js.Object
		anim  *js.Object
	})

	<-ready
	return browser
}

func preload() {
	browser.game.Get("canvas").Set("oncontextmenu", func(e *js.Object) { e.Call("preventDefault") })

	scale := browser.game.Get("scale")
	mode := phaser.Get("ScaleManager").Get("SHOW_ALL")
	scale.Set("scaleMode", mode)
	scale.Set("fullScreenScaleMode", mode)
	scale.Set("pageAlignHorizontally", true)
	scale.Set("pageAlignVertically", true)

	browser.load = browser.game.Get("load")
	browser.load.Call("spritesheet", "donut", "loading.png", 25, 25, 8)
}

func create() {
	browser.add = browser.game.Get("add")
	world := browser.game.Get("world")
	browser.centerx = world.Get("centerX").Int()
	browser.centery = world.Get("centerY").Int()

	// loading.Set("visible", true)
	// loading.Set("alpha", 0)
	// fadeIn := newTween(loading, js.M{"alpha": 1}, 1333)
	// fadeIn.Call("start")

	// bg = "res/bg.png"
	// taptostart = "res/taptostart.png"
	// username = "res/username.png"

	// onLoad, _ := jsutil.Callback()

	browser.load.Get("onFileComplete").Call("add", func(_, key *js.Object) {
		browser.addImage(key.String())
	})

	// browser.load.Call("image", bg, bg)
	// browser.load.Call("spritesheet", taptostart, taptostart, w, h)
	// browser.load.Call("spritesheet", username, username, 360, 216)

	ready <- true
	close(ready)
}

func render() {
	go renderLoading()
}

func renderLoading() {
	<-fps
	browser.images.RLock()
	for _, ld := range browser.images.m {
		browser.game.Get("debug").Call("geom", ld.i.js.o, "rgba(224,224,224,0.5)")
	}
	browser.images.RUnlock()
}

func (b *Browser) addImage(key string) {
	o := b.add.Call("sprite", b.centerx, b.centery, key)
	o.Get("anchor").Call("setTo", 0.5, 0.5)

	b.images.Lock()
	defer b.images.Unlock()

	ld := b.images.m[key]
	o.Set("width", ld.i.js.o.Get("width"))
	o.Set("height", ld.i.js.o.Get("height"))
	ld.donut.Set("visible", false)
	ld.anim.Call("stop")

	ld.i.js.Lock()
	ld.i.js.o = o
	ld.i.js.Unlock()
}

// func newTween(o *js.Object, params js.M, ms int) *js.Object {
// 	twn := browser.add.Call("tween", o).Call("to", params, ms)
// 	twn.Set("frameBased", true)
// 	return twn
// }

func (b *Browser) NewImage(url string, width, height int) *Image {
	var i Image

	b.images.Lock()
	defer b.images.Unlock()

	if _, exists := b.images.m[url]; exists {
		o := b.add.Call("sprite", b.centerx, b.centery, url)
		o.Get("anchor").Call("setTo", 0.5, 0.5)
		o.Set("width", width)
		o.Set("height", height)

		return &i
	}

	i.js.o = phaser.Get("Rectangle").New(
		b.width/2-width/2, b.height/2-height/2,
		width, height,
	)

	donut := b.add.Call("sprite", b.centerx, b.centery, "donut")
	donut.Get("anchor").Call("setTo", 0.5, 0.5)
	anim := donut.Get("animations").Call("add", "spin")
	anim.Call("play", 8, true)

	ld := b.images.m[url]
	ld.i = &i
	ld.donut = donut
	ld.anim = anim
	b.images.m[url] = ld

	b.load.Call("image", url, url)
	b.load.Call("start")

	return &i
}

// func newButton(url string) (*js.Object, <-chan bool) {
// 	x, y := world.Get("centerX").Int(), world.Get("centerY").Int()
// 	onHit, hit := jsutil.Callback()
// 	btn := browser.game.Get("add").Call("button", x, y, url, onHit, nil, 0, 0, 1, 0)
// 	btn.Set("visible", false)
// 	btn.Get("anchor").Call("setTo", 0.5, 0.5)
// 	btn.Get("onInputDown").Call("add", func() { btn.Set("y", y+util.Min(h-btn.Get("h").Int(), 3)) })
// 	btn.Get("onInputOver").Call("add", func() { btn.Set("y", y) })
// 	btn.Get("onInputOut").Call("add", func() { btn.Set("y", y) })
// 	btn.Get("onInputUp").Call("add", func() { btn.Set("y", y) })
// 	return btn, hit
// }
