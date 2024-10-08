---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "binarylane_server Resource - terraform-provider-binarylane"
subcategory: ""
description: |-
  Provides a Binary Lane Server resource. This can be used to create and delete servers.
---

# binarylane_server (Resource)

Provides a Binary Lane Server resource. This can be used to create and delete servers.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `image` (String) The slug of the selected operating system.
- `region` (String) The slug of the selected region.
- `size` (String) The slug of the selected size.

### Optional

- `backups` (Boolean) If true this will enable two daily backups for the server. `options.daily_backups` will override this value if provided. Setting this to false has no effect.
- `name` (String) The hostname of your server, such as vps01.yourcompany.com. If not provided, the server will be created with a random name.
- `password` (String, Sensitive) If this is provided the specified or default remote user's account password will be set to this value. Only valid if the server supports password change actions. If omitted and the server supports password change actions a random password will be generated and emailed to the account email address.
- `port_blocking` (Boolean) Port blocking of outgoing connections for email, SSH and Remote Desktop (TCP ports 22, 25, and 3389) is enabled by default for all new servers. If this is false port blocking will be disabled. Disabling port blocking is only available to reviewed accounts.
- `public_ipv4_count` (Number) The number of public IPv4 addresses to assign to the server. If this is not provided, the server will be created with the default number of public IPv4 addresses.
- `ssh_keys` (List of Number) This is a list of SSH key ids. If this is null or not provided, any SSH keys that have been marked as default will be deployed (assuming the operating system supports SSH Keys). Submit an empty list to disable deployment of default keys.
- `user_data` (String) If provided this will be used to initialise the new server. This must be left null if the Image does not support UserData, see DistributionInfo.Features for more information.
- `vpc_id` (Number) Leave null to use default (public) network for the selected region.
- `wait_for_create` (Number) The number of seconds to wait for the server to be created, after which, a timeout error will be reported. If `wait_seconds` is left empty or set to 0, Terraform will succeed without waiting for the server creation to complete.

### Read-Only

- `id` (Number) The ID of the server to fetch.
- `permalink` (String) A randomly generated two-word identifier assigned to servers in regions that support this feature
- `private_ipv4_addresses` (List of String) The private IPv4 addresses assigned to the server.
- `public_ipv4_addresses` (List of String) The public IPv4 addresses assigned to the server.
