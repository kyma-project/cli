# Adding a new Module using Kyma CLI
This tutorial shows how you can add a new Module using Kyma CLI. 

Keep in mind that Modules added without any specified form of CR have the policy field set to Ignore.
## Steps

### Adding a new module

1. To add a new Module with the default policy set to `CreateAndDelete`, use the command below:

    ```
    kyma alpha module add {MODULE-NAME} --default-cr
    ```
2. If you would like to add a Module with a different CR, use the `--cr-path={CR-FILEPATH}` flag, as shown below:
    ```
    kyma alpha module add {MODULE-NAME} --cr-path={CR-PATH-FILEPATH}
    ```
3. You can also specify which channel the Module should use with the `-c {CHANNEL-NAME}` flag, as shown below:
    ```
    kyma alpha module add {MODULE-NAME} -c {CHANNEL-NAME} --default-cr
    ```
### Deleting an existing module

1. To delete an existing module, use the command below:

    ```
    kyma alpha module delete {MODULE-NAME} 
    ```