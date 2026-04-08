package provider

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bluecat": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, envVar := range []string{"BLUECAT_ENDPOINT", "BLUECAT_USERNAME", "BLUECAT_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccCheckEnvVars(t *testing.T, envVars ...string) {
	t.Helper()
	testAccPreCheck(t)
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for this test", envVar)
		}
	}
}

func validateObjectID(value string) error {
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	if valueInt <= 0 {
		return fmt.Errorf("should be a value greater than 0")
	}
	return nil
}
