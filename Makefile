audio_clean:
	@rm -rf ./models/audio/s2t/archives
	@rm -rf ./models/audio/s2t/symphonia_core::probe
	@rm -rf ./models/audio/s2t/multipart_2021::server

whisper:
	@cd ./models/audio/s2t/ && wasmedge --dir .:. whisper-api-server.wasm -m ggml-small.bin

run:
	@go run main.go

build:
	@go build -o term-ollama main.go
