package types

// User defines the structure for a user.
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}
