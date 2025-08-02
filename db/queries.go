package db

import (
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

// BulkUpsertAddresses returns a job that bulk-upserts the provided addresses.
func BulkUpsertAddresses(addresses []Address) (func(*gorm.DB), <-chan []Address, <-chan error) {
    resultCh := make(chan []Address, 1)
    errCh := make(chan error, 1)
    job := func(db *gorm.DB) {
        defer close(resultCh)
        defer close(errCh)
        var allRows []Address

        if len(addresses) == 0 {
            // query all rows
            existing := db.Find(&allRows)
            if existing.Error != nil {
                errCh <- existing.Error
            } else {
                resultCh <- allRows
            }
            return
        }

        result := db.Clauses(clause.OnConflict{
            Columns:   []clause.Column{{Name: "mac"}}, // conflict target
            DoUpdates: clause.AssignmentColumns([]string{"ip"}),
        }).Create(&addresses)
        
        if result.Error != nil {
            errCh <- result.Error
        } else {
            existing := db.Find(&allRows)
            if existing.Error != nil {
                errCh <- existing.Error
            } else {
                resultCh <- allRows
            }
        }
    }
    return job, resultCh, errCh
}

// ReadAllAddresses returns a job that reads all addresses and logs them.
func ReadAllAddresses() (func(*gorm.DB), <-chan []Address, <-chan error) {
    resultCh := make(chan []Address, 1)
    errCh := make(chan error, 1)
    job := func(db *gorm.DB) {
        var addresses []Address
        result := db.Find(&addresses)
        if result.Error != nil {
            errCh <- result.Error
        } else {
            resultCh <- addresses
        }
    }
    return job, resultCh, errCh
}

// ReadAddressByMAC returns a job that finds a single address by MAC.
func ReadAddressByMAC(mac string) (func(*gorm.DB), <-chan Address, <-chan error) {
    resultCh := make(chan Address, 1)
    errCh := make(chan error, 1)

    job := func(db *gorm.DB) {
        defer close(resultCh)
        defer close(errCh)

        var addr Address
        result := db.First(&addr, "mac = ?", mac)
        if result.Error != nil {
            errCh <- result.Error
        } else {
            resultCh <- addr
        }
    }

    return job, resultCh, errCh
}


// ReadAddressByIP returns a job that finds a single address by IP.
func ReadAddressByIP(ip string) (func(*gorm.DB), <-chan Address, <-chan error) {
    resultCh := make(chan Address, 1)
    errCh := make(chan error, 1)
    job := func(db *gorm.DB) {
        var addr Address
        result := db.First(&addr, "ip = ?", ip)
        if result.Error != nil {
            errCh <- result.Error
        } else {
            resultCh <- addr
        }
    }
    return job, resultCh, errCh
}

