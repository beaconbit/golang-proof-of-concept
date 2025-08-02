package authflow

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)



type SpindleDeviceAuth struct{}

func init() {
    logger := logPrefix("init")
    logger.Println("Calling Register(spindle_device)")
    Register("spindle_device", &SpindleDeviceAuth{})
}

func (a *SpindleDeviceAuth) Authenticate(device map[string]string) (string, error) {
    logger := logPrefix("Authenticate")

    ip := device["ip"]
    username := device["username"]
    password := device["password"]
    configURL := fmt.Sprintf("http://%s/config", ip)

    resp, err := http.Get(configURL)
    if err != nil {
	return "", fmt.Errorf("failed to get seed data: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
	return "", fmt.Errorf("bad status from config page: %d", resp.StatusCode)
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
	return "", fmt.Errorf("failed to parse HTML: %w", err)
    }

    seed, exists := doc.Find("input[name='seeddata']").Attr("value")
    if !exists {
	return "", errors.New("seeddata not found in HTML")
    }
    logger.Println("seeddata: %s", seed)

    //hashInput := fmt.Sprintf("%s:%s:%s", seed, username, password)
    //hash := md5.Sum([]byte(hashInput))
    //hashStr := fmt.Sprintf("%x", hash)

    //loginURL := fmt.Sprintf("http://%s/config/index.html", ip)
    //form := url.Values{}
    //form.Add("seeddata", seed)
    //form.Add("authdata", hashStr)

    //req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
    //if err != nil {
    //    return "", err
    //}
    //req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    //client := &http.Client{}
    //resp, err = client.Do(req)
    //if err != nil {
    //    return "", fmt.Errorf("login request failed: %w", err)
    //}
    //defer resp.Body.Close()

    //if resp.StatusCode != 200 {
    //    return "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
    //}

    //for _, cookie := range resp.Cookies() {
    //    if strings.HasPrefix(cookie.Name, "adamsessionid") {
    //    	return cookie.Value, nil
    //    }
    //}

    return "", errors.New("no valid session cookie returned")
}

