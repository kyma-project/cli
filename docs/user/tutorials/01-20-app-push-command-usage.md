# Running an application using `app push` command

This tutorial shows how you can deploy your application using Kyma CLI.

## Procedure

### Downloading required modules

As a prerequisite to use `kyma@v3 alpha app push` you need to add necessary modules.

1. Run the following commands:

    ```
    kyma@v3 alpha module add istio --default-cr
    kyma@v3 alpha module add docker-registry -c experimental --default-cr
    kyma@v3 alpha module add api-gateway --default-cr
    ```
### Preparing project source code

To use `kyma@v3 alpha app push` you also need to provide either Dockerfile, Docker image, or application's source code. In this tutorial, we will use application source code. For example you can use one of the code samples from [paketo buildpacks code examples](https://github.com/paketo-buildpacks/samples/tree/main), but in this tutorial we will work on [paketo buildpacks java maven app](https://github.com/paketo-buildpacks/samples/tree/main/java/maven).

1. Clone [paketo code examples](https://github.com/paketo-buildpacks/samples/tree/main) repository into desired folder, using the following command:
   ```
   https://github.com/paketo-buildpacks/samples.git
   ```
2. Navigate to `java/maven` directory
   ```
   cd java/maven
   ```


### Run application Deployment on a cluster

After fulfilling all the prerequisites you can now deploy your application using `kyma@v3 alpha app push`. Besides required `--name` flag, you also need to use either `--dockerfile` flag to run application from Dockerfile, `--image` flag to run application from docker image, or `--code-path` flag to run application from source code.

1. To run application deployment on a cluster, run the following command in `java/maven` directory:
   ```
   kyma@v3 alpha app push --name=Test-App --code-path=.
   ``` 

### Run application Deployment with exposed right port, on a cluster 

You can also use the `--container-port` flag to deploy an application with own service, with exposed right port.

1. To run application deployment with exposed right port on a cluster, run the following command in `java/maven` directory:
   
    ```
    kyma@v3 alpha app push --name=Test-App --code-path=. --container-port=8888
    ```

### Run application Deployment with APIRule allowing outside access, on a cluster 

You can also use the `--expose` flag to deploy an application with own APIRule allowing access from outside the cluster. Note that when using `--expose` flag, you need to also provide `--container-port`.


1. To run application deployment with APIRule allowing outside access on a cluster, run the following command in `java/maven` directory:

    ```
    kyma@v3 alpha app push --name=Test-App --code-path=. --container-port=8888 --expose
    ```

### Check deployed application connection


To check if deployed application connection is working properly, you can perform a curl request. The procedure differs depending on how did you deploy your application, but you should always get the same result, presented below:

    
    {"status":"UP","groups":["liveness","readiness"]}
    

#### Application deployed without the --container-port flag
1. If you deployed your application without the --container-port flag, use the following command:

    ```
    kubectl port-forward deployment/Test-App 8080:8080
    curl localhost:8080/actuator/health
    ```

#### Application deployed with the --container-port flag
1. If you deployed your application with the --container-port flag, use the following command:

    ```
    kubectl port-forward svc/Test-App 8080:8080
    curl localhost:8080/actuator/health
    ```

#### Application deployed with the --expose flag
1. If you deployed your application with the --expose flag, use the following command:

    ```
    curl {ADDRESS-RETURNED-FROM-THE-APP-PUSH}:8080/actuator/health
    ```
