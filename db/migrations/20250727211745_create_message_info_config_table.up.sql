CREATE TABLE message_info_config (
    mac TEXT NOT NULL,
    data_field_index INTEGER NOT NULL,
    ip TEXT NOT NULL,
    source_name TEXT,
    zone TEXT,
    machine TEXT,
    machine_stage TEXT,
    event_type TEXT,
    units TEXT,
    pieces INTEGER,
    estimated_pieces INTEGER,
    rfid TEXT,
    PRIMARY KEY (mac, data_field_index)
);

