package admin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/duosecurity/duo_api_golang"
)

func buildAdminClient(url string, proxy func(*http.Request) (*url.URL, error)) *Client {
	ikey := "eyekey"
	skey := "esskey"
	host := strings.Split(url, "//")[1]
	userAgent := "GoTestClient"
	base := duoapi.NewDuoApi(ikey, skey, host, userAgent, duoapi.SetTimeout(1*time.Second), duoapi.SetInsecure(), duoapi.SetProxy(proxy))
	return New(*base)
}

func getBodyParams(r *http.Request) (url.Values, error) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return url.Values{}, err
	}
	reqParams, err := url.ParseQuery(string(body))
	return reqParams, err
}

const getUsersResponse = `{
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 1
	},
	"response": [{
		"alias1": "joe.smith",
		"alias2": "jsmith@example.com",
		"alias3": null,
		"alias4": null,
		"created": 1489612729,
		"email": "jsmith@example.com",
		"firstname": "Joe",
		"groups": [{
			"desc": "People with hardware tokens",
			"name": "token_users"
		}],
		"last_directory_sync": 1508789163,
		"last_login": 1343921403,
		"lastname": "Smith",
		"notes": "",
		"phones": [{
			"phone_id": "DPFZRS9FB0D46QFTM899",
			"number": "+15555550100",
			"extension": "",
			"name": "",
			"postdelay": null,
			"predelay": null,
			"type": "Mobile",
			"capabilities": [
				"sms",
				"phone",
				"push"
			],
			"platform": "Apple iOS",
			"activated": false,
			"sms_passcodes_sent": false
		}],
		"realname": "Joe Smith",
		"status": "active",
		"tokens": [{
			"serial": "0",
			"token_id": "DHIZ34ALBA2445ND4AI2",
			"type": "d1"
		}],
		"user_id": "DU3RP9I2WOC59VZX672N",
		"username": "jsmith"
	}]
}`

func TestGetUsers(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getUsersResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUsers()
	if err != nil {
		t.Errorf("Unexpected error from GetUsers call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 user, but got %d", len(result.Response))
	}
	if result.Response[0].UserID != "DU3RP9I2WOC59VZX672N" {
		t.Errorf("Expected user ID DU3RP9I2WOC59VZX672N, but got %s", result.Response[0].UserID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUsersPage1Response = `{
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": 1,
		"total_objects": 2
	},
	"response": [{
		"alias1": "joe.smith",
		"alias2": "jsmith@example.com",
		"alias3": null,
		"alias4": null,
		"created": 1489612729,
		"email": "jsmith@example.com",
		"firstname": "Joe",
		"groups": [{
			"desc": "People with hardware tokens",
			"name": "token_users"
		}],
		"last_directory_sync": 1508789163,
		"last_login": 1343921403,
		"lastname": "Smith",
		"notes": "",
		"phones": [{
			"phone_id": "DPFZRS9FB0D46QFTM899",
			"number": "+15555550100",
			"extension": "",
			"name": "",
			"postdelay": null,
			"predelay": null,
			"type": "Mobile",
			"capabilities": [
				"sms",
				"phone",
				"push"
			],
			"platform": "Apple iOS",
			"activated": false,
			"sms_passcodes_sent": false
		}],
		"realname": "Joe Smith",
		"status": "active",
		"tokens": [{
			"serial": "0",
			"token_id": "DHIZ34ALBA2445ND4AI2",
			"type": "d1"
		}],
		"user_id": "DU3RP9I2WOC59VZX672N",
		"username": "jsmith"
	}]
}`

const getUsersPage2Response = `{
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 2
	},
	"response": [{
		"alias1": "joe.smith",
		"alias2": "jsmith@example.com",
		"alias3": null,
		"alias4": null,
		"created": 1489612729,
		"email": "jsmith@example.com",
		"firstname": "Joe",
		"groups": [{
			"desc": "People with hardware tokens",
			"name": "token_users"
		}],
		"last_directory_sync": 1508789163,
		"last_login": 1343921403,
		"lastname": "Smith",
		"notes": "",
		"phones": [{
			"phone_id": "DPFZRS9FB0D46QFTM899",
			"number": "+15555550100",
			"extension": "",
			"name": "",
			"postdelay": null,
			"predelay": null,
			"type": "Mobile",
			"capabilities": [
				"sms",
				"phone",
				"push"
			],
			"platform": "Apple iOS",
			"activated": false,
			"sms_passcodes_sent": false
		}],
		"realname": "Joe Smith",
		"status": "active",
		"tokens": [{
			"serial": "0",
			"token_id": "DHIZ34ALBA2445ND4AI2",
			"type": "d1"
		}],
		"user_id": "DU3RP9I2WOC59VZX672N",
		"username": "jsmith"
	}]
}`

func TestGetUsersMultipage(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getUsersPage1Response)
			} else {
				fmt.Fprintln(w, getUsersPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUsers()

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if result.Metadata.TotalObjects != "2" {
		t.Errorf("Expected total obects to be two, found %s", result.Metadata.TotalObjects)
	}

	if len(result.Response) != 2 {
		t.Errorf("Expected two users in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

const getEmptyPageArgsResponse = `{
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": 2,
		"total_objects": 2
	},
	"response": []
}`

func TestGetUserPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetUsers(func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

func TestGetUser(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getUsersResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUser("DU3RP9I2WOC59VZX672N")
	if err != nil {
		t.Errorf("Unexpected error from GetUser call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 user, but got %d", len(result.Response))
	}
	if result.Response[0].UserID != "DU3RP9I2WOC59VZX672N" {
		t.Errorf("Expected user ID DU3RP9I2WOC59VZX672N, but got %s", result.Response[0].UserID)
	}
}

const getGroupsResponse = `{
	"response": [{
		"desc": "This is group A",
		"group_id": "DGXXXXXXXXXXXXXXXXXA",
		"name": "Group A",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	},
	{
		"desc": "This is group B",
		"group_id": "DGXXXXXXXXXXXXXXXXXB",
		"name": "Group B",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	}],
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetUserGroups(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getGroupsResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserGroups("DU3RP9I2WOC59VZX672N")
	if err != nil {
		t.Errorf("Unexpected error from GetUserGroups call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 2 {
		t.Errorf("Expected 2 groups, but got %d", len(result.Response))
	}
	if result.Response[0].GroupID != "DGXXXXXXXXXXXXXXXXXA" {
		t.Errorf("Expected group ID DGXXXXXXXXXXXXXXXXXA, but got %s", result.Response[0].GroupID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getGroupsPage1Response = `{
	"response": [{
		"desc": "This is group A",
		"group_id": "DGXXXXXXXXXXXXXXXXXA",
		"name": "Group A",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	},
	{
		"desc": "This is group B",
		"group_id": "DGXXXXXXXXXXXXXXXXXB",
		"name": "Group B",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	}],
	"stat": "OK",
	"metadata": {
		"prev_offset": null,
		"next_offset": 2,
		"total_objects": 4
	}
}`

const getGroupsPage2Response = `{
	"response": [{
		"desc": "This is group C",
		"group_id": "DGXXXXXXXXXXXXXXXXXC",
		"name": "Group C",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	},
	{
		"desc": "This is group D",
		"group_id": "DGXXXXXXXXXXXXXXXXXD",
		"name": "Group D",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	}],
	"stat": "OK",
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 4
	}
}`

func TestGetUserGroupsMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getGroupsPage1Response)
			} else {
				fmt.Fprintln(w, getGroupsPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserGroups("DU3RP9I2WOC59VZX672N")

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 4 {
		t.Errorf("Expected four groups in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetUserGroupsPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetUserGroups("DU3RP9I2WOC59VZX672N", func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUserPhonesResponse = `{
	"stat": "OK",
	"response": [{
		"activated": false,
		"capabilities": [
			"sms",
			"phone",
			"push"
		],
		"extension": "",
		"name": "",
		"number": "+15035550102",
		"phone_id": "DPFZRS9FB0D46QFTM890",
		"platform": "Apple iOS",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Mobile"
	},
	{
		"activated": false,
		"capabilities": [
			"phone"
		],
		"extension": "",
		"name": "",
		"number": "+15035550103",
		"phone_id": "DPFZRS9FB0D46QFTM891",
		"platform": "Unknown",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Landline"
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetUserPhones(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getUserPhonesResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserPhones("DU3RP9I2WOC59VZX672N")
	if err != nil {
		t.Errorf("Unexpected error from GetUserPhones call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 2 {
		t.Errorf("Expected 2 phones, but got %d", len(result.Response))
	}
	if result.Response[0].PhoneID != "DPFZRS9FB0D46QFTM890" {
		t.Errorf("Expected phone ID DPFZRS9FB0D46QFTM890, but got %s", result.Response[0].PhoneID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUserPhonesPage1Response = `{
	"stat": "OK",
	"response": [{
		"activated": false,
		"capabilities": [
			"sms",
			"phone",
			"push"
		],
		"extension": "",
		"name": "",
		"number": "+15035550102",
		"phone_id": "DPFZRS9FB0D46QFTM890",
		"platform": "Apple iOS",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Mobile"
	},
	{
		"activated": false,
		"capabilities": [
			"phone"
		],
		"extension": "",
		"name": "",
		"number": "+15035550103",
		"phone_id": "DPFZRS9FB0D46QFTM891",
		"platform": "Unknown",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Landline"
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 2,
		"total_objects": 4
	}
}`

const getUserPhonesPage2Response = `{
	"stat": "OK",
	"response": [{
		"activated": false,
		"capabilities": [
			"sms",
			"phone",
			"push"
		],
		"extension": "",
		"name": "",
		"number": "+15035550102",
		"phone_id": "DPFZRS9FB0D46QFTM890",
		"platform": "Apple iOS",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Mobile"
	},
	{
		"activated": false,
		"capabilities": [
			"phone"
		],
		"extension": "",
		"name": "",
		"number": "+15035550103",
		"phone_id": "DPFZRS9FB0D46QFTM891",
		"platform": "Unknown",
		"postdelay": null,
		"predelay": null,
		"sms_passcodes_sent": false,
		"type": "Landline"
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 4
	}
}`

func TestGetUserPhonesMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getUserPhonesPage1Response)
			} else {
				fmt.Fprintln(w, getUserPhonesPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserPhones("DU3RP9I2WOC59VZX672N")

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 4 {
		t.Errorf("Expected four phones in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetUserPhonesPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetUserPhones("DU3RP9I2WOC59VZX672N", func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUserTokensResponse = `{
	"stat": "OK",
	"response": [{
		"type": "d1",
		"serial": "0",
		"token_id": "DHEKH0JJIYC1LX3AZWO4"
	},
	{
		"type": "d1",
		"serial": "7",
		"token_id": "DHUNT3ZVS3ACF8AEV2WG",
		"totp_step": null
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetUserTokens(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getUserTokensResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserTokens("DU3RP9I2WOC59VZX672N")
	if err != nil {
		t.Errorf("Unexpected error from GetUserTokens call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 2 {
		t.Errorf("Expected 2 tokens, but got %d", len(result.Response))
	}
	if result.Response[0].TokenID != "DHEKH0JJIYC1LX3AZWO4" {
		t.Errorf("Expected token ID DHEKH0JJIYC1LX3AZWO4, but got %s", result.Response[0].TokenID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUserTokensPage1Response = `{
	"stat": "OK",
	"response": [{
		"type": "d1",
		"serial": "0",
		"token_id": "DHEKH0JJIYC1LX3AZWO4"
	},
	{
		"type": "d1",
		"serial": "7",
		"token_id": "DHUNT3ZVS3ACF8AEV2WG",
		"totp_step": null
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 2,
		"total_objects": 4
	}
}`

const getUserTokensPage2Response = `{
	"stat": "OK",
	"response": [{
		"type": "d1",
		"serial": "0",
		"token_id": "DHEKH0JJIYC1LX3AZWO4"
	},
	{
		"type": "d1",
		"serial": "7",
		"token_id": "DHUNT3ZVS3ACF8AEV2WG",
		"totp_step": null
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 4
	}
}`

func TestGetUserTokensMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getUserTokensPage1Response)
			} else {
				fmt.Fprintln(w, getUserTokensPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserTokens("DU3RP9I2WOC59VZX672N")

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 4 {
		t.Errorf("Expected four tokens in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetUserTokensPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetUserTokens("DU3RP9I2WOC59VZX672N", func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const associateUserTokenResponse = `{
	"stat": "OK",
	"response": ""
}`

func TestAssociateUserToken(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, associateUserTokenResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.AssociateUserToken("DU3RP9I2WOC59VZX672N", "DHEKH0JJIYC1LX3AZWO4")
	if err != nil {
		t.Errorf("Unexpected error from AssociateUserToken call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 0 {
		t.Errorf("Expected empty response, but got %s", result.Response)
	}
}

const getUserU2FTokensResponse = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV"
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 1
	}
}`

func TestGetUserU2FTokens(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getUserU2FTokensResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserU2FTokens("DU3RP9I2WOC59VZX672N")
	if err != nil {
		t.Errorf("Unexpected error from GetUserU2FTokens call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 token, but got %d", len(result.Response))
	}
	if result.Response[0].RegistrationID != "D21RU6X1B1DF5P54B6PV" {
		t.Errorf("Expected registration ID D21RU6X1B1DF5P54B6PV, but got %s", result.Response[0].RegistrationID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getUserU2FTokensPage1Response = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV"
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 1,
		"total_objects": 2
	}
}`

const getUserU2FTokensPage2Response = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV"
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetUserU2FTokensMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getUserU2FTokensPage1Response)
			} else {
				fmt.Fprintln(w, getUserU2FTokensPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetUserU2FTokens("DU3RP9I2WOC59VZX672N")

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 2 {
		t.Errorf("Expected two tokens in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetUserU2FTokensPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetUserU2FTokens("DU3RP9I2WOC59VZX672N", func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

func TestGetGroups(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getGroupsResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetGroups()
	if err != nil {
		t.Errorf("Unexpected error from GetGroups call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 2 {
		t.Errorf("Expected 2 groups, but got %d", len(result.Response))
	}
	if result.Response[0].Name != "Group A" {
		t.Errorf("Expected group name Group A, but got %s", result.Response[0].Name)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

func TestGetGroupsMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getGroupsPage1Response)
			} else {
				fmt.Fprintln(w, getGroupsPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetGroups()

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 4 {
		t.Errorf("Expected four groups in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetGroupsPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetGroups(func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getGroupResponse = `{
	"response": {
		"desc": "Group description",
		"group_id": "DGXXXXXXXXXXXXXXXXXX",
		"name": "Group Name",
		"push_enabled": true,
		"sms_enabled": true,
		"status": "active",
		"voice_enabled": true,
		"mobile_otp_enabled": true
	},
	"stat": "OK"
}`

func TestGetGroup(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getGroupResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetGroup("DGXXXXXXXXXXXXXXXXXX")
	if err != nil {
		t.Errorf("Unexpected error from GetGroups call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if result.Response.GroupID != "DGXXXXXXXXXXXXXXXXXX" {
		t.Errorf("Expected group ID DGXXXXXXXXXXXXXXXXXX, but got %s", result.Response.GroupID)
	}
	if !result.Response.PushEnabled {
		t.Errorf("Expected push to be enabled, but got %v", result.Response.PushEnabled)
	}
}

const getPhonesResponse = `{
	"stat": "OK",
	"response": [{
		"activated": true,
		"capabilities": [
			"push",
			"sms",
			"phone",
			"mobile_otp"
		],
		"encrypted": "Encrypted",
		"extension": "",
		"fingerprint": "Configured",
		"name": "",
		"number": "+15555550100",
		"phone_id": "DPFZRS9FB0D46QFTM899",
		"platform": "Google Android",
		"postdelay": "",
		"predelay": "",
		"screenlock": "Locked",
		"sms_passcodes_sent": false,
		"tampered": "Not tampered",
		"type": "Mobile",
		"users": [{
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_login": 1474399627,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith"
		}]
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 1
	}
}`

func TestGetPhones(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getPhonesResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetPhones()
	if err != nil {
		t.Errorf("Unexpected error from GetPhones call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 phone, but got %d", len(result.Response))
	}
	if result.Response[0].PhoneID != "DPFZRS9FB0D46QFTM899" {
		t.Errorf("Expected phone ID DPFZRS9FB0D46QFTM899, but got %s", result.Response[0].PhoneID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getPhonesPage1Response = `{
	"stat": "OK",
	"response": [{
		"activated": true,
		"capabilities": [
			"push",
			"sms",
			"phone",
			"mobile_otp"
		],
		"encrypted": "Encrypted",
		"extension": "",
		"fingerprint": "Configured",
		"name": "",
		"number": "+15555550100",
		"phone_id": "DPFZRS9FB0D46QFTM899",
		"platform": "Google Android",
		"postdelay": "",
		"predelay": "",
		"screenlock": "Locked",
		"sms_passcodes_sent": false,
		"tampered": "Not tampered",
		"type": "Mobile",
		"users": [{
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_login": 1474399627,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith"
		}]
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 1,
		"total_objects": 2
	}
}`

const getPhonesPage2Response = `{
	"stat": "OK",
	"response": [{
		"activated": true,
		"capabilities": [
			"push",
			"sms",
			"phone",
			"mobile_otp"
		],
		"encrypted": "Encrypted",
		"extension": "",
		"fingerprint": "Configured",
		"name": "",
		"number": "+15555550100",
		"phone_id": "DPFZRS9FB0D46QFTM899",
		"platform": "Google Android",
		"postdelay": "",
		"predelay": "",
		"screenlock": "Locked",
		"sms_passcodes_sent": false,
		"tampered": "Not tampered",
		"type": "Mobile",
		"users": [{
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_login": 1474399627,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith"
		}]
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetPhonesMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getPhonesPage1Response)
			} else {
				fmt.Fprintln(w, getPhonesPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetPhones()

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 2 {
		t.Errorf("Expected two phones in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetPhonesPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetPhones(func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getPhoneResponse = `{
	"stat": "OK",
	"response": {
		"phone_id": "DPFZRS9FB0D46QFTM899",
		"number": "+15555550100",
		"name": "",
		"extension": "",
		"postdelay": null,
		"predelay": null,
		"type": "Mobile",
		"capabilities": [
			"sms",
			"phone",
			"push"
		],
		"platform": "Apple iOS",
		"activated": false,
		"sms_passcodes_sent": false,
		"users": [{
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith",
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"realname": "Joe Smith",
			"email": "jsmith@example.com",
			"status": "active",
			"last_login": 1343921403,
			"notes": ""
		}]
	}
}`

func TestGetPhone(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getPhoneResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetPhone("DPFZRS9FB0D46QFTM899")
	if err != nil {
		t.Errorf("Unexpected error from GetPhone call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if result.Response.PhoneID != "DPFZRS9FB0D46QFTM899" {
		t.Errorf("Expected phone ID DPFZRS9FB0D46QFTM899, but got %s", result.Response.PhoneID)
	}
}

const getTokensResponse = `{
	"stat": "OK",
	"response": [{
		"serial": "0",
		"token_id": "DHIZ34ALBA2445ND4AI2",
		"type": "d1",
		"totp_step": null,
		"users": [{
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith",
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"realname": "Joe Smith",
			"email": "jsmith@example.com",
			"status": "active",
			"last_login": 1343921403,
			"notes": ""
		}]
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 1
	}
}`

func TestGetTokens(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getTokensResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetTokens()
	if err != nil {
		t.Errorf("Unexpected error from GetTokens call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 token, but got %d", len(result.Response))
	}
	if result.Response[0].TokenID != "DHIZ34ALBA2445ND4AI2" {
		t.Errorf("Expected token ID DHIZ34ALBA2445ND4AI2, but got %s", result.Response[0].TokenID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getTokensPage1Response = `{
	"stat": "OK",
	"response": [{
		"serial": "0",
		"token_id": "DHIZ34ALBA2445ND4AI2",
		"type": "d1",
		"totp_step": null,
		"users": [{
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith",
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"realname": "Joe Smith",
			"email": "jsmith@example.com",
			"status": "active",
			"last_login": 1343921403,
			"notes": ""
		}]
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 1,
		"total_objects": 2
	}
}`

const getTokensPage2Response = `{
	"stat": "OK",
	"response": [{
		"serial": "0",
		"token_id": "DHIZ34ALBA2445ND4AI2",
		"type": "d1",
		"totp_step": null,
		"users": [{
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith",
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"realname": "Joe Smith",
			"email": "jsmith@example.com",
			"status": "active",
			"last_login": 1343921403,
			"notes": ""
		}]
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetTokensMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getTokensPage1Response)
			} else {
				fmt.Fprintln(w, getTokensPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetTokens()

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 2 {
		t.Errorf("Expected two tokens in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetTokensPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetTokens(func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getTokenResponse = `{
	"stat": "OK",
	"response": {
		"serial": "0",
		"token_id": "DHIZ34ALBA2445ND4AI2",
		"type": "d1",
		"totp_step": null,
		"users": [{
			"user_id": "DUJZ2U4L80HT45MQ4EOQ",
			"username": "jsmith",
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"realname": "Joe Smith",
			"email": "jsmith@example.com",
			"status": "active",
			"last_login": 1343921403,
			"notes": ""
		}]
	}
}`

func TestGetToken(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getTokenResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetToken("DPFZRS9FB0D46QFTM899")
	if err != nil {
		t.Errorf("Unexpected error from GetToken call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if result.Response.TokenID != "DHIZ34ALBA2445ND4AI2" {
		t.Errorf("Expected token ID DHIZ34ALBA2445ND4AI2, but got %s", result.Response.TokenID)
	}
}

const getU2FTokensResponse = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV",
		"user": {
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"created": 1384275337,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_directory_sync": 1384275337,
			"last_login": 1514922986,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DU3RP9I2WOC59VZX672N",
			"username": "jsmith"
		}
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": null,
		"total_objects": 1
	}
}`

func TestGetU2FTokens(t *testing.T) {
	var last_request *http.Request
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getU2FTokensResponse)
			last_request = r
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetU2FTokens()
	if err != nil {
		t.Errorf("Unexpected error from GetU2FTokens call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 token, but got %d", len(result.Response))
	}
	if result.Response[0].RegistrationID != "D21RU6X1B1DF5P54B6PV" {
		t.Errorf("Expected registration ID D21RU6X1B1DF5P54B6PV, but got %s", result.Response[0].RegistrationID)
	}

	request_query := last_request.URL.Query()
	if request_query["limit"][0] != "100" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "0" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

const getU2FTokensPage1Response = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV",
		"user": {
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"created": 1384275337,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_directory_sync": 1384275337,
			"last_login": 1514922986,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DU3RP9I2WOC59VZX672N",
			"username": "jsmith"
		}
	}],
	"metadata": {
		"prev_offset": null,
		"next_offset": 1,
		"total_objects": 2
	}
}`

const getU2FTokensPage2Response = `{
	"stat": "OK",
	"response": [{
		"date_added": 1444678994,
		"registration_id": "D21RU6X1B1DF5P54B6PV",
		"user": {
			"alias1": "joe.smith",
			"alias2": "jsmith@example.com",
			"alias3": null,
			"alias4": null,
			"created": 1384275337,
			"email": "jsmith@example.com",
			"firstname": "Joe",
			"last_directory_sync": 1384275337,
			"last_login": 1514922986,
			"lastname": "Smith",
			"notes": "",
			"realname": "Joe Smith",
			"status": "active",
			"user_id": "DU3RP9I2WOC59VZX672N",
			"username": "jsmith"
		}
	}],
	"metadata": {
		"prev_offset": 0,
		"next_offset": null,
		"total_objects": 2
	}
}`

func TestGetU2fTokensMultiple(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(requests) == 0 {
				fmt.Fprintln(w, getU2FTokensPage1Response)
			} else {
				fmt.Fprintln(w, getU2FTokensPage2Response)
			}
			requests = append(requests, r)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetU2FTokens()

	if len(requests) != 2 {
		t.Errorf("Expected two requets, found %d", len(requests))
	}

	if len(result.Response) != 2 {
		t.Errorf("Expected two tokens in the response, found %d", len(result.Response))
	}

	if err != nil {
		t.Errorf("Expected err to be nil, found %s", err)
	}
}

func TestGetU2FTokensPageArgs(t *testing.T) {
	requests := []*http.Request{}
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getEmptyPageArgsResponse)
			requests = append(requests, r)
		}),
	)

	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	_, err := duo.GetU2FTokens(func(values *url.Values){
		values.Set("limit", "200")
		values.Set("offset", "1")
		return
	})

	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(requests) != 1 {
		t.Errorf("Expected there to be one request, found %d", len(requests))
	}
	request := requests[0]
	request_query := request.URL.Query()
	if request_query["limit"][0] != "200" {
		t.Errorf("Expected to see a limit of 100 in request, bug got %s", request_query["limit"])
	}
	if request_query["offset"][0] != "1" {
		t.Errorf("Expected to see an offset of 0 in request, bug got %s", request_query["offset"])
	}
}

func TestGetU2FToken(t *testing.T) {
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, getU2FTokensResponse)
		}),
	)
	defer ts.Close()

	duo := buildAdminClient(ts.URL, nil)

	result, err := duo.GetU2FToken("D21RU6X1B1DF5P54B6PV")
	if err != nil {
		t.Errorf("Unexpected error from GetU2FToken call %v", err.Error())
	}
	if result.Stat != "OK" {
		t.Errorf("Expected OK, but got %s", result.Stat)
	}
	if len(result.Response) != 1 {
		t.Errorf("Expected 1 token, but got %d", len(result.Response))
	}
	if result.Response[0].RegistrationID != "D21RU6X1B1DF5P54B6PV" {
		t.Errorf("Expected registration ID D21RU6X1B1DF5P54B6PV, but got %s", result.Response[0].RegistrationID)
	}
}
