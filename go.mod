module github.com/ruupert/paste

go 1.24.4

replace github.com/ruupert/paste => ./

require (
	github.com/hexops/vecty v0.6.0
	go.etcd.io/bbolt v1.4.1
)

require golang.org/x/sys v0.33.0 // indirect
