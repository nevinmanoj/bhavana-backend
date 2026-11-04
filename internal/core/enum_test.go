package core

import "testing"

func TestAcceptsStudent(t *testing.T) {
	tests := []struct {
		name    string
		event   Category
		student Category
		want    bool
	}{
		{"same category", CategoryHI, CategoryHI, true},
		{"different category", CategoryHI, CategoryMC, false},
		{"common accepts HI", CategoryCommon, CategoryHI, true},
		{"common accepts MC", CategoryCommon, CategoryMC, true},
		{"common accepts PC", CategoryCommon, CategoryPC, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.AcceptsStudent(tt.student); got != tt.want {
				t.Errorf("AcceptsStudent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEventCategory(t *testing.T) {
	tests := []struct {
		in      string
		want    Category
		wantErr bool
	}{
		{"HI", CategoryHI, false},
		{"mc", CategoryMC, false},
		{"PC", CategoryPC, false},
		{"COMMON", CategoryCommon, false},
		{"common", CategoryCommon, false},
		{"XX", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseEventCategory(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseEventCategory(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseEventCategory(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// COMMON is event-only: a student must never be able to hold it.
func TestCommonIsEventOnly(t *testing.T) {
	if _, err := ParseCategory("COMMON"); err == nil {
		t.Error("ParseCategory accepted COMMON")
	}
	if CategoryCommon.IsValid() {
		t.Error("IsValid accepted COMMON")
	}
	if !CategoryCommon.IsValidForEvent() {
		t.Error("IsValidForEvent rejected COMMON")
	}
	if !CategoryHI.IsValidForEvent() {
		t.Error("IsValidForEvent rejected HI")
	}
	if Category("XX").IsValidForEvent() {
		t.Error("IsValidForEvent accepted an unknown category")
	}
}
