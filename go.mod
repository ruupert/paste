module github.com/ruupert/paste

go 1.25.12

replace github.com/ruupert/paste => ./

require (
	github.com/hexops/vecty v0.6.0
	go.etcd.io/bbolt v1.5.0
)

require golang.org/x/sys v0.47.0 // indirect
