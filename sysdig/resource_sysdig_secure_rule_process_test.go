//go:build tf_acc_sysdig_secure || tf_acc_policies

package sysdig_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/draios/terraform-provider-sysdig/sysdig"
)

func TestAccRuleProcess(t *testing.T) {
	t.Cleanup(func() {
		handleReport(t)
	})

	rText := func() string { return acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum) }

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			if v := os.Getenv("SYSDIG_SECURE_API_TOKEN"); v == "" {
				t.Fatal("SYSDIG_SECURE_API_TOKEN must be set for acceptance tests")
			}
		},
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"sysdig": func() (*schema.Provider, error) {
				return sysdig.Provider(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: ruleProcessWithName(rText()),
			},
			{
				Config: ruleProcessWithoutTags(rText()),
			},
			{
				ResourceName:      "sysdig_secure_rule_process.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: ruleProcessWithMinimalConfig(rText()),
			},
		},
	})
}

func ruleProcessWithName(name string) string {
	return fmt.Sprintf(`
resource "sysdig_secure_rule_process" "foo" {
  name = "TERRAFORM TEST %s"
  description = "TERRAFORM TEST %s"
  tags = ["container", "cis"]

  matching = true // default
  processes = ["bash", "sh"]
}`, name, name)
}

func ruleProcessWithoutTags(name string) string {
	return fmt.Sprintf(`
resource "sysdig_secure_rule_process" "foo" {
  name = "TERRAFORM TEST %s"
  description = "TERRAFORM TEST %s"

  matching = true // default
  processes = ["bash", "sh"]
}`, name, name)
}

func ruleProcessWithMinimalConfig(name string) string {
	return fmt.Sprintf(`
resource "sysdig_secure_rule_process" "foo-minimal" {
  name = "TERRAFORM TEST %s"
  description = "TERRAFORM TEST %s"
}`, name, name)
}
