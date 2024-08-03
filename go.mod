module github.com/ruupert/paste

go 1.22.5

replace github.com/ruupert/paste => ./

require (
	github.com/hexops/vecty v0.6.0
	go.etcd.io/bbolt v1.3.10
)

require golang.org/x/sys v0.22.0 // indirect
