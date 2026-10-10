package ledger

func (a *Account) Available() int64 {
	return a.Balance - a.Reserved
}

func (a *Account) Reserve(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if amount > a.Available() {
		return ErrInsufficientFunds
	}

	a.Reserved += amount
	return nil
}
