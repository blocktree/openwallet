package keypair

type EncryptError struct {
	detail string
}

func (e *EncryptError) Error() string {
	return "encrypt private key error: " + e.detail
}

func NewEncryptError(msg string) *EncryptError {
	return &EncryptError{detail: msg}
}

type DecryptError EncryptError

func (e *DecryptError) Error() string {
	return "decrypt private key error: " + e.detail
}

func NewDecryptError(msg string) *DecryptError {
	return &DecryptError{detail: msg}
}
