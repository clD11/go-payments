package model

type ChargesInformation struct {
	BearerCode              string   `json:"bearer_code"`
	SenderCharges           []Charge `json:"sender_charges"`
	ReceiverChargesAmount   string   `json:"receiver_charges_amount"`
	ReceiverChargesCurrency string   `json:"receiver_charges_currency"`
}
