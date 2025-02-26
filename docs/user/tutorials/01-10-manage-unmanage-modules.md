# Setting Your Module to the Managed/Unmanaged State in Kyma CLI

In some cases, for example, for testing, you may need to modify your module beyond what is supported by its configuration. By default, when a module is in the managed state, Kyma Control Plane governs its Kubernetes resources, reverting any manual changes during the next reconciliation loop. 

To modify Kubernetes objects directly without them being reverted, you must set the module to the unmanaged state. In this state, reconciliation is disabled, ensuring your manual changes are preserved.

> [!CAUTION]
> Setting your module to the Unmanaged state may lead to instability and data loss within your cluster. It may also be impossible to revert the changes. In addition, we don't guarantee any service level agreement (SLA) or provide updates and maintenance for the module.


## Steps

### Setting a Module to the Managed State


1. To set Module to the Managed state, use the command below:

    ```
    kyma alpha module manage {MODULE-NAME}
    ```
2. Even if the Module is already Managed, you can change it's policy by adding optional flag ``--policy {POLICY-NAME}``. The default policy is ``CreateAndDelete``.

### Setting a Module to the Unmanaged State

1. To set Module to the Unmanaged state, use the command below:

    ```
    kyma alpha module unmanage {MODULE-NAME}
    ```

> [!CAUTION]
> Depending on the introduced changes, bringing back the module to the Managed state might not be possible.
