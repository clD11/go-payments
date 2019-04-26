# **Form3 Payment API**

* Clone the repository
* Modules are used so if inside GOPATH run export GO111MODULE=on to allow modules
* Docker Compose is need to run the application

### Run Tests
Tests can be run using below commands or through IDE
 
    go build
    go test

Integration tests (_main_test.go_) uses TestContainers to spin up a
containerised postgres database.

### Run Application
    
    docker-compose up

Application runs _localhost:8080_
