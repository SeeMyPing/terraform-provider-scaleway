package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_iam_group", &resource.Sweeper{
		Name: "scaleway_iam_group",
		F:    testSweepIamGroup,
	})
}

func testSweepIamGroup(_ string) error {
	return sweep(func(scwClient *scw.Client) error {
		api := iam.NewAPI(scwClient)

		orgID, exists := scwClient.GetDefaultOrganizationID()
		if !exists {
			return fmt.Errorf("missing organizationID")
		}

		listApps, err := api.ListGroups(&iam.ListGroupsRequest{
			OrganizationID: orgID,
		})
		if err != nil {
			return fmt.Errorf("failed to list groups: %w", err)
		}
		for _, group := range listApps.Groups {
			if !isTestResource(group.Name) {
				continue
			}
			err = api.DeleteGroup(&iam.DeleteGroupRequest{
				GroupID: group.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete group: %w", err)
			}
		}
		return nil
	})
}

func TestAccScalewayIamGroup_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamGroupDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
						resource "scaleway_iam_group" "main_basic" {
							name = "tf_tests_iam_group_basic"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "name", "tf_tests_iam_group_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "description", ""),
				),
			},
			{
				Config: `
						resource "scaleway_iam_group" "main_basic" {
							name = "tf_tests_iam_group_basic"
							description = "basic description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "name", "tf_tests_iam_group_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "description", "basic description"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_group" "main_basic" {
							name = "tf_tests_iam_group_basic_renamed"
							description = "basic description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "name", "tf_tests_iam_group_basic_renamed"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "description", "basic description"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_group" "main_basic" {
							name = "tf_tests_iam_group_basic_renamed"
							description = "this is another description"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_basic"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "name", "tf_tests_iam_group_basic_renamed"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_basic", "description", "this is another description"),
				),
			},
			{
				Config: `
						resource "scaleway_iam_group" "main" {
							name = "tf_tests_iam_group_basic_renamed"
						}
					`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main", "name", "tf_tests_iam_group_basic_renamed"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main", "description", ""),
				),
			},
		},
	})
}

func TestAccScalewayIamGroup_Applications(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamGroupDestroy(tt),
			testAccCheckScalewayIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "app01" {
						name = "tf_tests_iam_group_app"
					}
					resource "scaleway_iam_group" "main_app" {
						name = "tf_tests_iam_group_app"
						application_ids = [
							scaleway_iam_application.app01.id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "name", "tf_tests_iam_group_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "application_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_app", "application_ids.0", "scaleway_iam_application.app01", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app01" {
						name = "tf_tests_iam_group_app"
					}
					resource "scaleway_iam_application" "app02" {
						name = "tf_tests_iam_group_app2"
					}
					resource "scaleway_iam_group" "main_app" {
						name = "tf_tests_iam_group_app"
						application_ids = [
							scaleway_iam_application.app01.id,
							scaleway_iam_application.app02.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "name", "tf_tests_iam_group_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "application_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_app", "application_ids.*", "scaleway_iam_application.app01", "id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_app", "application_ids.*", "scaleway_iam_application.app02", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app01" {
						name = "tf_tests_iam_group_app"
					}
					resource "scaleway_iam_application" "app02" {
						name = "tf_tests_iam_group_app2"
					}
					resource "scaleway_iam_group" "main_app" {
						name = "tf_tests_iam_group_app"
						application_ids = [
							scaleway_iam_application.app02.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "name", "tf_tests_iam_group_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "application_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_app", "application_ids.0", "scaleway_iam_application.app02", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app01" {
						name = "tf_tests_iam_group_app"
					}
					resource "scaleway_iam_application" "app02" {
						name = "tf_tests_iam_group_app2"
					}
					resource "scaleway_iam_group" "main_app" {
						name = "tf_tests_iam_group_app"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "name", "tf_tests_iam_group_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_app", "application_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_app", "application_ids.0"),
				),
			},
		},
	})
}

func TestAccScalewayIamGroup_Users(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamGroupDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_iam_user" "user00" {
						user_id = "84d20ae1-9650-419a-ab74-7ab09b6262e0"
					}

					resource "scaleway_iam_group" "main_user" {
						name = "tf_tests_iam_group_user"
						user_ids = [
							data.scaleway_iam_user.user00.user_id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "name", "tf_tests_iam_group_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "user_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_user", "user_ids.0", "data.scaleway_iam_user.user00", "user_id"),
				),
			},
			{
				Config: `
					data "scaleway_iam_user" "user00" {
						user_id = "84d20ae1-9650-419a-ab74-7ab09b6262e0"
					}
					data "scaleway_iam_user" "user01" {
						user_id = "ef29ce05-3f2b-4fa0-a259-d76110850d57"
					}

					resource "scaleway_iam_group" "main_user" {
						name = "tf_tests_iam_group_user"
						user_ids = [
							data.scaleway_iam_user.user00.user_id,
							data.scaleway_iam_user.user01.user_id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "name", "tf_tests_iam_group_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "user_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_user", "user_ids.*", "data.scaleway_iam_user.user00", "user_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_user", "user_ids.*", "data.scaleway_iam_user.user01", "user_id"),
				),
			},
			{
				Config: `
					data "scaleway_iam_user" "user02" {
						user_id = "b6360d4f-831c-45a8-889e-0b65ed079e63"
					}

					resource "scaleway_iam_group" "main_user" {
						name = "tf_tests_iam_group_user"
						user_ids = [
							data.scaleway_iam_user.user02.user_id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "name", "tf_tests_iam_group_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "user_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_user", "user_ids.0", "data.scaleway_iam_user.user02", "user_id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_group" "main_user" {
						name = "tf_tests_iam_group_user"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "name", "tf_tests_iam_group_user"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_user", "user_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_user", "user_ids.0"),
				),
			},
		},
	})
}

func TestAccScalewayIamGroup_UsersAndApplications(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayIamGroupDestroy(tt),
			testAccCheckScalewayIamApplicationDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_iam_application" "app03" {
						name = "tf_tests_iam_group_app3"
					}

					data "scaleway_iam_user" "user00" {
						user_id = "84d20ae1-9650-419a-ab74-7ab09b6262e0"
					}

					resource "scaleway_iam_group" "main_mix" {
						name = "tf_tests_iam_group_user_app"
						application_ids = [
							scaleway_iam_application.app03.id
						]
						user_ids = [
							data.scaleway_iam_user.user00.user_id
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_mix"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "name", "tf_tests_iam_group_user_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "user_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_mix", "user_ids.0", "data.scaleway_iam_user.user00", "user_id"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "application_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_mix", "application_ids.0", "scaleway_iam_application.app03", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app03" {
						name = "tf_tests_iam_group_app3"
					}
					resource "scaleway_iam_application" "app04" {
						name = "tf_tests_iam_group_app4"
					}

					resource "scaleway_iam_group" "main_mix" {
						name = "tf_tests_iam_group_user_app"
						application_ids = [
							scaleway_iam_application.app03.id,
							scaleway_iam_application.app04.id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_mix"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "name", "tf_tests_iam_group_user_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "application_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "application_ids.*", "scaleway_iam_application.app03", "id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "application_ids.*", "scaleway_iam_application.app04", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "user_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_mix", "user_ids.0"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app03" {
						name = "tf_tests_iam_group_app3"
					}
					resource "scaleway_iam_application" "app04" {
						name = "tf_tests_iam_group_app4"
					}

					data "scaleway_iam_user" "user00" {
						user_id = "ef29ce05-3f2b-4fa0-a259-d76110850d57"
					}
					data "scaleway_iam_user" "user01" {
						user_id = "b6360d4f-831c-45a8-889e-0b65ed079e63"
					}

					resource "scaleway_iam_group" "main_mix" {
						name = "tf_tests_iam_group_user_app"
						application_ids = [
							scaleway_iam_application.app04.id,
						]
						user_ids = [
							data.scaleway_iam_user.user00.user_id,
							data.scaleway_iam_user.user01.user_id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_mix"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "name", "tf_tests_iam_group_user_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "application_ids.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_iam_group.main_mix", "application_ids.0", "scaleway_iam_application.app04", "id"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "user_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "user_ids.*", "data.scaleway_iam_user.user00", "user_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "user_ids.*", "data.scaleway_iam_user.user01", "user_id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app03" {
						name = "tf_tests_iam_group_app3"
					}
					resource "scaleway_iam_application" "app04" {
						name = "tf_tests_iam_group_app4"
					}

					data "scaleway_iam_user" "user01" {
						user_id = "ef29ce05-3f2b-4fa0-a259-d76110850d57"
					}
					data "scaleway_iam_user" "user03" {
						user_id = "b6360d4f-831c-45a8-889e-0b65ed079e63"
					}
					data "scaleway_iam_user" "user04" {
						user_id = "84d20ae1-9650-419a-ab74-7ab09b6262e0"
					}

					resource "scaleway_iam_group" "main_mix" {
						name = "tf_tests_iam_group_user_app"
						user_ids = [
							data.scaleway_iam_user.user03.user_id,
							data.scaleway_iam_user.user01.user_id,
							data.scaleway_iam_user.user04.user_id,
						]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_mix"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "name", "tf_tests_iam_group_user_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "application_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_mix", "application_ids.0"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "user_ids.#", "3"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "user_ids.*", "data.scaleway_iam_user.user01", "user_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "user_ids.*", "data.scaleway_iam_user.user03", "user_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_iam_group.main_mix", "user_ids.*", "data.scaleway_iam_user.user04", "user_id"),
				),
			},
			{
				Config: `
					resource "scaleway_iam_application" "app03" {
						name = "tf_tests_iam_group_app3"
					}
					resource "scaleway_iam_application" "app04" {
						name = "tf_tests_iam_group_app4"
					}

					resource "scaleway_iam_group" "main_mix" {
						name = "tf_tests_iam_group_user_app"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamGroupExists(tt, "scaleway_iam_group.main_mix"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "name", "tf_tests_iam_group_user_app"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "application_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_mix", "application_ids.0"),
					resource.TestCheckResourceAttr("scaleway_iam_group.main_mix", "user_ids.#", "0"),
					resource.TestCheckNoResourceAttr("scaleway_iam_group.main_mix", "user_ids.0"),
				),
			},
		},
	})
}

func testAccCheckScalewayIamGroupExists(tt *TestTools, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		iamAPI := iamAPI(tt.Meta)

		_, err := iamAPI.GetGroup(&iam.GetGroupRequest{
			GroupID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("could not find group: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayIamGroupDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_iam_group" {
				continue
			}

			iamAPI := iamAPI(tt.Meta)

			_, err := iamAPI.GetGroup(&iam.GetGroupRequest{
				GroupID: rs.Primary.ID,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("resource %s(%s) still exist", rs.Type, rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
