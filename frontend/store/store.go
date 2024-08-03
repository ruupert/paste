package store

import (
	"github.com/ruupert/paste/frontend/actions"
	"github.com/ruupert/paste/frontend/dispatcher"
	"github.com/ruupert/paste/frontend/store/model"
	"github.com/ruupert/paste/frontend/store/storeutil"
)

var (
	// Items represents all of the TODO items in the store.
	PasteItem *model.PasteItem
	PostPaste *model.PostPaste
	// Filter represents the active viewing filter for items.
	// Listeners is the listeners that will be invoked when the store changes.
	Listeners = storeutil.NewListenerRegistry()
)

func init() {
	dispatcher.Register(onAction)
}

func onAction(action interface{}) {
	switch a := action.(type) {
	case *actions.SetPaste:
		PasteItem = &model.PasteItem{Text: a.Text}
	case *actions.PostPaste:
		PostPaste = &model.PostPaste{Completed: true, Location: a.Hash}
	case *actions.RenderPosted:
	default:
		return
	}
	Listeners.Fire()
}
