//go:build wasm && js
// +build wasm,js

package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
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
	Input string
}

func (p *PageView) SetPath(path string) {

	p.Render()
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
				fmt.Println("KeyUp")
			}).StopPropagation(),
			event.KeyDown(func(e *vecty.Event) {
				ctrlDown := e.Get("ctrlKey").Bool()
				metaDown := e.Get("metaKey").Bool()
				fmt.Println("KeyDown")
				fmt.Println(e.Value.Get("key"))
				fmt.Println(p.Input)
				switch e.Get("keyCode").Int() {
				case 83: // S
					if ctrlDown || metaDown {
						e.Call("preventDefault")
						if p.Input != "" {
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
								p.SetPath(resp.Header.Get("Location"))
							}()
						}
					}
				}
			}).StopPropagation(),
		),
	)
}
