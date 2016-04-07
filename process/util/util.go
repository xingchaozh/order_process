package util

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
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

// Get current path of app
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logrus.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func JoinPath(path string, subPath string) string {
	return filepath.Join(path, subPath)
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0744)
}
