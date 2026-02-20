package main

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

const (
	groupAnonymous     = "@_anonymous"
	groupAuthenticated = "@_authenticated"
)

type (
	acl struct {
		RuleSets []aclRuleSet `yaml:"rule_sets"`
	}

	aclRule struct {
		Field       string  `yaml:"field"`
		Invert      bool    `yaml:"invert"`
		IsPresent   *bool   `yaml:"present"`
		MatchRegex  *string `yaml:"regexp"`
		MatchString *string `yaml:"equals"`
	}

	aclAccessResult uint

	aclRuleSet struct {
		Rules []aclRule `yaml:"rules"`

		Allow []string `yaml:"allow"`
		Deny  []string `yaml:"deny"`
	}
)

const (
	accessDunno aclAccessResult = iota
	accessAllow
	accessDeny
)

// --- ACL

func (a acl) HasAccess(user string, groups []string, r *http.Request) bool {
	var (
		collectionAllow = map[string]bool{}
		collectionDeny  = map[string]bool{}
	)

	for _, rs := range a.RuleSets {
		if !rs.AppliesToRequest(r) {
			continue
		}

		// Collect the allows from all matching rulesets
		for _, a := range rs.Allow {
			collectionAllow[a] = true
		}

		// Collect the denies from all matching rulesets
		for _, d := range rs.Deny {
			collectionDeny[d] = true
		}
	}

	// Form lists from the collections
	var allowed, denied []string

	for k := range collectionAllow {
		allowed = append(allowed, k)
	}
	for k := range collectionDeny {
		denied = append(denied, k)
	}

	return a.checkAccess(user, groups, allowed, denied)
}

func (a acl) Validate() error {
	for i, r := range a.RuleSets {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("RuleSet on position %d is invalid: %s", i+1, err)
		}
	}

	return nil
}

func (a acl) checkAccess(user string, groups, allowed, denied []string) bool {
	if !slices.Contains([]string{"", "\x00"}, user) {
		// The user is set to a non-anon user, we add the pseudo-group
		// authenticated to the groups list the user has
		groups = append(groups, groupAuthenticated)
	} else {
		// The user did match anon, therefore we set the pseudo-group
		// for anonymous to be used in group matching
		groups = []string{groupAnonymous}
	}

	// Quoting the documentation here:
	// "There is a simple logic: Users before groups, denies before allows."

	// Lets check the user
	if slices.Contains(denied, user) {
		// Explicit deny on the user, they're out!
		return false
	}

	if slices.Contains(allowed, user) {
		// Explicit allow on the user, they're in!
		return true
	}

	// The user yielded no result, lets check the groups
	for _, group := range groups {
		if slices.Contains(denied, a.fixGroupName(group)) {
			// The group is denied access
			return false
		}

		if slices.Contains(allowed, a.fixGroupName(group)) {
			// The group is allowed access
			return true
		}
	}

	// We found no match for the user and/or group. Last chance is
	// no ruleset denied anonymous access and at least one ruleset
	// enabled anonymous access
	if !slices.Contains(denied, groupAnonymous) && slices.Contains(allowed, groupAnonymous) {
		return true
	}

	// We found neither a user nor a group with access or deny config
	// so we fall back to the default: No access.
	return false
}

func (acl) fixGroupName(group string) string {
	return "@" + strings.TrimLeft(group, "@")
}

// --- ACL Rule

// AppliesToFields checks whether the given rule conditions matches
// the given fields
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

func (a aclRule) Validate() error {
	if a.Field == "" {
		return fmt.Errorf("field is not set")
	}

	if a.IsPresent == nil && a.MatchRegex == nil && a.MatchString == nil {
		return fmt.Errorf("no matcher (present, regexp, equals) is set")
	}

	if a.MatchRegex != nil {
		if _, err := regexp.Compile(*a.MatchRegex); err != nil {
			return fmt.Errorf("regexp is invalid: %s", err)
		}
	}

	return nil
}

// --- ACL Rule Set

// AppliesToRequest checks whether every rule in the aclRuleSet
// matches the http.Request. If not this rule-set must not be applied
// to the given request
func (a aclRuleSet) AppliesToRequest(r *http.Request) bool {
	fields := a.buildFieldSet(r)

	for _, rule := range a.Rules {
		if !rule.AppliesToFields(fields) {
			// At least one rule does not match the request
			return false
		}
	}

	return true
}

func (a aclRuleSet) Validate() error {
	for i, r := range a.Rules {
		if err := r.Validate(); err != nil {
			return fmt.Errorf("rule on position %d is invalid: %s", i+1, err)
		}
	}

	return nil
}

func (a aclRuleSet) buildFieldSet(r *http.Request) map[string]string {
	result := map[string]string{}

	for k := range r.Header {
		result[strings.ToLower(k)] = r.Header.Get(k)
	}

	return result
}
