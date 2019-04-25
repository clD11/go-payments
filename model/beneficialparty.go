package model

type BeneficiaryParty struct {
	AccountName       string `json:"account_name"`
	AccountNumber     string `json:"account_number"`
	AccountNumberCode string `json:"account_number_code"`
	AccountType       int    `json:"account_type"`
	Address           string `json:"address"`
	BankID            string `json:"bank_id"`
	BankIDCode        string `json:"bank_id_code"`
	Name              string `json:"name"`
}
