//go:build wasm && js
// +build wasm,js

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/hexops/vecty"
	"github.com/ruupert/paste/frontend/actions"
	"github.com/ruupert/paste/frontend/components"
	"github.com/ruupert/paste/frontend/dispatcher"
	"github.com/ruupert/paste/frontend/store"
)

func main() {
	attachLocalStorage()
	vecty.SetTitle("Paste")
	vecty.AddStylesheet("/css")
	p := &components.PageView{}
	store.Listeners.Add(p, func() {
		p.PasteItem = store.PasteItem
		p.PostPaste = store.PostPaste
		vecty.Rerender(p)
	})
	vecty.RenderBody(p)
}

func attachLocalStorage() {
	store.Listeners.Add(nil, func() {
		data, err := json.Marshal(store.PasteItem)
		if err != nil {
			println("failed to store items: " + err.Error())
		}
		js.Global().Get("localStorage").Set("items", string(data))
	})
	if data := js.Global().Get("localStorage").Get("items"); !data.IsUndefined() {
		var items string
		if err := json.Unmarshal([]byte(data.String()), &items); err != nil {
			println("failed to load items: " + err.Error())
		}
		dispatcher.Dispatch(&actions.SetPaste{
			Text: items,
		})
	}
}
