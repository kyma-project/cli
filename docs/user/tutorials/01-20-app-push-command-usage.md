# Running an application using `app push` command

This tutorial shows how you can deploy your application using Kyma CLI.

## Procedure

### Downloading required modules

As a prerequisite to use `kyma alpha app push` you need to add necessary modules, using the following commands:

    ```
    kyma alpha module add istio --default-cr
    kyma alpha module add docker-registry -c experimental --default-cr
    kyma alpha module add api-gateway --default-cr
    ```

### Preparing project source code

To use `kyma alpha app push` you also need to provide either Dockerfile, Docker image, or application's source code. If you are using the latter, you can use one of the code samples from [paketo buildpacks code examples](https://github.com/paketo-buildpacks/samples/tree/main). Below you can see the main function of one of the examples, [paketo buildpacks java maven app](https://github.com/paketo-buildpacks/samples/tree/main/java/maven).

    ```
    package io.paketo.demo;

    import org.springframework.boot.SpringApplication;
    import org.springframework.boot.autoconfigure.SpringBootApplication;

    @SpringBootApplication(proxyBeanMethods=false)
    public class DemoApplication {

	public static void main(String[] args) {
		SpringApplication.run(DemoApplication.class, args);
	    }

    }
    ```

### Run application Deployment on a cluster

After fulfilling all the prerequisites you can now deploy your application using `kyma alpha app push`. Besides required `--name` flag, you also need to use either `--dockerfile` flag to run application from Dockerfile, `--image` flag to run application from docker image, or `--code-path` flag to run application from source code, as shown below:

    ```
    kyma alpha app push --name={APPLICATION-NAME} --code-path={SOURCE-CODE-PATHFILE}
    ```

### Run application Deployment with exposed right port, on a cluster 

You can also use the `--container-port` flag to deploy an application with own service, with exposed right port.

    ```
    kyma alpha app push --name={APPLICATION-NAME} --code-path={SOURCE-CODE-PATHFILE} --container-port={PORT-TO-EXPOSE-APPLICATION-ON}
    ```

### Run application Deployment with APIRule allowing outside access, on a cluster 

You can also use the `--expose` flag to deploy an application with own APIRule allowing access from outside the cluster. Note that when using `--expose` flag, you need to also provide `--container-port`.

    ```
    kyma alpha app push --name={APPLICATION-NAME} --code-path={SOURCE-CODE-PATHFILE} --container-port={PORT-TO-EXPOSE-APPLICATION-ON} --expose
    ```

### Check deployed application connection


To check if deployed application connection is working properly, you can perform a curl request. The procedure differs depending on how did you deploy your application.

#### Application deployed without the --container-port flag

    ```
    kubectl port-forward deployment/{APPLICATION-NAME} {PORT}:{PORT}
    curl localhost:{PORT}/{ENDPOINT}
    ```

#### Application deployed with the --container-port flag

    ```
    kubectl port-forward svc/{SVC_NAME} {PORT}:{PORT}
    curl localhost:{PORT}/{ENDPOINT}
    ```

#### Application deployed with the --exposed flag

    ```
    curl {ADDRESS-RETURNED-FROM-THE-APP-PUSH}:{PORT}/{ENDPOINT}
    ```
