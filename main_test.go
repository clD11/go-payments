package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/clD11/form3-payments/app"
	. "github.com/clD11/form3-payments/model"
	"github.com/go-pg/pg"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var sut app.App
var server *http.Server

func TestMain(m *testing.M) {
	// Setup database for testing
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432:5432"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
	}

	postgresContainer, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer postgresContainer.Terminate(ctx)

	host, _ := postgresContainer.Host(ctx)

	conn, err := sql.Open("postgres", fmt.Sprintf("host=%s port=5432 user=postgres password=postgres dbname=postgres sslmode=disable", host))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// ping the database until it comes up
	timeout := time.Now().Add(time.Second * 20)
	for time.Now().Before(timeout) {
		if err = conn.Ping(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Setup and start app for testing
	config := app.Config{
		DB: &pg.Options{
			Addr:     fmt.Sprintf("%s:5432", host),
			Database: "postgres",
			User:     "postgres",
			Password: "postgres",
		},
	}

	sut = app.App{}
	sut.Initialize(&config)
	// Use server for testing instead of sut.RUN(port) which blocks (could use goroutine in app)
	server = &http.Server{Addr: ":9807", Handler: sut.Router}

	code := m.Run()
	os.Exit(code)
}

func TestGetPaymentShouldReturnStatusBadRequestWhenIDInvalid(t *testing.T) {
	truncateTables(t)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", "randomText"), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "Invalid ID", getErrorMsg(rw))
}

func TestGetPaymentShouldReturnStatusNotFoundWhenPaymentDoesNotExists(t *testing.T) {
	truncateTables(t)

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", uuid.NewV1()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, "Payment not found", getErrorMsg(rw))
}

func TestGetPaymentShouldReturnSinglePaymentWhenPaymentExistsIntegration(t *testing.T) {
	truncateTables(t)

	expectedPayment := createPayment()
	if err := sut.DB.Insert(&expectedPayment); err != nil {
		t.Fatalf("Could not insert seed data payments - %s", err.Error())
	}

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", expectedPayment.ID), nil)

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	var actualPayment Payment
	json.NewDecoder(rw.Body).Decode(&actualPayment)

	expectedStatusCode := http.StatusOK
	actualStatusCode := rw.Code

	assert.Equal(t, expectedStatusCode, actualStatusCode)
	assert.Equal(t, expectedPayment, actualPayment)
}

func TestCreatePaymentShouldReturnStatusBadRequestWhenInvalidPayload(t *testing.T) {
	truncateTables(t)
	invalidPayload, _ := json.Marshal("{ random: 'random' }")

	request := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBuffer(invalidPayload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "Could not decode request body", getErrorMsg(rw))
}

func TestCreatePaymentShouldReturnStatusBadRequestWhenPaymentAlreadyExists(t *testing.T) {
	truncateTables(t)

	expectedPayment := createPayment()
	if err := sut.DB.Insert(&expectedPayment); err != nil {
		t.Fatalf("Could not insert seed data payments - %s", err.Error())
	}

	payload, _ := json.Marshal(expectedPayment)

	request := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBuffer(payload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code, "Expected Status Bad Request")
	assert.Equal(t, "Cannot create payment already exists", getErrorMsg(rw))
}

func TestCreatePaymentShouldReturnStatusCreated(t *testing.T) {
	truncateTables(t)

	expectedPayment := createPayment()
	payload, _ := json.Marshal(expectedPayment)

	request := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBuffer(payload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	expectedStatusCode := http.StatusCreated
	actualStatusCode := rw.Code

	actualPayment := Payment{ID: expectedPayment.ID}
	if err := sut.DB.Select(&actualPayment); err != nil {
		t.Fatalf("Payment was not created by request")
	}

	assert.Equal(t, expectedStatusCode, actualStatusCode)
	assert.Equal(t, expectedPayment, actualPayment)
}

func TestDeletePaymentShouldReturnStatusBadRequestWhenInvalidID(t *testing.T) {
	truncateTables(t)

	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/payments/%s", "randomText"), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "Invalid ID", getErrorMsg(rw))
}

func TestDeletePaymentShouldReturnStatusNotFoundWhenPaymentDoesNotExist(t *testing.T) {
	truncateTables(t)

	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/payments/%s", uuid.NewV1()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, getErrorMsg(rw), "Payment not found cannot delete")
}

func TestDeletePaymentShouldReturnStatusOKWhenPaymentDeleted(t *testing.T) {
	truncateTables(t)

	expectedPayment := createPayment()
	if err := sut.DB.Insert(&expectedPayment); err != nil {
		t.Log(err)
	}

	request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/payments/%s", expectedPayment.ID), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusOK, rw.Code)
	assertPaymentDoseNotExist(t, expectedPayment.ID)
}

func TestUpdatePaymentShouldReturnStatusBadRequestWhenInvalidID(t *testing.T) {
	truncateTables(t)

	request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", "invalidUUID"), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "Invalid ID", getErrorMsg(rw))
}

func TestUpdatePaymentShouldReturnStatusBadRequestWhenPayloadInvalid(t *testing.T) {
	truncateTables(t)

	invalidPayload, _ := json.Marshal("{ random: 'random' }")

	request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", uuid.NewV1()), bytes.NewBuffer(invalidPayload))
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "Could not decode request body", getErrorMsg(rw))
}

func TestUpdatePaymentShouldReturnStatusBadRequestWhenPayloadDoesNotMatchID(t *testing.T) {
	truncateTables(t)

	payment := createPayment()
	payload, _ := json.Marshal(payment)

	request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", uuid.NewV1()), bytes.NewBuffer(payload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, getErrorMsg(rw), "Could not update payment - request id does not match update payment")
	assertPaymentDoseNotExist(t, payment.ID)
}

func TestUpdatePaymentShouldReturnStatusNotFoundWhenPaymentNotInDatabase(t *testing.T) {
	truncateTables(t)

	payment := createPayment()
	payload, _ := json.Marshal(payment)

	request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", payment.ID), bytes.NewBuffer(payload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	assert.Equal(t, getErrorMsg(rw), "Could not update payment as not found")
	assertPaymentDoseNotExist(t, payment.ID)
}

func TestUpdatePaymentShouldReturnStatusCreated(t *testing.T) {
	truncateTables(t)

	expectedPayment := createPayment()
	if err := sut.DB.Insert(&expectedPayment); err != nil {
		t.Fatalf("Could not insert payment")
	}

	expectedPayment.Version = 1

	payload, _ := json.Marshal(expectedPayment)
	request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", expectedPayment.ID), bytes.NewBuffer(payload))

	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	actualPayment := Payment{ID: expectedPayment.ID}
	if err := sut.DB.Select(&actualPayment); err != nil {
		t.Fatalf("Could not find payment")
	}

	assert.Equal(t, http.StatusCreated, rw.Code)
	assert.Equal(t, expectedPayment, actualPayment)
}

func TestGetPaymentShouldReturnAllPayments(t *testing.T) {
	truncateTables(t)

	expectedPayments := createPayments()
	for _, payment := range expectedPayments {
		if err := sut.DB.Insert(&payment); err != nil {
			t.Fatalf("Could not insert seed data payments - %s", err.Error())
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, request)

	var actualPayments []Payment
	json.NewDecoder(rw.Body).Decode(&actualPayments)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, expectedPayments, actualPayments)
}

func getErrorMsg(rw *httptest.ResponseRecorder) string {
	var msg map[string]string
	json.NewDecoder(rw.Body).Decode(&msg)
	return msg["error"]
}

func truncateTables(t *testing.T) {
	for _, table := range getTables() {
		if _, err := sut.DB.Model(table).Where("1=1").Delete(); err != nil {
			t.Fatalf("Error truncating tables")
		}
	}
}

func getTables() []interface{} {
	return []interface{}{
		(*Payment)(nil),
		(*Attributes)(nil),
		(*BeneficiaryParty)(nil),
		(*ChargesInformation)(nil),
		(*SponsorParty)(nil),
		(*DebtorParty)(nil),
		(*Charge)(nil),
		(*Fx)(nil)}
}

func createPayment() Payment {
	return Payment{
		Type:           "Payment",
		ID:             uuid.NewV1(),
		Version:        0,
		OrganisationID: uuid.NewV1(),
		Attributes: Attributes{
			Amount: "100.21",
			BeneficiaryParty: BeneficiaryParty{
				AccountName:       "W Owens",
				AccountNumber:     "31926819",
				AccountNumberCode: "BBAN",
				AccountType:       0,
				Address:           "1 The Beneficiary Localtown SE2",
				BankID:            "403000",
				BankIDCode:        "GBDSC",
				Name:              "Wilfred Jeremiah Owens",
			},
			ChargesInformation: ChargesInformation{
				BearerCode: "SHAR",
				SenderCharges: []Charge{
					{
						Amount:   "5.00",
						Currency: "GBP",
					},
					{
						Amount:   "10.00",
						Currency: "USD",
					},
				},
				ReceiverChargesAmount:   "1.00",
				ReceiverChargesCurrency: "USD",
			},
			Currency: "GBP",
			DebtorParty: DebtorParty{
				AccountName:       "EJ Brown Black",
				AccountNumber:     "GB29XABC10161234567801",
				AccountNumberCode: "IBAN",
				Address:           "10 Debtor Crescent Sourcetown NE1",
				BankID:            "203301",
				BankIDCode:        "GBDSC",
				Name:              "Emelia Jane Brown",
			},
			EndToEndReference: "Wil piano Jan",
			Fx: Fx{
				ContractReference: "FX123",
				ExchangeRate:      "2.00000",
				OriginalAmount:    "200.42",
				OriginalCurrency:  "USD",
			},
			NumericReference:     "1002001",
			PaymentID:            "123456789012345678",
			PaymentPurpose:       "Paying for goods/services",
			PaymentScheme:        "FPS",
			PaymentType:          "Credit",
			ProcessingDate:       "2017-01-18",
			Reference:            "Payment for Em's piano lessons",
			SchemePaymentSubType: "InternetBanking",
			SchemePaymentType:    "ImmediatePayment",
			SponsorParty: SponsorParty{
				AccountNumber: "56781234",
				BankID:        "123123",
				BankIDCode:    "GBDSC",
			},
		},
	}
}

func assertPaymentDoseNotExist(t *testing.T, uuid uuid.UUID) {
	payment := Payment{ID: uuid}
	if err := sut.DB.Select(&payment); err != pg.ErrNoRows {
		t.Fatalf("Payment should not exist in database")
	}
}

func createPayments() (payments []Payment) {
	data, _ := ioutil.ReadFile("seeddata.json")
	json.Unmarshal(data, &payments)
	return
}
