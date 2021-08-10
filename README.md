<p align="center">
  <img src="tf.png" alt="Terraform" width="180"/>

  <h3 align="center">Terraform Workspace State CLI</h3>

  <p align="center">
    <a href="https://github.com/mupuri/go-tfdr/actions?query=workflow%3Abuild"><img alt="Build" src="https://github.com/mupuri/go-tfdr/workflows/build/badge.svg"></a>
    <a href="https://github.com/mupuri/go-tfdr/actions?query=workflow%3Atest"><img alt="Test" src="https://github.com/mupuri/go-tfdr/workflows/test/badge.svg"></a>
    <a href="https://github.com/mupuri/go-tfdr/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/tyler-technologies/go-tfdr"></a>
    <a href="https://github.com/mupuri/go-tfdr/releases/latest"><img alt="Downloads" src="https://img.shields.io/github/downloads/tyler-technologies/go-tfdr/total?color=orange"></a>
    <a href="https://github.com/mupuri/go-tfdr/tree/main"><img alt="Latest Commit" src="https://img.shields.io/github/last-commit/tyler-technologies/go-tfdr?color=ff69b4"></a>
    <a href="https://github.com/mupuri/go-tfdr/releases/latest"><img alt="Go Report" src="https://goreportcard.com/badge/github.com/mupuri/go-tfdr?style=flat-square"></a>

  </p>
</p>

## Synopsis
Go cli to copy terraform state file from one terraform cloud workspace to another or to 
delete terraform state of specified resources from a workspace.
This cli can be used in a disaster recovery scenario where new infrastructure needs to 
be set up for select resources in a template.
For example, if some infrastructure is set up in AWS using a terraform template and a 
region failure occurs, this cli can be used while setting up new infrastructure in a new 
region.

## Usage
```
tfdr [command]
```

#### Usage Docs
- [tfdr cli docs](./docs/tfdr.md)

## Example Disaster Recovery Steps
Here is an exmple of how this cli could be used in a disaster recovery scenario.
The following steps assumes there is already a terraform cloud workspace (`test1`) with a 
terraform template for setting up infrastructure 

### Steps
1. Create a new terraform workspace (`test2`) for the disaster recovery infrastructure
2. Create a json file (`filters.json`) with the list of resources whose state we need to copy 
   over from one workspace to another
3. Run the following command to copy state from the original workspace to the new 
   workspace
   ```
   tfdr state copy -f filters.json -o test1 -n test2
   ```
4. Plan and apply the new workspace
5. Run the following command to delete state of the copied over resources from the original 
   workspace
   ```
   tfdr state delete -f filters.json -w test1
   ```

## Example filters.json file
- `global_resource_types` contains any resource types you would like to be moved to the new 
  workspace regardless of resource or module name. In the example below, this list was populated
  with global AWS resources. While copying infrastructure from one AWS region to another, global 
  AWS resources that do not have any region should not be re-created. This property allows us to 
  copy the state of those global AWS resources when setting up a disaster recovery infrastructure.
- `filters` contains a list of specific resources in the terrafrom template to copy state of.
  - `filter_properties` contains information about the resource whose state we want to copy.
  - `new_properties` can contain any properties in the state we would like to replace for that resource. 
    Currently the cli allows updating the name of the copied over resource or any instance attributes 
    in the state of the copied over resource.
```
{
    "global_resource_types": [
        "aws_cloudfront_distribution",
        "aws_cloudfront_origin_access_identity",
        "aws_iam_access_key",
        "aws_iam_policy_document",
        "aws_iam_policy",
        "aws_iam_role_policy_attachment",
        "aws_iam_role_policy",
        "aws_iam_role",
        "aws_iam_user_policy",
        "aws_iam_user",
        "aws_route53_record"
    ],
    "filters": [
        {
            "filter_properties": {
                "module": "module.test_module_1",
                "type": "type_1",
                "name": "orig_name_1"
            },
            "new_properties": {
                "name": "new_name_1",
                "attributes": {
                    "attr1": "new_value_1",
                    "attr2": ""
                }
            }
        },
        {
            "filter_properties": {
                "module": "module.test_module_2",
                "type": "type_2",
                "name": "orig_name_2"
            },
            "new_properties": {
                "name": "new_name_2"
            }
        },
        {
            "filter_properties": {
                "module": "module.test_module_3",
                "type": "type_3",
                "name": "orig_name_3"
            },
            "new_properties": {
                "name": "new_name_3",
                "attributes": {
                    "attr1": "new_value_3",
                    "attr2": "new_value_3b"
                }
            }
        }
    ]
}
```