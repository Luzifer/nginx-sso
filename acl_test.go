package main

import (
	"net/http"
	"testing"
)

var (
	aclTestUser   = "test"
	aclTestGroups = []string{"group_a", "group_b"}
)

func aclTestRequest(headers map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "http://localhost/auth", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

func aclTestString(in string) *string { return &in }
func aclTestBool(in bool) *bool       { return &in }

func TestEmptyACL(t *testing.T) {
	a := acl{}

	if a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(map[string]string{})) {
		t.Fatal("Empty ACL (= default action) was ALLOW instead of DENY")
	}
}

func TestRuleSetMatcher(t *testing.T) {
	r := aclRuleSet{
		Rules: []aclRule{
			{
				Field:       "field_a",
				MatchString: aclTestString("expected"),
			},
			{
				Field:       "field_c",
				MatchString: aclTestString("expected"),
			},
		},
		Allow: []string{aclTestUser},
	}
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
		"field_c": "expected",
	}

	if r.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(fields)) != accessAllow {
		t.Error("Access was denied")
	}

	delete(fields, "field_c")
	if r.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(fields)) != accessDunno {
		t.Error("Access was not unknown")
	}
}

func TestInvertedRegexMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:      "field_a",
		Invert:     true,
		MatchRegex: aclTestString("^expected$"),
	}

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}

	fields["field_a"] = "unexpected"

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}
}

func TestRegexMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:      "field_a",
		MatchRegex: aclTestString("^expected$"),
	}

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}

	fields["field_a"] = "unexpected"

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}
}

func TestInvertedEqualsMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:       "field_a",
		Invert:      true,
		MatchString: aclTestString("expected"),
	}

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}

	fields["field_a"] = "unexpected"

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}
}

func TestEqualsMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:       "field_a",
		MatchString: aclTestString("expected"),
	}

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}

	fields["field_a"] = "unexpected"

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}
}

func TestInvertedIsPresentMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:     "field_a",
		Invert:    true,
		IsPresent: aclTestBool(true),
	}

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(false)

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(true)
	delete(fields, "field_a")

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(false)
	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}
}

func TestIsPresentMatcher(t *testing.T) {
	fields := map[string]string{
		"field_a": "expected",
		"field_b": "unchecked",
	}

	ar := aclRule{
		Field:     "field_a",
		IsPresent: aclTestBool(true),
	}

	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(false)

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(true)
	delete(fields, "field_a")

	if ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v matches fields %#v", ar, fields)
	}

	ar.IsPresent = aclTestBool(false)
	if !ar.AppliesToFields(fields) {
		t.Errorf("Rule %#v does not match fields %#v", ar, fields)
	}
}
