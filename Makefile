wasm-serve:
	go run github.com/hajimehoshi/wasmserve@latest .

target-wasm: html/index.html
	mkdir -p out
	env GOOS=js GOARCH=wasm go build -o out/main.wasm .
	cp html/index.html out/
	cp /usr/lib/go/misc/wasm/wasm_exec.js out/
	cp -r assets out/

clean:
	rm -rf out
	rm -rf go.sum
