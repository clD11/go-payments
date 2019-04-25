# **Form3-Payments App**

* Requires Docker and Docker Compose

### Run Tests
 
    go build
    go test

Integration tests (_main_test.go_) uses TestContainers to spin up 
containerised postgres database to run against.

### Run Application
    
    docker-compose up

Application runs _localhost:8080_