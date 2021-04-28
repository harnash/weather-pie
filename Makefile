build:
	GOOS=linux GOARCH=arm GOARM=5 go build -o ./bin/weather-pie main.go
	scp ./bin/weather-pie wisienka:~/