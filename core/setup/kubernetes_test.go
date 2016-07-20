package setup

import (
	"fmt"
	"strings"
	"testing"
)

/*
kubernetes coredns.local {
        # Use url for k8s API endpoint
        endpoint http://localhost:8080
        # Assemble k8s record names with the template
        template {service}.{namespace}.{zone}
        # Only expose the k8s namespace "demo"
        #namespaces demo
    }
*/

func TestKubernetesParse(t *testing.T) {
	tests := []struct {
		input              string
		shouldErr          bool
		expectedErrContent string // substring from the expected error. Empty for positive cases.
	}{
		// positive
		{
			`kubernetes`,
			false,
			"",
		},
		{
			`kubernetes coredns.local`,
			false,
			"",
		},
   		{
			`kubernetes coredns.local {
    endpoint http://localhost:9090
}`,
			false,
			"",
		},
   		{
			`kubernetes coredns.local {
	template {service}.{namespace}.{zone}
}`,
			false,
			"",
		},
   		{
			`kubernetes coredns.local {
	namespace demo
}`,
			false,
			"",
		},
   		{
			`kubernetes coredns.local {
	namespace demo test
}`,
			false,
			"",
		},

		// negative
   		{
			`kubernetes coredns.local {
    endpoint
}`,
			true,
			"Wrong argument count or unexpected line ending after 'endpoint'",
		},
/*
		// No template provided for template line.
   		{
			`kubernetes coredns.local {
    template
}`,
			true,
			"",
		},
*/
/*
		// No template provided for template line.
   		{
			`kubernetes coredns.local {
    namespaces
}`,
			true,
			"",
		},
*/
	}
	
	for i, test := range tests {
		c := NewTestController(test.input)
		k8sController, err := kubernetesParse(c)
		fmt.Printf("i: %v\n", i)
		fmt.Printf("err: %v\n", err)
		fmt.Printf("controller: %v\n", k8sController)
		fmt.Printf("zones: %v\n", k8sController.Zones)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: Expected error, but found one for input '%s'. Error was: '%v'", i, test.input, err)
		}
	
		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: Expected no error but found one for input %s. Error was: %v", i, test.input, err)
			}

			if test.shouldErr && (len(test.expectedErrContent) < 1) {
				t.Fatalf("Test %d: Test marked as expecting an error, but no expectedErrContent provided for input '%s'. Error was: '%v'", i, test.input, err)
			}

			if !strings.Contains(err.Error(), test.expectedErrContent) {
				t.Errorf("Test %d: Expected error to contain: %v, found error: %v, input: %s", i, test.expectedErrContent, err, test.input)
			}
		}
	}
}
