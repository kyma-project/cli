# Output Error Format

Creation date: 2024.03.13

## Description

I would propose a standard error format for any CLI command:

```text
kyma provision --credentials-path ~/Desktop/fs-binding.txt --plan aws --region northeurope

Error:
  failed to provision Kyma runtime

Error Details:
  failed to provision: User is unauthorized for this operation

Hints:
  - make sure that the provided credentials are valid are represent the local CIS instance
  - make sure your Subaccount has the Kyma entitlement enabled
```

Elements:

* `Error` contains information about the error cause.
* `Error Details` contains an error returned from the library/endpoint.
* `Hints` contains suggestions and hints on what users can do to avoid the problem.

## Reasons

The `kyma` cli integrates other tools with different output formats. Additionally, CIS doesn't clearly communicate what caused the error. The following error results from the provision command when the subaccount does not enable the Kyma entitlement. The error suggests that the problem is related to the unauthorized operation, and it's really hard to say what is going on.

```text
kyma provision --credentials-path ~/Desktop/fs-binding.txt --plan aws --region northeurope
Error: failed to provision kyma runtime: failed to provision: User is unauthorized for this operation
exit status 1
```

In the `Error` section, I proposed an error output standard that can help us improve error readability and allow us to print more details and hints for the user.

Because we know which command code fails, we can predict what causes the problem. In the example above, we (from a code perspective) know that the code fails on the Kyma instance provisioning, so we know the operation context. It goes wrong after getting a token based on the given credentials, so we know the given token (by the XSUAA) is valid.
Based on this information we can predict a few possible causes and print some tips and hints (for example prerequisites for the specific operation).

Another important feature is control over stdout/stderr channels. We can decide that for example, the content of the `Error Description` section should go to the stderr, but everything else should go to the stdout.
