package yaml

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestToJSON(t *testing.T) {
	src := `Services:
-   Orders:
    -   ID: $save ID1
        SupplierOrderCode: $SupplierOrderCode
    -   ID: $save ID2
        SupplierOrderCode: 111111
`

	expected := `{"Services":[{"Orders":[{"ID":"$save ID1","SupplierOrderCode":"$SupplierOrderCode"},{"ID":"$save ID2","SupplierOrderCode":111111}]}]}
`

	j, err := ToJSON(strings.NewReader(src))
	if err != nil {
		t.Fatalf("ToJSON failed: %s", err)
	}

	real, _ := ioutil.ReadAll(j)
	realString := string(real)

	if expected != realString {
		t.Errorf("Expected JSON was not created:\nEXPE: %q\nREAL: %q", expected, realString)
	}
}
