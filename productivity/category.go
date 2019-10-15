package productivity

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// A Category represents a degree of productivity.
type Category uint8

// The set of valid Categories.
const (
	_ Category = iota
	CatVeryDistracting
	CatDistracting
	CatNeutral
	CatProductive
	CatVeryProductive
)

var categoryNames = map[Category]string{
	CatVeryDistracting: "Very Distracting",
	CatDistracting:     "Distracting",
	CatNeutral:         "Neutral",
	CatProductive:      "Productive",
	CatVeryProductive:  "Very Productive",
}

// Name returns the name of the Category.
func (c Category) Name() string {
	name, ok := categoryNames[c]
	if ok {
		return name
	}
	return "%!Category(" + strconv.FormatUint(uint64(c), 10) + ")"
}

func (c Category) String() string {
	name, ok := categoryNames[c]
	if ok {
		return strings.ReplaceAll(name, " ", "")
	}
	return fmt.Sprintf("Category(%d)", uint(c))
}

// Weight returns the weight of the Category, used for productivity score
// calculations.
func (c Category) Weight() uint { return uint(c) - 1 }

// MarshalJSON implements json.Marshaller for a Category.
func (c Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID     uint8  `json:"id"`
		Name   string `json:"name"`
		Weight uint   `json:"weight"`
	}{uint8(c), c.Name(), c.Weight()})
}

// UnmarshalJSON implements json.Unmarshaller for a Category.
func (c *Category) UnmarshalJSON(b []byte) error {
	var data struct {
		ID uint8 `json:"id"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	*c = Category(data.ID)
	return nil
}
