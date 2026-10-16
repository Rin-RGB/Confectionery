package domain

type Payment struct {
	ID       string
	OrderID  string
	Amount   int64
	Currency string
	Status   string
}
