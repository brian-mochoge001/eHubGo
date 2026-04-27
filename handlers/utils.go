package handlers

import (
	"database/sql"
	"time"
)

// NullStringToString converts sql.NullString to *string for JSON
func NullStringToString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// NullInt32ToInt32 converts sql.NullInt32 to *int32 for JSON
func NullInt32ToInt32(ni sql.NullInt32) *int32 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int32
}

// NullBoolToBool converts sql.NullBool to *bool for JSON
func NullBoolToBool(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

// NullTimeToTime converts sql.NullTime to *time.Time for JSON
func NullTimeToTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// NullStringToStringPtr converts sql.NullString to *string for JSON
func NullStringToStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// NullTimeToTimePtr converts sql.NullTime to *time.Time for JSON
func NullTimeToTimePtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
