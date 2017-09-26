package brwsr

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/jsutil"
)

const (
	ShowDelayMilli = 250
)

var (
	phaser  *js.Object
	browser *Browser
	ready   = make(chan bool, 1)
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
			i    *Image
			anim *js.Object
		}
	}
}

type Image struct {
	js struct {
		sync.RWMutex
		o     *js.Object
		donut *js.Object
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
		}),
	}
	browser.images.m = make(map[string]struct {
		i    *Image
		anim *js.Object
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
	browser.load.Call("image", "placeholder", "placeholder.png")
	browser.load.Call("spritesheet", "donut", "loading.png", 32, 32, 8)
}

func create() {
	browser.add = browser.game.Get("add")
	world := browser.game.Get("world")
	browser.centerx = world.Get("centerX").Int()
	browser.centery = world.Get("centerY").Int()

	browser.load.Get("onFileComplete").Call("add", func(_, key *js.Object) {
		go func() {
			// time.Sleep(3 * time.Second)
			browser.addImage(key.String())
		}()
	})

	ready <- true
	close(ready)
}

func (b *Browser) addImage(key string) {
	b.images.Lock()
	ld := b.images.m[key]
	defer b.images.Unlock()
	ld.i.js.Lock()
	defer ld.i.js.Unlock()

	o := b.add.Call("sprite", ld.i.js.o.Get("x"), ld.i.js.o.Get("y"), key)
	o.Get("anchor").Call("setTo", 0.5, 0.5)
	o.Set("width", ld.i.js.o.Get("width"))
	o.Set("height", ld.i.js.o.Get("height"))
	t := fade(o, false, ShowDelayMilli)

	placeholder := ld.i.js.o

	t.Get("onComplete").Call("add", func() {
		ld.i.js.donut.Set("visible", false)
		ld.anim.Call("stop")
		placeholder.Set("visible", false)
	})

	ld.i.js.o = o
}

func tween(o *js.Object, params js.M, ms int) *js.Object {
	t := browser.add.Call("tween", o).Call("to", params, ms)
	t.Set("frameBased", true)
	defer t.Call("start")
	return t
}

func fade(o *js.Object, out bool, ms int) *js.Object {
	src, dst := 0, 1
	if out {
		src, dst = 1, 0
	}
	o.Set("alpha", src)
	return tween(o, js.M{"alpha": dst}, ms)
}

func (b *Browser) NewImage(url string, width, height int) *Image {
	var i Image

	b.images.Lock()
	defer b.images.Unlock()

	if _, exists := b.images.m[url]; exists {
		o := b.add.Call("sprite", b.centerx, b.centery, url)
		fade(o, false, ShowDelayMilli)
		o.Get("anchor").Call("setTo", 0.5, 0.5)
		o.Set("width", width)
		o.Set("height", height)
		i.js.o = o
		return &i
	}

	i.js.o = b.add.Call("sprite", b.centerx, b.centery, "placeholder")
	fade(i.js.o, false, ShowDelayMilli)
	i.js.o.Get("anchor").Call("setTo", 0.5, 0.5)
	i.js.o.Set("width", width)
	i.js.o.Set("height", height)

	donut := b.add.Call("sprite", b.centerx, b.centery, "donut")
	fade(donut, false, ShowDelayMilli)
	donut.Get("anchor").Call("setTo", 0.5, 0.5)
	anim := donut.Get("animations").Call("add", "spin")
	anim.Call("play", 12, true)

	ld := b.images.m[url]
	ld.i = &i
	ld.i.js.donut = donut
	ld.anim = anim
	b.images.m[url] = ld

	b.load.Call("image", url, url)
	b.load.Call("start")

	return &i
}

func (i *Image) Shift(x, y int) {
	i.js.RLock()
	defer i.js.RUnlock()
	i.js.o.Set("x", browser.centerx+x)
	i.js.o.Set("y", browser.centery+y)
	i.js.donut.Call("alignIn", i.js.o, phaser.Get("CENTER"))
}

func (i *Image) Hide(b bool) {
	i.js.RLock()
	defer i.js.RUnlock()
	i.js.o.Set("visible", !b)
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
