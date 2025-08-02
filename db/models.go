package db

type Address struct {
    Mac           string         `gorm:"primaryKey;column:mac"`
    IP            string         `gorm:"not null;column:ip"`
}

type Device struct {
    Mac           string         `gorm:"primaryKey;column:mac"`
    IP            string         `gorm:"not null;column:ip"`
    Valid         bool           `gorm:"default:true;column:valid"`
    Failures      int            `gorm:"default:0;column:failures"`
    Username      *string        `gorm:"column:username"`        // Nullable
    Password      *string        `gorm:"column:password"`        // Nullable
    Cookie        *string        `gorm:"type:text;column:cookie"`// Nullable
    CookieExpires int64          `gorm:"default:0;column:cookie_expires"` // Unix timestamp
    AuthFlow      *string        `gorm:"column:auth_flow"`       // Nullable
    Scraper       *string        `gorm:"column:scraper"`         // Nullable
    LastData      *string        `gorm:"type:text;column:last_data"` // Nullable
    LastSeen      *int64         `gorm:"column:last_seen"`       // Unix timestamp
}

type MessageInfoConfig struct {
    Mac             string  `gorm:"primaryKey;column:mac;not null"`
    DataFieldIndex  int     `gorm:"primaryKey;column:data_field_index;not null"`
    IP              string  `gorm:"column:ip;not null"`
    SourceName      *string `gorm:"column:source_name"`       // Nullable
    Zone            *string `gorm:"column:zone"`              // Nullable
    Machine         *string `gorm:"column:machine"`           // Nullable
    MachineStage    *string `gorm:"column:machine_stage"`     // Nullable
    EventType       *string `gorm:"column:event_type"`        // Nullable
    Units           *string `gorm:"column:units"`             // Nullable
    Pieces          *int    `gorm:"column:pieces"`            // Nullable
    EstimatedPieces *int    `gorm:"column:estimated_pieces"`  // Nullable
    RFID            *string `gorm:"column:rfid"`              // Nullable
}
