//go:build wasm && js
// +build wasm,js

package components

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall/js"

	"github.com/hexops/vecty"
	"github.com/hexops/vecty/elem"
	"github.com/hexops/vecty/event"
	"github.com/hexops/vecty/prop"
	"github.com/ruupert/paste/frontend/actions"
	"github.com/ruupert/paste/frontend/dispatcher"
	"github.com/ruupert/paste/frontend/store/model"
)

type PageView struct {
	vecty.Core
	PasteItem *model.PasteItem `vecty:"prop"`
	pasteText string
	PostPaste *model.PostPaste `vecty:"prop"`
}

func (p *PageView) onChangeSave(event *vecty.Event) {
	dispatcher.Dispatch(&actions.SetPaste{Text: event.Target.Get("value").String()})
	vecty.Rerender(p)

}
func (p *PageView) onNewInput(event *vecty.Event) {
	p.pasteText = event.Target.Get("value").String()
	p.onChangeSave(event)
	vecty.Rerender(p)
}

func (p *PageView) handleKeyDown(e *vecty.Event) {
	fmt.Println(e.Value.Get("key"))
	ctrlDown := e.Value.Get("ctrlKey").Bool()
	metaDown := e.Value.Get("metaKey").Bool()
	switch e.Value.Get("keyCode").Int() {
	case 83: // S
		if ctrlDown || metaDown {
			e.Value.Call("preventDefault")
			fmt.Println("PREVENT")
			go func() {
				resp, err := http.Post("//", "text/html", bytes.NewReader([]byte(p.pasteText)))
				if err != nil {
					fmt.Println(err)
				}
				val := resp.Header.Get("location")
				resp.Body.Close()
				dispatcher.Dispatch(&actions.SetLocation{Location: val})
				js.Global().Get("window").Set("location", val)
			}()
			dispatcher.Dispatch(&actions.SetComplete{Complete: true})
			dispatcher.Dispatch(&actions.SetSaved{Saved: true})
		}
	}
}

func (p *PageView) render404() *vecty.HTML {
	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Property("id", "float404"),
			),
			elem.Image(
				vecty.Markup(
					vecty.Attribute("src", "/404"),
				),
			),
		),
	)
}

func (p *PageView) renderPaste() *vecty.HTML {
	return elem.Section(
		elem.TextArea(
			vecty.Markup(
				vecty.Markup(vecty.Property("placeholder", "[ paste text  -  Meta+s to save ]")),
				vecty.Property("spellcheck", false),
				vecty.Property("id", "paste"),
				vecty.Class("textarea"),
				vecty.Property("autofocus", true),
				event.KeyDown(p.handleKeyDown),
				event.Input(p.onNewInput),
				prop.Value(p.pasteText),
			),
		),
	)
}

func (p *PageView) renderView() *vecty.HTML {
	return elem.Section(
		elem.TextArea(
			vecty.Markup(
				vecty.Markup(vecty.Property("placeholder", "[ paste text  -  Meta+s to save ]")),
				vecty.Property("spellcheck", false),
				vecty.Property("id", "paste"),
				vecty.Class("textarea"),
				vecty.Attribute("readonly", true),
				prop.Value(p.pasteText),
			),
		),
	)
}

func (p *PageView) Render() vecty.ComponentOrHTML {
	v := js.Global().Get("window").Get("location").Get("pathname").String()
	if v == "/" {
		return elem.Body(
			elem.Section(
				p.renderPaste(),
			),
		)
	} else {
		results := make(chan []byte)
		go func() {
			s, f := strings.CutPrefix(v, "/")
			if !f {
				fmt.Println(f)
			}
			resp, err := http.Get("/?q=" + s)
			if err != nil {
				fmt.Println("err")
			}
			defer resp.Body.Close()
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			if resp.StatusCode == http.StatusOK {
				results <- bodyBytes
			} else {
				results <- []byte("")
			}
		}()
		msg := <-results
		p.pasteText = string(msg)
		if p.pasteText == "" {
			return elem.Body(
				p.render404())
		} else {
			return elem.Body(
				elem.Section(
					p.renderView(),
				),
			)
		}
	}
}
