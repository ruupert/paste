module github.com/ruupert/paste

go 1.25.3

replace github.com/ruupert/paste => ./

require (
	github.com/hexops/vecty v0.6.0
	go.etcd.io/bbolt v1.4.3
)

require golang.org/x/sys v0.37.0 // indirect
