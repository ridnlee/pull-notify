grpc:
	protoc -I=./ --go-grpc_out=./../ --go_out=./../  ./api/messages.proto
build:
	go build .
docker:
	docker build -t test-notify . 