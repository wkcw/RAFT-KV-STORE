package service

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

type State struct {
	CurrentTerm int64
	VoteFor     string
	logs        *Log
}

func InitState() *State  {
	var state State
	file, err1 := os.Open("../util/STATE_CONFIG")

	if err1 != nil {
		state = State{CurrentTerm: 0, VoteFor: "", logs:  NewLog()}
	} else {
		defer file.Close()

		contents, err2 := ioutil.ReadAll(file)

		if err2 != nil {
			log.Fatalf("Fail to read the STATE_CONFIG: %v", err2)
		}

		decoder := gob.NewDecoder(bytes.NewReader(contents))

		err3 := decoder.Decode(&state)

		if err3 != nil {
			log.Fatalf("state object could not be deserialized: %v", err2)
		}

	}

	return &state
}

func (state *State)PersistentStore() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	err1 := encoder.Encode(&state)

	if err1 != nil {
		log.Fatalf("state object could not be serialized: %v", err1)
	}

	file, err2 := os.Create("../util/STATE_CONFIG")

	if err2 != nil {
		log.Fatalf("Fail to create STATE_CONFIG: %v", err2)
	}


	_, err3 := file.Write(buffer.Bytes())

	if err3 != nil {
		log.Fatalf("Fail to write the STATE_CONFIG: %v", err3)
	}

	file.Sync()

}