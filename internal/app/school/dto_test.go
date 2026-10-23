package school

import (
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/validation"
)

func intp(n int) *int { return &n }

func validFields() SchoolFields {
	return SchoolFields{
		Name: "  St Mary  ", Address: "Road 1", ContactName: "Anu",
		ContactEmail: "  Anu@Example.COM ", ContactPhone: "+91 98765 43210",
		PrincipalName: "Mr P", PrincipalPhone: "098765-43211",
		ParticipantsMC: intp(0), ParticipantsPC: intp(3), ParticipantsHC: intp(5), NonParticipants: intp(0),
	}
}

func TestSchoolFieldsNormalizeAndValidate(t *testing.T) {
	v := validation.NewValidator()

	f := validFields()
	f.normalize()
	if err := v.Struct(RegisterSchoolRequest{SchoolFields: f, Password: "secret1"}); err != nil {
		t.Fatalf("valid registration rejected: %v", err)
	}
	if f.ContactEmail != "anu@example.com" || f.ContactPhone != "9876543210" || f.PrincipalPhone != "9876543211" || f.Name != "St Mary" {
		t.Errorf("not normalized: %+v", f)
	}
	if s := f.toSchool(); s.ParticipantsMC != 0 || s.ParticipantsPC != 3 {
		t.Errorf("headcounts not mapped: %+v", s)
	}

	bad := map[string]func(*SchoolFields){
		"placeholder principal phone": func(f *SchoolFields) { f.PrincipalPhone = "-" },
		"short contact phone":         func(f *SchoolFields) { f.ContactPhone = "12345" },
		"landline-style number":       func(f *SchoolFields) { f.ContactPhone = "1234567890" },
		"missing headcount":           func(f *SchoolFields) { f.NonParticipants = nil },
		"negative headcount":          func(f *SchoolFields) { f.ParticipantsHC = intp(-1) },
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

	// short password only matters for registration
	f = validFields()
	f.normalize()
	if err := v.Struct(RegisterSchoolRequest{SchoolFields: f, Password: "123"}); err == nil {
		t.Error("short password accepted")
	}
}
