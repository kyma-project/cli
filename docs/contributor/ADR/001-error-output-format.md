# Output Error Format

Creation date: 03.13.2024

## Description

I would propose standard error format for any CLI command:

```text
kyma provision --credentials-path ~/Desktop/fs-binding.txt --plan aws --region northeurope

Error:
  failed to provision kyma runtime

Raw Error:
  failed to provision: User is unauthorized for this operation

Hints:
  - make sure that the provided credentials are valid are represent the local CIS instance
  - make sure the Subaccount has kyma entitlement enabled
```

Elements:

* The `Error` contains information which operation caused the error.
* The `Raw Error` contains error returned from the library/endpoint.
* The `Hints` contains suggestions, hints what user can do to avoid problem.

## Reasons

The `kyma` cli integrates other tools with different output formats. Another problem is not all CIS communicates clearly what causes the error. You can see below an error from the provision command when the subaccount does not enable kyma entitlement. Error suggests that the problem is related to the unauthorized operation and it's really hard to say what is going on.

```text
kyma provision --credentials-path ~/Desktop/fs-binding.txt --plan aws --region northeurope
Error: failed to provision kyma runtime: failed to provision: User is unauthorized for this operation
exit status 1
```

In the `Description` section I proposed an error output standard that can help us improve error readability and allow us to print more details and hints to the user.

Because we know which command code fails we can predict which caused the problem. In the example above we (from a code perspective) know that the code fails on the kyma instance provisioning so we know the operation context and it gets wrong after getting a token based on the given credentials so we know that the given token (by the XSUAA) is valid.
Based on this information we can predict a few possible causes and print some tips and hints (for example prerequisites for the specific operation).

Another important feature is control over stdout/stderr channels. We can decide that for example content of the `Error` section should go to the stderr but everything else should go to the stdout.
