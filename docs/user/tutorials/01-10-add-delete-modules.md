# Adding and Deleting a Kyma Module Using Kyma CLI

This tutorial shows how you can add and delete a new module using Kyma CLI.

> [!WARNING]
> Modules added without any specified form of the custom resource have the policy field set to `Ignore`.

## Procedure

### Adding a New Module

1. Check the list of modules that can be added:

   ```bash
   kyma module catalog
   ```

2. Add a new module:
   * To add a new module with the default policy set to `CreateAndDelete`, use the following command:

      ```bash
      kyma module add {MODULE-NAME} --default-cr
      ```

   * To add a module with a different CR, use the `--cr-path={CR-FILEPATH}` flag:

      ```bash
      kyma module add {MODULE-NAME} --cr-path={CR-PATH-FILEPATH}
      ```

To specify which channel the module should use, add the `-c {CHANNEL-NAME}` flag:

   ```bash
   kyma module add {MODULE-NAME} -c {CHANNEL-NAME} --default-cr
   ```
3. To see if your module is added, run the following command:

   ```bash
   kyma module list
   ```

### Deleting an Existing Module

To delete an existing module, use the following command:

   ```bash
   kyma module delete {MODULE-NAME} 
   ```
