package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Luzifer/go_helpers/str"
)

type aclRule struct {
	Field       string  `hcl:"field"`
	Invert      bool    `hcl:"invert"`
	IsPresent   *bool   `hcl:"present"`
	MatchRegex  *string `hcl:"regexp"`
	MatchString *string `hcl:"equals"`
}

func (a aclRule) Validate() error {
	if a.Field == "" {
		return fmt.Errorf("Field is not set")
	}

	if a.IsPresent == nil && a.MatchRegex == nil && a.MatchString == nil {
		return fmt.Errorf("No matcher (present, regexp, equals) is set")
	}

	if a.MatchRegex != nil {
		if _, err := regexp.Compile(*a.MatchRegex); err != nil {
			return fmt.Errorf("Regexp is invalid: %s", err)
		}
	}

	return nil
}

func (a aclRule) AppliesToFields(fields map[string]string) bool {
	var field, value string

	for f, v := range fields {
		if strings.ToLower(a.Field) == f {
			field = f
			value = v
			break
		}
	}

	if a.IsPresent != nil {
		if !a.Invert && *a.IsPresent && field == "" {
			// Field is expected to be present but isn't, rule does not apply
			return false
		}
		if !a.Invert && !*a.IsPresent && field != "" {
			// Field is expected not to be present but is, rule does not apply
			return false
		}
		if a.Invert && *a.IsPresent && field != "" {
			// Field is expected not to be present but is, rule does not apply
			return false
		}
		if a.Invert && !*a.IsPresent && field == "" {
			// Field is expected to be present but isn't, rule does not apply
			return false
		}

		return true
	}

	if field == "" {
		// We found a rule which has no matching field, rule does not apply
		return false
	}

	if a.MatchString != nil {
		if (*a.MatchString != value) == !a.Invert {
			// Value does not match expected string, rule does not apply
			return false
		}
	}

	if a.MatchRegex != nil {
		if regexp.MustCompile(*a.MatchRegex).MatchString(value) == a.Invert {
			// Value does not match expected regexp, rule does not apply
			return false
		}
	}

	return true
}

type aclAccessResult uint

const (
	accessDunno aclAccessResult = iota
	accessAllow
	accessDeny
)

type aclRuleSet struct {
	Rules []aclRule `hcl:"rule"`

	Allow []string `hcl:"allow"`
	Deny  []string `hcl:"deny"`
}

func (a aclRuleSet) buildFieldSet(r *http.Request) map[string]string {
	result := map[string]string{}

	for k := range r.Header {
		result[strings.ToLower(k)] = r.Header.Get(k)
	}

	return result
}

func (a aclRuleSet) HasAccess(user string, groups []string, r *http.Request) aclAccessResult {
	fields := a.buildFieldSet(r)

	for _, rule := range a.Rules {
		if !rule.AppliesToFields(fields) {
			// At least one rule does not match the request
			return accessDunno
		}
	}

	// All rules do apply to this request, we can judge

	if str.StringInSlice(user, a.Deny) {
		// Explicit deny, final result
		return accessDeny
	}

	if str.StringInSlice(user, a.Allow) {
		// Explicit allow, final result
		return accessAllow
	}

	for _, group := range groups {
		if str.StringInSlice("@"+group, a.Deny) {
			// Deny through group, final result
			return accessDeny
		}

		if str.StringInSlice("@"+group, a.Allow) {
			// Allow through group, final result
			return accessAllow
		}
	}

	// Neither user nor group are handled
	return accessDunno
}

func (a aclRuleSet) Validate() error {
	for i, r := range a.Rules {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("Rule on position %d is invalid: %s", i+1, err)
		}
	}

	return nil
}

type acl struct {
	RuleSets []aclRuleSet `hcl:"rule_set"`
}

func (a acl) Validate() error {
	for i, r := range a.RuleSets {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("RuleSet on position %d is invalid: %s", i+1, err)
		}
	}

	return nil
}

func (a acl) HasAccess(user string, groups []string, r *http.Request) bool {
	result := accessDunno

	for _, rs := range a.RuleSets {
		if intermediateResult := rs.HasAccess(user, groups, r); intermediateResult > result {
			result = intermediateResult
		}
	}

	return result == accessAllow
}
