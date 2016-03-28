package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/nu7hatch/gouuid"
)

// Generate uuid
func NewUUID() string {
	t, _ := uuid.NewV4()
	return t.String()
}

// ValidateUUID used to check whether a string is a valid uuid
func ValidateUUID(uuid string) error {
	regex := regexp.MustCompile("^[a-z0-9]{8}-[a-z0-9]{4}-[1-5][a-z0-9]{3}-[a-z0-9]{4}-[a-z0-9]{12}$")
	if !regex.MatchString(uuid) {
		return fmt.Errorf(`%v is not a valid uuid`, uuid)
	}
	return nil
}

// Is event with 5% ratio happens
func IsEventWithSpecifiedRatioHappens() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(100/5) == 0
}
