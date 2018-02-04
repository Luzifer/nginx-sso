package crowd

// Error represents a error response from Crowd.
// Error reasons are documented at https://developer.atlassian.com/display/CROWDDEV/Using+the+Crowd+REST+APIs#UsingtheCrowdRESTAPIs-HTTPResponseCodesandErrorResponses
type Error struct {
	XMLName struct{} `xml:"error"`
	Reason  string   `xml:"reason"`
	Message string   `xml:"message"`
}
