package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type State struct {
	CurrentTerm int64
	VoteFor     string
	logs        Log
	selfID    int64
}

func InitState(selfID int64) *State {
	var state State
	file, err1 := os.Open(fmt.Sprintf("./src/util/STATE_CONFIG_%d.json", selfID), )
	defer file.Close()

	if err1 != nil {
		state = State{CurrentTerm: 0, VoteFor: "", logs: *NewLog(), selfID:selfID}
	} else {

		contents, err2 := ioutil.ReadAll(file)

		if err2 != nil {
			log.Fatalf("Fail to read the STATE_CONFIG.json: %v", err2)
		}


		state = JsonToObject(contents)

	}


	return &state
}

func (state *State) PersistentStore() {
	//var buffer bytes.Buffer


	bData := ObjectToJson(state)

	file, err2 := os.Create(fmt.Sprintf("./src/util/STATE_CONFIG_%d.json", state.selfID))

	defer file.Close()

	if err2 != nil {
		log.Fatalf("Fail to create STATE_CONFIG.json: %v", err2)
	} else {

		_, err3 := file.Write(bData)

		if err3 != nil {
			log.Fatalf("Fail to write the STATE_CONFIG.json: %v", err3)
		}

		err4 := file.Sync()

		if err4 != nil {
			log.Fatalf("Fail to sync to Disk: %v", err4)
		}
	}

}

func (state *State) GetLog() Log {
	return state.logs
}
func ObjectToJson (state *State) []byte {
	jsonMap := make(map[string]interface{})

	jsonMap["CurrentTerm"] = state.CurrentTerm
	jsonMap["VoteFor"] = state.VoteFor

	var logs []map[string]interface{}

	for _, entry := range state.logs.EntryList {
		tmpMap := map[string]interface{}{"op": entry.op, "key": entry.key, "val": entry.val, "term": entry.term}

		logs = append(logs, tmpMap)
	}

	if logs == nil {
		logs = make([]map[string]interface{}, 0)
	}

	jsonMap["logs"] = logs

	bData, err1 := json.Marshal(jsonMap)
	if err1 != nil {
		log.Fatalf("State could not be serialized : %v", err1)
	}

	return bData
}

func JsonToObject(bData []byte) State {
	var jsonMap map[string]interface{}

	err3 := json.Unmarshal(bData, &jsonMap)

	if err3 != nil {
		log.Fatalf("Fail to Parse Json File: %v", err3)
	}

	state := State{CurrentTerm: int64(jsonMap["CurrentTerm"].(float64)), VoteFor: jsonMap["VoteFor"].(string),
		logs: *NewLog()}


	for _, e := range jsonMap["logs"].([]interface{}) {
		tmpMap := e.(map[string]interface{})

		state.logs.appendEntry(entry{op: tmpMap["op"].(string), key: tmpMap["key"].(string),
			val: tmpMap["val"].(string), term: int64(tmpMap["term"].(float64))})
	}

	return state
}
