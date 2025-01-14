package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const SSHKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICJEoOOgQBLJPs4g/XcPTKT82NywNPpxeuA20FlOPlpO opensource@scaleway.com"

func init() {
	resource.AddTestSweepers("scaleway_iam_ssh_key", &resource.Sweeper{
		Name: "scaleway_iam_ssh_key",
		F:    testSweepIamSSHKey,
	})
}

func testSweepIamSSHKey(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		iamAPI := iam.NewAPI(scwClient)

		l.Debugf("sweeper: destroying the SSH keys")

		listSSHKeys, err := iamAPI.ListSSHKeys(&iam.ListSSHKeysRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing SSH keys in sweeper: %s", err)
		}

		for _, sshKey := range listSSHKeys.SSHKeys {
			if !isTestResource(sshKey.Name) {
				continue
			}
			err := iamAPI.DeleteSSHKey(&iam.DeleteSSHKeyRequest{
				SSHKeyID: sshKey.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting SSH key in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayIamSSHKey_basic(t *testing.T) {
	name := "tf-test-iam-ssh-key-basic"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "%2s"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s-updated"
						public_key = "%2s"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_WithNewLine(t *testing.T) {
	name := "tf-test-iam-ssh-key-newline"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_ChangeResourceName(t *testing.T) {
	name := "tf-test-iam-ssh-key-change-resource-name"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "first" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.first"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.first", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.first", "public_key", SSHKey),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "second" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.second"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.second", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.second", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayIamSSHKey_Disabled(t *testing.T) {
	name := "tf-test-iam-ssh-key-disabled"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
						disabled = "true"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%1s"
						public_key = "\n\n%2s\n\n"
						disabled = "false"
					}
				`, name, SSHKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_iam_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "public_key", SSHKey),
					resource.TestCheckResourceAttr("scaleway_iam_ssh_key.main", "disabled", "false"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamSSHKeyDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_iam_ssh_key" {
				continue
			}

			iamAPI := iamAPI(tt.Meta)

			_, err := iamAPI.GetSSHKey(&iam.GetSSHKeyRequest{
				SSHKeyID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}

func testAccCheckScalewayIamSSHKeyExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetSSHKey(&iam.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
