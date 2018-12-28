package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/duosecurity/duo_api_golang"
)

// Client provides access to Duo's admin API.
type Client struct {
	duoapi.DuoApi
}

type ListResultMetadata struct {
	NextOffset   json.Number `json:"next_offset"`
	PrevOffset   json.Number `json:"prev_offset"`
	TotalObjects json.Number `json:"total_objects"`
}

type ListResult struct {
	Metadata ListResultMetadata `json:"metadata"`
}

func (l *ListResult) metadata() ListResultMetadata {
	return l.Metadata
}

// New initializes an admin API Client struct.
func New(base duoapi.DuoApi) *Client {
	return &Client{base}
}

// User models a single user.
type User struct {
	Alias1            *string
	Alias2            *string
	Alias3            *string
	Alias4            *string
	Created           uint64
	Email             string
	FirstName         *string
	Groups            []Group
	LastDirectorySync *uint64 `json:"last_directory_sync"`
	LastLogin         *uint64 `json:"last_login"`
	LastName          *string
	Notes             string
	Phones            []Phone
	RealName          *string
	Status            string
	Tokens            []Token
	UserID            string `json:"user_id"`
	Username          string
}

// Group models a group to which users may belong.
type Group struct {
	Desc             string
	GroupID          string `json:"group_id"`
	MobileOTPEnabled bool   `json:"mobile_otp_enabled"`
	Name             string
	PushEnabled      bool `json:"push_enabled"`
	SMSEnabled       bool `json:"sms_enabled"`
	Status           string
	VoiceEnabled     bool `json:"voice_enabled"`
}

// Phone models a user's phone.
type Phone struct {
	Activated        bool
	Capabilities     []string
	Encrypted        string
	Extension        string
	Fingerprint      string
	Name             string
	Number           string
	PhoneID          string `json:"phone_id"`
	Platform         string
	Postdelay        string
	Predelay         string
	Screenlock       string
	SMSPasscodesSent bool
	Type             string
	Users            []User
}

// Token models a hardware security token.
type Token struct {
	TokenID  string `json:"token_id"`
	Type     string
	Serial   string
	TOTPStep *int `json:"totp_step"`
	Users    []User
}

// U2FToken models a U2F security token.
type U2FToken struct {
	DateAdded      uint64 `json:"date_added"`
	RegistrationID string `json:"registration_id"`
	User           *User
}

// Common URL options

// Limit sets the optional limit parameter for an API request.
func Limit(limit uint64) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("limit", strconv.FormatUint(limit, 10))
	}
}

// Offset sets the optional offset parameter for an API request.
func Offset(offset uint64) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("offset", strconv.FormatUint(offset, 10))
	}
}

// User methods

// GetUsersUsername sets the optional username parameter for a GetUsers request.
func GetUsersUsername(name string) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("username", name)
	}
}

// GetUsersResult models responses containing a list of users.
type GetUsersResult struct {
	duoapi.StatResult
	ListResult
	Response []User
}

func (result *GetUsersResult) getResponse() interface{} {
	return result.Response
}

func (result *GetUsersResult) appendResponse(users interface{}) {
	asserted_users := users.([]User)
	result.Response = append(result.Response, asserted_users...)
}

// GetUsers calls GET /admin/v1/users
// See https://duo.com/docs/adminapi#retrieve-users
func (c *Client) GetUsers(options ...func(*url.Values)) (*GetUsersResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveUsers(params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetUsersResult), nil
}

type responsePage interface {
	metadata() ListResultMetadata
	getResponse() interface{}
	appendResponse(interface{})
}

type pageFetcher func(params url.Values) (responsePage, error)

func (c *Client) retrieveItems(
	params url.Values,
	fetcher pageFetcher,
) (responsePage, error) {
	if params.Get("offset") == "" {
		params.Set("offset", "0")
	}

	if params.Get("limit") == "" {
		params.Set("limit", "100")
		accumulator, firstErr := fetcher(params)

		if firstErr != nil {
			return nil, firstErr
		}

		params.Set("offset", accumulator.metadata().NextOffset.String())
		for ; params.Get("offset") != "" ; {
			nextResult, err := fetcher(params)
			if err != nil {
				return nil, err
			}
			nextResult.appendResponse(accumulator.getResponse())
			accumulator = nextResult
			params.Set("offset", accumulator.metadata().NextOffset.String())
		}
		return accumulator, nil
	}

	return fetcher(params)
}

func (c *Client) retrieveUsers(params url.Values) (*GetUsersResult, error) {
	_, body, err := c.SignedCall(http.MethodGet, "/admin/v1/users", params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetUsersResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUser calls GET /admin/v1/users/:user_id
// See https://duo.com/docs/adminapi#retrieve-user-by-id
func (c *Client) GetUser(userID string) (*GetUsersResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s", userID)

	_, body, err := c.SignedCall(http.MethodGet, path, nil, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetUsersResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserGroups calls GET /admin/v1/users/:user_id/groups
// See https://duo.com/docs/adminapi#retrieve-groups-by-user-id
func (c *Client) GetUserGroups(userID string, options ...func(*url.Values)) (*GetGroupsResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveUserGroups(userID, params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetGroupsResult), nil
}

func (c *Client) retrieveUserGroups(userID string, params url.Values) (*GetGroupsResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s/groups", userID)

	_, body, err := c.SignedCall(http.MethodGet, path, params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetGroupsResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserPhones calls GET /admin/v1/users/:user_id/phones
// See https://duo.com/docs/adminapi#retrieve-phones-by-user-id
func (c *Client) GetUserPhones(userID string, options ...func(*url.Values)) (*GetPhonesResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveUserPhones(userID, params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetPhonesResult), nil
}

func (c *Client) retrieveUserPhones(userID string, params url.Values) (*GetPhonesResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s/phones", userID)

	_, body, err := c.SignedCall(http.MethodGet, path, params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetPhonesResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserTokens calls GET /admin/v1/users/:user_id/tokens
// See https://duo.com/docs/adminapi#retrieve-hardware-tokens-by-user-id
func (c *Client) GetUserTokens(userID string, options ...func(*url.Values)) (*GetTokensResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveUserTokens(userID, params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetTokensResult), nil
}

func (c *Client) retrieveUserTokens(userID string, params url.Values) (*GetTokensResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s/tokens", userID)

	_, body, err := c.SignedCall(http.MethodGet, path, params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetTokensResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StringResult models responses containing a simple string.
type StringResult struct {
	duoapi.StatResult
	Response string
}

// AssociateUserToken calls POST /admin/v1/users/:user_id/tokens
// See https://duo.com/docs/adminapi#associate-hardware-token-with-user
func (c *Client) AssociateUserToken(userID, tokenID string) (*StringResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s/tokens", userID)

	params := url.Values{}
	params.Set("token_id", tokenID)

	_, body, err := c.SignedCall(http.MethodPost, path, params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &StringResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetUserU2FTokens calls GET /admin/v1/users/:user_id/u2ftokens
// See https://duo.com/docs/adminapi#retrieve-u2f-tokens-by-user-id
func (c *Client) GetUserU2FTokens(userID string, options ...func(*url.Values)) (*GetU2FTokensResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveUserU2FTokens(userID, params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetU2FTokensResult), nil
}

func (c *Client) retrieveUserU2FTokens(userID string, params url.Values) (*GetU2FTokensResult, error) {
	path := fmt.Sprintf("/admin/v1/users/%s/u2ftokens", userID)

	_, body, err := c.SignedCall(http.MethodGet, path, params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetU2FTokensResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Group methods

// GetGroupsResult models responses containing a list of groups.
type GetGroupsResult struct {
	duoapi.StatResult
	ListResult
	Response []Group
}

func (result *GetGroupsResult) getResponse() interface{} {
	return result.Response
}

func (result *GetGroupsResult) appendResponse(groups interface{}) {
	asserted_groups := groups.([]Group)
	result.Response = append(result.Response, asserted_groups...)
}

// GetGroups calls GET /admin/v1/groups
// See https://duo.com/docs/adminapi#retrieve-groups
func (c *Client) GetGroups(options ...func(*url.Values)) (*GetGroupsResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveGroups(params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetGroupsResult), nil
}

func (c *Client) retrieveGroups(params url.Values) (*GetGroupsResult, error) {
	_, body, err := c.SignedCall(http.MethodGet, "/admin/v1/groups", params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetGroupsResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetGroupResult models responses containing a single group.
type GetGroupResult struct {
	duoapi.StatResult
	Response Group
}

// GetGroup calls GET /admin/v2/group/:group_id
// See https://duo.com/docs/adminapi#get-group-info
func (c *Client) GetGroup(groupID string) (*GetGroupResult, error) {
	path := fmt.Sprintf("/admin/v2/groups/%s", groupID)

	_, body, err := c.SignedCall(http.MethodGet, path, nil, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetGroupResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Phone methods

// GetPhonesNumber sets the optional number parameter for a GetPhones request.
func GetPhonesNumber(number string) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("number", number)
	}
}

// GetPhonesExtension sets the optional extension parameter for a GetPhones request.
func GetPhonesExtension(ext string) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("extension", ext)
	}
}

// GetPhonesResult models responses containing a list of phones.
type GetPhonesResult struct {
	duoapi.StatResult
	ListResult
	Response []Phone
}

func (result *GetPhonesResult) getResponse() interface{} {
	return result.Response
}

func (result *GetPhonesResult) appendResponse(phones interface{}) {
	asserted_phones := phones.([]Phone)
	result.Response = append(result.Response, asserted_phones...)
}


// GetPhones calls GET /admin/v1/phones
// See https://duo.com/docs/adminapi#phones
func (c *Client) GetPhones(options ...func(*url.Values)) (*GetPhonesResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrievePhones(params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetPhonesResult), nil
}

func (c *Client) retrievePhones(params url.Values) (*GetPhonesResult, error) {
	_, body, err := c.SignedCall(http.MethodGet, "/admin/v1/phones", params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetPhonesResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPhoneResult models responses containing a single phone.
type GetPhoneResult struct {
	duoapi.StatResult
	Response Phone
}

// GetPhone calls GET /admin/v1/phones/:phone_id
// See https://duo.com/docs/adminapi#retrieve-phone-by-id
func (c *Client) GetPhone(phoneID string) (*GetPhoneResult, error) {
	path := fmt.Sprintf("/admin/v1/phones/%s", phoneID)

	_, body, err := c.SignedCall(http.MethodGet, path, nil, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetPhoneResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Token methods

// GetTokensTypeAndSerial sets the optional type and serial parameters for a GetTokens request.
func GetTokensTypeAndSerial(typ, serial string) func(*url.Values) {
	return func(opts *url.Values) {
		opts.Set("type", typ)
		opts.Set("serial", serial)
	}
}

// GetTokensResult models responses containing a list of tokens.
type GetTokensResult struct {
	duoapi.StatResult
	ListResult
	Response []Token
}

func (result *GetTokensResult) getResponse() interface{} {
	return result.Response
}

func (result *GetTokensResult) appendResponse(tokens interface{}) {
	asserted_tokens := tokens.([]Token)
	result.Response = append(result.Response, asserted_tokens...)
}


// GetTokens calls GET /admin/v1/tokens
// See https://duo.com/docs/adminapi#retrieve-hardware-tokens
func (c *Client) GetTokens(options ...func(*url.Values)) (*GetTokensResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveTokens(params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetTokensResult), nil
}

func (c *Client) retrieveTokens(params url.Values) (*GetTokensResult, error) {
	_, body, err := c.SignedCall(http.MethodGet, "/admin/v1/tokens", params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetTokensResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTokenResult models responses containing a single token.
type GetTokenResult struct {
	duoapi.StatResult
	Response Token
}

// GetToken calls GET /admin/v1/tokens/:token_id
// See https://duo.com/docs/adminapi#retrieve-hardware-tokens
func (c *Client) GetToken(tokenID string) (*GetTokenResult, error) {
	path := fmt.Sprintf("/admin/v1/tokens/%s", tokenID)

	_, body, err := c.SignedCall(http.MethodGet, path, nil, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetTokenResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// U2F token methods

// GetU2FTokensResult models responses containing a list of U2F tokens.
type GetU2FTokensResult struct {
	duoapi.StatResult
	ListResult
	Response []U2FToken
}

func (result *GetU2FTokensResult) getResponse() interface{} {
	return result.Response
}

func (result *GetU2FTokensResult) appendResponse(tokens interface{}) {
	asserted_tokens := tokens.([]U2FToken)
	result.Response = append(result.Response, asserted_tokens...)
}


// GetU2FTokens calls GET /admin/v1/u2ftokens
// See https://duo.com/docs/adminapi#retrieve-u2f-tokens
func (c *Client) GetU2FTokens(options ...func(*url.Values)) (*GetU2FTokensResult, error) {
	params := url.Values{}
	for _, o := range options {
		o(&params)
	}

	cb := func(params url.Values) (responsePage, error) {
		return c.retrieveU2FTokens(params)
	}
	response, err := c.retrieveItems(params, cb)
	if err != nil {
		return nil, err
	}

	return response.(*GetU2FTokensResult), nil
}

func (c *Client) retrieveU2FTokens(params url.Values) (*GetU2FTokensResult, error) {
	_, body, err := c.SignedCall(http.MethodGet, "/admin/v1/u2ftokens", params, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetU2FTokensResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetU2FToken calls GET /admin/v1/u2ftokens/:registration_id
// See https://duo.com/docs/adminapi#retrieve-u2f-token-by-id
func (c *Client) GetU2FToken(registrationID string) (*GetU2FTokensResult, error) {
	path := fmt.Sprintf("/admin/v1/u2ftokens/%s", registrationID)

	_, body, err := c.SignedCall(http.MethodGet, path, nil, duoapi.UseTimeout)
	if err != nil {
		return nil, err
	}

	result := &GetU2FTokensResult{}
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
