# distributedservice-gin

# steps
go run main.go

curl -X GET http://localhost:5000/anila

go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger version

swagger generate spec -o ./swagger.json

swagger serve ./swagger.json

swagger serve -F swagger ./swagger.json
