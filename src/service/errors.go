package service

import "fmt"

type KeyError struct{
	key string
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("Key Error -> Key: %s", e.key)
}
