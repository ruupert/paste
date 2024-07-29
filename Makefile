build:
	cd frontend && GOOS=js GOARCH=wasm go build -o ../backend/public/out.wasm . && cd ../backend && go build -o ./paste ./
run:
	./backend/paste
