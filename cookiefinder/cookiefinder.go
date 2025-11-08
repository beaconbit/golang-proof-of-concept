package cookiefinder

import (
    "log"
    "time"
    "crypto/md5"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "gorm.io/gorm"
    "graphite/publisher/db"
    "graphite/publisher/utils"

    "github.com/PuerkitoBio/goquery"
)

type CookieResult struct {
	Error string
	Value string
}

type CookieFinder struct {
    jobCh               chan utils.Job[*gorm.DB]
    frequencyTicker     *time.Ticker
}

func NewCookieFinder(period time.Duration, jobCh chan utils.Job[*gorm.DB]) *CookieFinder {
    return &CookieFinder{
        jobCh: jobCh,
        frequencyTicker: time.NewTicker(period * time.Second),
    }
}

func (c *CookieFinder) Run() {
    log.Println("cookie finder started")
    for {
        select {
        case <-c.frequencyTicker.C:
            go c.ReadAllAddresses()
            go c.ReadAllDevices()
        }
    }
}

func (c *CookieFinder) ReadAllAddresses() {
    job, resultCh, errCh := db.ReadAllAddresses()
    c.jobCh <- job
    select {
    case err := <-errCh:
        log.Println("error: ", err)
    case result := <-resultCh:
	log.Println("response from database received: all addresses", len(result))
        for i := 0; i < len(result); i++ {
            _ = result[i]
        }
	// testing addDevice function
	device := db.Device{
	    Mac: result[5].Mac,
	    IP: result[5].IP,
	}
	log.Println("try to add device", device)
	go c.AddDevice(device)
    }
}

func (c *CookieFinder) ReadAllDevices() {
    job, resultCh, errCh := db.ReadAllDevices()
    c.jobCh <- job
    select {
    case err := <-errCh:
        log.Println("error: ", err)
    case result := <-resultCh:
	log.Println("response from database received: all devices", len(result))
        for i := 0; i < len(result); i++ {
            r := result[i]
            log.Println(r)
        }
    }
}

func (c *CookieFinder) AddDevice(device db.Device) {
    job, resultCh, errCh := db.AddDevice(device)
    c.jobCh <- job
    select {
    case err := <-errCh:
        log.Println("error: ", err)
    case _ = <-resultCh:
	log.Println("added device to database", device)
    }
}

func (c *CookieFinder) FetchCookie(address db.Address, username string, password string) {

}

func (c *CookieFinder) ConditionalCookieRefresh(mac string) {
    // read from devices table
    // if read fails
        // add device


    // if read succeeds
        // if password exists
            // if cookie timeout is past
                // fetch cookie
                // update database
                // store new timeout
            
            // if cookie timeout in future
                //skip

        // if password missing
           // try password from config  
                // if success
                    // store password in database
}


func getCookie(ip string) CookieResult {
    username := "root"
    password := "ubuntu"

    client := &http.Client{
            Timeout: 5 * time.Second,
    }

    configURL := fmt.Sprintf("http://%s/config", ip)
    resp, err := client.Get(configURL)
    if err != nil {
            return CookieResult{Error: fmt.Sprintf("Timeout when fetching %s", configURL)}
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
            return CookieResult{Error: fmt.Sprintf("Got response status %d when fetching %s", resp.StatusCode, configURL)}
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
            return CookieResult{Error: "Failed to parse HTML"}
    }

    seeddata, exists := doc.Find("input[name='seeddata']").Attr("value")
    if !exists {
            return CookieResult{Error: "Seeddata not found in the response."}
    }

    hashInput := fmt.Sprintf("%s:%s:%s", seeddata, username, password)
    hash := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))

    if len(hash) <= 1 {
            return CookieResult{Error: "Could not create hash"}
    }

    form := url.Values{}
    form.Add("seeddata", seeddata)
    form.Add("authdata", hash)

    req, err := http.NewRequest("POST", configURL+"/index.html", strings.NewReader(form.Encode()))
    if err != nil {
            return CookieResult{Error: "Failed to create POST request"}
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp2, err := client.Do(req)
    if err != nil {
            return CookieResult{Error: "Timeout when requesting cookie"}
    }
    defer resp2.Body.Close()

    if resp2.StatusCode != 200 {
            return CookieResult{Error: fmt.Sprintf("Response status: %d", resp2.StatusCode)}
    }

    for _, c := range resp2.Cookies() {
            return CookieResult{Value: c.Value} // Return first cookie
    }

    return CookieResult{Error: "No cookie found"}
}

