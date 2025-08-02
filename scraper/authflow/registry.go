package authflow

import (
    "crypto/md5"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "strings"

    "github.com/PuerkitoBio/goquery"
)



// AuthFlow is the interface that all authentication strategies must implement.
type AuthFlow interface {
    Authenticate(device map[string]string) (string, error) // returns cookie
}

var registry = map[string]AuthFlow{}

func Register(name string, flow AuthFlow) {
    logger := logPrefix("Register")
    if _, exists := registry[name]; exists {
	log.Fatalf("auth flow already registered: %s", name)
    }
    registry[name] = flow
    logger.Println("Successful authflow registration: %s", name)
}

func Get(name string) (AuthFlow, bool) {
    flow, ok := registry[name]
    return flow, ok
}

