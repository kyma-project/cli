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
### Prepare project source code

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

After fulfilling all the prerequisites you can now deploy your application using `kyma@v3 alpha app push`. Besides required `--name` flag, you also need to use `--code-path` flag to run application from source code. 

3. To run application deployment on a cluster, with own APIRule allowing outside access, run the following command in current directory:

   ```
   kyma@v3 alpha app push --name=Test-App --code-path=. --container-port=8888 --expose
   ``` 

> [!NOTE]
> Depending on your needs, you can also create deployments of your applications without `--expose` or `--container-port` flags. This will change the way you can communicate with your application.

4. After running application deployment, you should get an URL address in return. Keep track of it, as we will use it in the next step.

### Check deployed application connection

5. To check if deployed application connection is working properly, you can perform a curl request. 
   
   ```
   curl {ADDRESS-RETURNED-FROM-THE-APP-PUSH}:8080/actuator/health
   ```  

   You should get the following response:

   ``` 
   {"status":"UP","groups":["liveness","readiness"]}
   ```


> [!NOTE]
> Depending on how did you deploy your application, the way you communicate with it differs. 
> 
> Without the `--container-port` you should port-forward the deployment in one terminal, and then check the health of an application in another one.
> 
> kubectl port-forward deployment/Test-App 8080:8080
> 
> curl localhost:8080/actuator/health
> 
> With the `--container-port` you should port-forward the deployment in one terminal, and then check the health of an application in another one.
> 
> kubectl port-forward svc/Test-App 8080:8080
> 
> curl localhost:8080/actuator/health
