package service

import "fmt"

type KeyError struct{
	key string
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("Key Error -> Key: %s", e.key)
}

type PacketLossError struct{
	Msg string
}

func (e *PacketLossError) Error() string{
	return fmt.Sprintf("Packet Loss Error -> msg: %s", e.Msg)
}