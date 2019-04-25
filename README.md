# **Form3 Payment API**

* Clone the repository and if inside GOPATH might require export GO111MODULE=on to allow modules
* Requires Docker and Docker Compose

### Run Tests
Tests can be run using below commands or through IDE
 
    go build
    go test

Integration tests (_main_test.go_) uses TestContainers to spin up 
containerised postgres database to run against.

### Run Application
    
    docker-compose up

Application runs _localhost:8080_
