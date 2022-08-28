build:
	@echo "Building app"
	@GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -ldflags="-s" -o ./bin/weather-pie main.go
pack:
	@echo "Compressing"
	@upx -9 -k ./bin/weather-pie
send:
	@echo "Sending to a remote server"
	@scp ./bin/weather-pie wisienka.harnash.com:~/
deploy: build pack send
	@echo "Deploying on a remote server"
	@ssh wisienka.harnash.com "sudo cp ~/weather-pie /usr/local/bin/weather-pie"
