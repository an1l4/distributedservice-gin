# distributedservice-gin

# steps
go run main.go

curl -X GET http://localhost:5000/anila

go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger version

swagger generate spec -o ./swagger.json

swagger serve ./swagger.json

swagger serve -F swagger ./swagger.json

docker run -d --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password -p 27017:27017 mongo:4.4.3

"mongodb://admin:password@localhost:27017"

export MONGO_URI=mongodb://admin:password@localhost:27017 && export MONGO_DATABASE=demo && go run main.go

mongoimport --username admin --password password --authenticationDatabase admin --db demo --collection recipes --file recipes.json --jsonArray