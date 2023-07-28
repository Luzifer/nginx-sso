package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	aclTestUser   = "test"
	aclTestGroups = []string{"group_a", "group_b"}
	ptrBoolTrue   = func(v bool) *bool { return &v }(true)
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

	assert.True(t, r.AppliesToRequest(aclTestRequest(fields)))

	delete(fields, "field_c")
	assert.False(t, r.AppliesToRequest(aclTestRequest(fields)))
}

func TestGroupAuthenticated(t *testing.T) {
	a := acl{RuleSets: []aclRuleSet{{
		Rules: []aclRule{
			{
				Field:       "field_a",
				MatchString: aclTestString("expected"),
			},
		},
		Allow: []string{"@_authenticated"},
	}}}
	fields := map[string]string{
		"field_a": "expected",
	}

	assert.True(t, a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(fields)))
	assert.False(t, a.HasAccess("\x00", nil, aclTestRequest(fields)), "access to anon user")
	assert.False(t, a.HasAccess("", nil, aclTestRequest(fields)), "access to empty user")

	a.RuleSets[0].Allow = []string{"testgroup"}
	assert.False(t, a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(fields)))
}

func TestAnonymousAccess(t *testing.T) {
	a := acl{RuleSets: []aclRuleSet{{
		Rules: []aclRule{
			{
				Field:       "field_a",
				MatchString: aclTestString("expected"),
			},
		},
		Allow: []string{groupAnonymous},
	}}}
	fields := map[string]string{
		"field_a": "expected",
	}

	assert.True(t, a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(fields)))
	assert.True(t, a.HasAccess("", nil, aclTestRequest(fields)), "access to empty user")
	assert.True(t, a.HasAccess("\x00", nil, aclTestRequest(fields)), "access to anon user")
}

func TestAnonymousAccessExplicitDeny(t *testing.T) {
	a := acl{
		RuleSets: []aclRuleSet{
			{
				Rules: []aclRule{{Field: "field_a", IsPresent: ptrBoolTrue}},
				Allow: []string{groupAnonymous},
			},
			{
				Rules: []aclRule{{Field: "field_b", IsPresent: ptrBoolTrue}},
				Allow: []string{"somerandomuser"},
				Deny:  []string{groupAnonymous},
			},
		},
	}

	assert.True(t,
		a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(map[string]string{"field_a": ""})),
		"anon access with only allowed field should be possible")
	assert.False(t,
		a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(map[string]string{"field_b": ""})),
		"anon access with only denied field should not be possible")
	assert.False(t,
		a.HasAccess(aclTestUser, aclTestGroups, aclTestRequest(map[string]string{"field_a": "", "field_b": ""})),
		"anon access with one allowed and one denied field should not be possible")
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
