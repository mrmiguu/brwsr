package brwsr

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/brwsr/util"
	"github.com/mrmiguu/jsutil"
)

var (
	alert             func(interface{})
	document          *js.Object
	documentBody      *js.Object
	documentBodyStyle *js.Object
	window            *js.Object

	phaser    *js.Object
	game      *js.Object
	gameScale *js.Object
	gameLoad  *js.Object
	gameAdd   *js.Object
	gameWorld *js.Object
	sck       *js.Object

	width      int
	height     int
	address    string
	bg         string
	taptostart string
	username   string
	loading    *js.Object
	spin       *js.Object
	midtext    *js.Object
	goBtn      *js.Object
	goHit      <-chan bool
	// ws                *js.Object
	// t                 *js.Object
	// button            *js.Object
)

func init() {
	alert = func(x interface{}) { js.Global.Call("alert", x) }
	document = js.Global.Get("document")
	documentBody = document.Get("body")
	documentBodyStyle = documentBody.Get("style")
	window = js.Global.Get("window")

	documentBodyStyle.Set("background", "#000000")
	documentBodyStyle.Set("margin", 0)

	// load libraries
	<-jsutil.Lib("https://github.com/photonstorm/phaser-ce/releases/download/v2.8.3/phaser.min.js")

	phaser = js.Global.Get("Phaser")
}

// New creates a new browser instance.
func New(w, h int, addr string) {
	width, height = w, h
	address = addr
	game = phaser.Get("Game").New(w, h, phaser.Get("AUTO"), "phaser-example", js.M{"preload": preload, "create": create})
}

func preload() {
	game.Get("canvas").Set("oncontextmenu", func(e *js.Object) { e.Call("preventDefault") })
	gameScale = game.Get("scale")
	showAll := phaser.Get("ScaleManager").Get("SHOW_ALL")
	gameScale.Set("scaleMode", showAll)
	gameScale.Set("fullScreenScaleMode", showAll)
	gameScale.Set("pageAlignHorizontally", true)
	gameScale.Set("pageAlignVertically", true)

	gameLoad = game.Get("load")
	gameLoad.Call("spritesheet", "loading", "res/loading.png", width, height)
}

func create() {
	gameAdd = game.Get("add")
	gameWorld = game.Get("world")

	loading = newSprite("loading")
	loading.Set("visible", true)
	loading.Set("alpha", 0)
	fadeIn := newTween(loading, js.M{"alpha": 1}, 1333)
	fadeIn.Call("start")

	bg = "res/bg.png"
	taptostart = "res/taptostart.png"
	username = "res/username.png"

	onLoad, loaded := jsutil.Callback()
	gameLoad.Get("onLoadComplete").Call("add", onLoad)

	spin = loading.Get("animations").Call("add", "spin")
	spin.Call("play", 9, true)
	onSpin, spun := jsutil.Callback()
	spin.Get("onLoop").Call("add", onSpin)
	spunAndLoaded := onSpinLoop(spun, loaded)

	gameLoad.Call("image", bg, bg)
	gameLoad.Call("spritesheet", taptostart, taptostart, width, height)
	gameLoad.Call("spritesheet", username, username, 360, 216)

	gameLoad.Call("start")

	onSpinAndLoad(spunAndLoaded)

	sck = js.Global.Get("WebSocket").New("ws://" + address + "/connected")
	sck.Set("onopen", onConnectionOpen)
	sck.Set("onclose", onConnectionClose)
	sck.Set("onmessage", onConnectionMessage)
	sck.Set("onerror", onConnectionError)
}

func onSpinLoop(spun, loaded <-chan bool) <-chan bool {
	spunAndLoaded := make(chan bool)
	go func() {
		for {
			<-spun
			select {
			case <-loaded:
				spunAndLoaded <- true
				return
			default:
			}
		}
	}()
	return spunAndLoaded
}

func onSpinAndLoad(spunAndLoaded <-chan bool) {
	go func() {
		<-spunAndLoaded
		spin.Call("stop")
		loading.Set("visible", false)

		bgSpr := newSprite(bg)
		midtext = gameAdd.Call("text", 0, 0, "_", js.M{
			"font":  "65px Arial",
			"fill":  "#ffffff",
			"align": "center",
		})

		goBtn, goHit = newButton(taptostart)
		goBtn.Set("visible", true)
		bgSpr.Set("visible", true)
		onGo()
	}()
}

func onGo() {
	go func() {
		<-goHit
		gameScale.Call("startFullScreen")
		usrBtn, usrHit := newButton(username)
		usrBtn.Set("visible", true)
		goBtn.Set("visible", false)

		txt := jsutil.OpenKeyboard()
		go func() {
			for {
				midtext.Call("setText", <-txt)
			}
		}()

		<-usrHit
		sck.Call("send", midtext.Get("text"))
		jsutil.CloseKeyboard()
		usrBtn.Set("visible", false)
		midtext.Call("setText", "")
	}()
}

func newTween(o *js.Object, params js.M, ms int) *js.Object {
	twn := gameAdd.Call("tween", o).Call("to", params, ms)
	twn.Set("frameBased", true)
	return twn
}

func newSprite(id string) *js.Object {
	spr := gameAdd.Call("sprite", gameWorld.Get("centerX"), gameWorld.Get("centerY"), id)
	spr.Set("visible", false)
	spr.Get("anchor").Call("setTo", 0.5, 0.5)
	return spr
}

func newButton(url string) (*js.Object, <-chan bool) {
	x, y := gameWorld.Get("centerX").Int(), gameWorld.Get("centerY").Int()
	onHit, hit := jsutil.Callback()
	btn := game.Get("add").Call("button", x, y, url, onHit, nil, 0, 0, 1, 0)
	btn.Set("visible", false)
	btn.Get("anchor").Call("setTo", 0.5, 0.5)
	btn.Get("onInputDown").Call("add", func() { btn.Set("y", y+util.Min(height-btn.Get("height").Int(), 3)) })
	btn.Get("onInputOver").Call("add", func() { btn.Set("y", y) })
	btn.Get("onInputOut").Call("add", func() { btn.Set("y", y) })
	btn.Get("onInputUp").Call("add", func() { btn.Set("y", y) })
	return btn, hit
}

func onConnectionOpen(evt *js.Object) {
	alert("Connected")
}

func onConnectionClose(evt *js.Object) {
	alert("Disconnected")
	sck = nil
}

func onConnectionMessage(evt *js.Object) {
	alert("Server: " + evt.Get("data").String())
}

func onConnectionError(evt *js.Object) {
	alert("Error: " + evt.Get("data").String())
}

// func print(message interface{}) {
// 	t.Set("text", message)
// }
