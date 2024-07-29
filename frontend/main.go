//go:build wasm && js
// +build wasm,js

package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/hexops/vecty"
	"github.com/hexops/vecty/elem"
	"github.com/hexops/vecty/event"
)

const ConnectMaxWaitTime = 1 * time.Second

var client http.Client

func main() {
	vecty.SetTitle("wpastebin")
	vecty.RenderBody(&PageView{})
	client = http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: ConnectMaxWaitTime,
			}).DialContext,
		},
	}
}

type PageView struct {
	vecty.Core
	Input      string
	MetaToggle bool
	sync.RWMutex
}

func (p *PageView) Render() vecty.ComponentOrHTML {
	return elem.Body(
		elem.TextArea(
			vecty.Markup(
				vecty.Property("placeholder", "[ paste text  -  Meta+s to save ]"),
				vecty.Style("font-family", "monospace"),
				vecty.Property("id", "paste"),
				event.Input(func(e *vecty.Event) {
					p.Input = e.Target.Get("value").String()
				}).StopPropagation(),
			),
			vecty.Text(p.Input),
		),
		vecty.Markup(
			event.Paste(func(e *vecty.Event) {
				fmt.Println("PASTE")
				p.Input = e.Value.String()
			}),
			event.KeyUp(func(e *vecty.Event) {
				p.RWMutex.Lock()
				defer p.RWMutex.Unlock()
				fmt.Println("KeyUp")
				p.MetaToggle = false
			}).StopPropagation(),
			event.KeyDown(func(e *vecty.Event) {
				p.RWMutex.Lock()
				defer p.RWMutex.Unlock()
				fmt.Println("KeyDown")
				k := e.Value.Get("key").String()
				fmt.Println(e.Value.Get("key"))
				if p.MetaToggle {
					switch k {
					case "s":
						fmt.Println("Input: " + p.Input)
						fmt.Println("Meta+" + e.Value.Get("key").String())
						go func() {
							ctx := context.Background()
							req, err := http.NewRequestWithContext(ctx, "POST", "//", bytes.NewReader([]byte(p.Input)))
							if err != nil {
								fmt.Println("err")
								return
							}
							req = req.WithContext(ctx)
							resp, err := http.DefaultClient.Do(req)
							if err != nil {
								fmt.Println("Error do req")
								return
							}
							fmt.Println(resp.Header.Get("Location"))
						}()
					case "w":
						fmt.Println("Meta+" + e.Value.Get("key").String())
					case "v":
						// slurp clipboard here I guess
						fmt.Println("Meta+" + e.Value.Get("key").String())
					case "r":
						fmt.Println("Meta+" + e.Value.Get("key").String())
						fmt.Println("Reloaded")
					}
				}
				if k == "Meta" {
					fmt.Println("Meta toggled")
					p.MetaToggle = true
				}
			}).StopPropagation(),
		),
	)
}
