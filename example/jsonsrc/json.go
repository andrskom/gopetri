package jsonsrc

import (
	"encoding/json"
	"errors"
)

func GetByID(id string) ([]byte, error) {
	var registry map[string]json.RawMessage
	if err := json.Unmarshal([]byte(Src), &registry); err != nil {
		return nil, err
	}
	res, ok := registry[id]
	if !ok {
		return nil, errors.New("unregistered petri net cfg")
	}

	return res, nil
}

var Src = `{
  "example_v1": {
    "start": "placeStart",
    "finish": [
      "placeFinish"
    ],
    "places": [
      "placeStart",
      "branch1Place1",
      "branch1Place2",
      "branch2Place1",
      "branchMergePlace1",
      "placeFinish"
    ],
    "transitions": {
      "placeStart__branching": {
        "from": ["placeStart"],
        "to": [
          "branch1Place1",
          "branch2Place1"
        ]
      },
      "branch1Place1__branch1Place2": {
        "from": ["branch1Place1"],
        "to": ["branch1Place2"]
      },
      "branch1Place2_branch2Place1__branchMergePlace1": {
        "from": [
          "branch1Place2",
          "branch2Place1"
        ],
        "to": ["branchMergePlace1"]
      },
      "branchMergePlace1__placeFinish": {
        "from": ["branchMergePlace1"],
        "to": ["placeFinish"]
      }
    }
  }
}`
