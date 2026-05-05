# Movies API

A simple CRUD REST service for managing movies, backed by SAP BTP Object Store. Built with Spring Boot and designed to be deployed to SAP BTP Kyma runtime using `kyma app push`.

## Prerequisites

The service uses SAP BTP Object Store (S3-compatible) for persistence. A Kubernetes Secret named `object-store-binding` containing the Object Store service binding credentials must exist in the same namespace before deployment. See `k8s/service-instance.yaml` and `k8s/service-binding.yaml` for the required resources.

## Deploy

```bash
kyma app push \
  --name my-prototype \
  --code-path . \
  --container-port 8080 \
  --expose \
  --istio-inject=true \
  --mount-service-binding-secret object-store-binding \
  --env-from-file .env
```

## API Documentation

Once deployed, the Swagger UI is available at:

```
https://<YOUR-APP-URL>/swagger-ui.html
```

The OpenAPI spec (JSON) is at:

```
https://<YOUR-APP-URL>/v3/api-docs
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /movies | List all movies |
| GET | /movies/{id} | Get a movie by ID |
| POST | /movies | Create a new movie |
| PUT | /movies/{id} | Update a movie |
| DELETE | /movies/{id} | Delete a movie |
