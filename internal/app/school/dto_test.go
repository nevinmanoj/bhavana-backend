package school

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	domain "github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/validation"
)

func intp(n int) *int { return &n }

func validFields() SchoolFields {
	return SchoolFields{
		Name: "  St Mary  ", Address: "Road 1", ContactName: "Anu",
		ContactEmail: "  Anu@Example.COM ", ContactPhone: "+91 98765 43210",
		PrincipalName: "Mr P", PrincipalPhone: "098765-43211",
		ParticipantsMC: intp(0), ParticipantsPC: intp(3), ParticipantsHI: intp(5), NonParticipants: intp(0),
		TotalStaff: intp(2), TotalParents: intp(0),
	}
}

func TestSchoolFieldsNormalizeAndValidate(t *testing.T) {
	v := validation.NewValidator()

	f := validFields()
	f.normalize()
	if err := v.Struct(RegisterSchoolRequest{SchoolFields: f}); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	if f.ContactEmail != "anu@example.com" || f.ContactPhone != "9876543210" || f.PrincipalPhone != "9876543211" || f.Name != "St Mary" {
		t.Errorf("not normalized: %+v", f)
	}
	if s := f.toSchool(); s.ParticipantsMC != 0 || s.ParticipantsPC != 3 || s.TotalStaff != 2 || s.TotalParents != 0 {
		t.Errorf("headcounts not mapped: %+v", s)
	}

	bad := map[string]func(*SchoolFields){
		"placeholder principal phone": func(f *SchoolFields) { f.PrincipalPhone = "-" },
		"short contact phone":         func(f *SchoolFields) { f.ContactPhone = "12345" },
		"landline-style number":       func(f *SchoolFields) { f.ContactPhone = "1234567890" },
		"missing headcount":           func(f *SchoolFields) { f.NonParticipants = nil },
		"missing staff count":         func(f *SchoolFields) { f.TotalStaff = nil },
		"missing parent count":        func(f *SchoolFields) { f.TotalParents = nil },
		"negative staff count":        func(f *SchoolFields) { f.TotalStaff = intp(-1) },
		"negative parent count":       func(f *SchoolFields) { f.TotalParents = intp(-1) },
		"negative headcount":          func(f *SchoolFields) { f.ParticipantsHI = intp(-1) },
		"blank principal":             func(f *SchoolFields) { f.PrincipalName = "   " },
		"bad email":                   func(f *SchoolFields) { f.ContactEmail = "nope" },
	}
	for name, mutate := range bad {
		f := validFields()
		mutate(&f)
		f.normalize()
		if err := v.Struct(CreateSchoolRequest{SchoolFields: f}); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

// Registration no longer takes a password: acceptance emails a generated one.
// Handlers decode with DisallowUnknownFields, so a stale client still sending
// one gets a clear error instead of having it silently ignored.
func TestRegisterSchoolRequestRejectsPassword(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"name":"x","password":"secret1"}`))
	dec.DisallowUnknownFields()
	var req RegisterSchoolRequest
	if err := dec.Decode(&req); err == nil {
		t.Error("password field accepted")
	}
}

func TestParseDOB(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Format(time.DateOnly)
	cases := []struct {
		in string
		ok bool
	}{
		{"2012-05-14", true},
		{"1900-01-01", true},
		{"1899-12-31", false},
		{"0212-05-14", false},
		{tomorrow, false},
		{"2012-02-30", false},
		{"14/05/2012", false},
		{"", false},
	}
	for _, c := range cases {
		if _, err := parseDOB(c.in); (err == nil) != c.ok {
			t.Errorf("parseDOB(%q) err = %v, want ok = %v", c.in, err, c.ok)
		}
	}
}

func TestToStudentResponseDOB(t *testing.T) {
	dob := time.Date(2012, time.May, 14, 0, 0, 0, 0, time.UTC)
	if got := ToStudentResponse(&domain.Student{DOB: &dob}).DOB; got == nil || *got != "2012-05-14" {
		t.Errorf("DOB = %v, want 2012-05-14", got)
	}
	if got := ToStudentResponse(&domain.Student{}).DOB; got != nil {
		t.Errorf("DOB = %v, want nil for a student with no date of birth", *got)
	}
}
