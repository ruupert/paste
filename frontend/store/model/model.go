package model

type PasteItem struct {
	Text      string
	Saved     bool
	Encrypted bool
	Completed bool
	Location  string
}

type PostPaste struct {
	Completed bool
	Err       error
	Location  string
}
