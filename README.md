# **Form3 Payment API**
This is a simple web app offering an API to fetch, create, update and delete payments

* Clone the repository
* The project uses versioned modules, if cloned into `GOPATH` run `export GO111MODULE=on` to enable this feature
* Docker Compose is need to run the application

### Running the Application
    
    docker-compose up

Application runs on _localhost:8080_

### Running the Tests
Integration tests are located in `main_test.go` and create a postgres database running in a docker container.
Tests can be run using below command or through IDE. The container is created before the test suite runs and destroyed after.

    go test
