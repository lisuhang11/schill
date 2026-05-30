package model

import "encoding/json"

type CanalMessage struct {
	Data      []json.RawMessage `json:"data"`
	Database  string            `json:"database"`
	Table     string            `json:"table"`
	Type      string            `json:"type"`
	IsDDL     bool              `json:"isDdl"`
	Es        int64             `json:"es"`
	Timestamp int64             `json:"ts"`
	Old       []json.RawMessage `json:"old"`
	SQL       string            `json:"sql"`
}
