# Setting Your Module to the Managed and Unmanaged State in Kyma CLI

In some cases, for example, for testing, you may need to modify your module beyond what is supported by its configuration. By default, when a module is in the managed state, Kyma Control Plane governs its Kubernetes resources, reverting any manual changes during the next reconciliation loop. 

To modify Kubernetes objects directly without them being reverted, you must set the module to the unmanaged state. In this state, reconciliation is disabled, ensuring your manual changes are preserved.

> [!CAUTION]
> Setting your module to the unmanaged state may lead to instability and data loss within your cluster. It may also be impossible to revert the changes. In addition, we don't guarantee any service level agreement (SLA) or provide updates and maintenance for the module.


## Procedure

### Setting a Module to the Managed State


1. To set a module to the managed state, use the following command:

    ```
    kyma alpha module manage {MODULE-NAME}
    ```
2. Even if the module is already in the managed state, you can change its policy by adding the optional flag ``--policy {POLICY-NAME}``. The default policy is ``CreateAndDelete``.

### Setting a Module to the Unmanaged State

1. To set a module to the unmanaged state, use the following command::

    ```
    kyma alpha module unmanage {MODULE-NAME}
    ```

> [!CAUTION]
> Depending on the introduced changes, bringing back the module to the managed state might not be possible.
