package actions

import "github.com/ruupert/paste/frontend/store/model"

type Action interface{}

type Response interface{}

type SetPaste struct {
	Text string
}
type SetSaved struct {
	Saved bool
}

type SetComplete struct {
	Complete bool
}

type SetLocation struct {
	Location string
}

type ReplaceItems struct {
	Items *model.PasteItem
}

type PostPaste struct {
	Text     string
	Hash     string
	Complete bool
}

type GetPaste struct {
	Hash string
	Text string
}
type SetPosting struct {
	Text string
}
type SetPosted struct {
	Text string
}

type RenderPosted struct {
	Location string
}
