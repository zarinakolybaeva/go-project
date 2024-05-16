package data

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Define an error that our UnmarshalJSON() method
// can return if we're unable to parse  or convert the JSON string successfully.
var ErrInvalidTimeFormat = errors.New("invalid Time format")

type CustomTime time.Time

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formattedTime := time.Time(ct).Format("2006-01-02 15:04:05")
	quotedJSONValue := strconv.Quote(formattedTime)
	return []byte(quotedJSONValue), nil
}

// Implement the database/sql/driver Val() method to convert CustomTime to a value that can be stored in the database.
func (ct CustomTime) Value() (driver.Value, error) {
	return time.Time(ct), nil
}

// Implement the database/sql/driver Scan() method to convert a database value to a CustomTime.
func (ct *CustomTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*ct = CustomTime(v)
		return nil
	case nil:
		*ct = CustomTime(time.Time{})
		return nil
	default:
		return errors.New("unsupported type for CustomTime")
	}
}

// Implement a UnmarshalJSON() method on the Runtime type so that it satisfies the json.Unmarshaler interface.
// IMPORTANT: Because UnmarshalJSON() needs to modify the receiver (our Runtime type),
// we must use a pointer receiver for this to work correctly.
// Otherwise, we will only be modifying a copy (which is then discarded when this method returns).
func (ct *CustomTime) UnmarshalJSON(jsonValue []byte) error {
	// We expect the incoming JSON value to be a string in the format "YYYY-MM-DD HH:MM:SS",
	// so we first remove the surrounding double-quotes from this string.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidTimeFormat
	}

	// Now, parse the unquoted JSON string into a time.Time value using a specified layout.
	// If the layout doesn't match the expected format, return the ErrInvalidTimeFormat error.
	const layout = "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(layout, unquotedJSONValue)
	fmt.Println(parsedTime)
	if err != nil {
		return ErrInvalidTimeFormat
	}

	// Convert the parsed time.Time value to the CustomTime type and assign it to the receiver.
	// Use the * operator to dereference the receiver (which is a pointer to CustomTime)
	// 		to set the underlying value of the pointer.
	*ct = CustomTime(parsedTime)
	return nil
}

func (ct CustomTime) IsZero() bool {
	return time.Time(ct).IsZero()
}

func (ct CustomTime) Before(t time.Time) bool {
	return time.Time(ct).Before(t)
}

func (ct CustomTime) After(t time.Time) bool {
	return time.Time(ct).After(t)
}
