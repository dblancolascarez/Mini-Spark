// Serialización JSON/MsgPack
package protocol

import (
	"encoding/json"
)

// Serialize convierte un objeto a JSON
func Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize convierte JSON a un objeto
func Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}