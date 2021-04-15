---
title: Function's configuration file
type: Details
---

When you initialize a Function (`init`), CLI creates the `config.yaml` file in your workspace folder. This file contains the whole Function's configuration and specification not only for the Function custom resource but also any other related resources you create for it, such as Subscriptions and API Rules.

## Definition

The specification of the `config.yaml` looks as follows:

```yaml
name: function-practical-filip5
namespace: testme
runtime: nodejs12
source:
    sourceType: inline
    sourcePath: /tmp/cli
subscriptions:
    - name: function-practical-filip5
      protocol: ""
      filter:
        dialect: ""
        filters:
            - eventSource:
                property: source
                type: exact
                value: ""
              eventType:
                property: type
                type: exact
                value: sap.kyma.custom.demo-app.order.created.v1
apiRules:
    - name: function-practical-filip5
      service:
        host: path.34.90.136.181.xip.io
      rules:
        - methods:
            - GET
            - POST
            - PUT
            - PATCH
            - DELETE
            - HEAD
          accessStrategies: []
        - path: /path1/something1
          methods:
            - PUT
            - PATCH
            - DELETE
          accessStrategies:
            - handler: noop
        - path: /path1/something2
          methods:
            - GET
            - HEAD
          accessStrategies:
            - config:
                requiredScope:
                    - asd1.asd2.asd3
                    - scope1
                    - scope2
              handler: oauth2_introspection
        - path: /path2
          methods:
            - DELETE
          accessStrategies:
              handler: jwt
            - config:
                jwksUrls:
                    - http://dex-service.kyma-system.svc.cluster.local:5556/keys
                    - http://dex-service.kyma-system.svc.cluster.local:5556
                trustedIssuers:
                    - https://dex.34.90.136.181.xip.io
                    - https://dex.34.90.136.181.xip.io
env:
    - name: REDIS_PASS
      value: YgJUg8z6eA
    - name: REDIS_PORT
      value: "6379"
    - name: REDIS_HOST
      value: hb-redis-enterp-6541066a-edbc-422f-8bef-fafca0befea8-redis.testme.svc.cluster.local
```

## Parameters

See all parameter descriptions:

| Parameter         | Required | Related custom resource | Description                                   |
| ---------------------------------------- | :------------: | ---------| ---------|
| **name**              | Yes | Function | Specifies the name of your Function.         |
| **namespace**             | No | Function | Defines the Namespace in which the Function is created. It is set to `default` unless you specify otherwise.         |
| **runtime**             | No | Function | Specifies the execution environment for your Function. The available values are `nodejs12`, `python38` and deprecated `nodejs10`. It is set to `nodejs12` unless specified otherwise.         |
| **source**            | Yes | Function | Provides details on the type and location of your Function's source code.         |
| **source.sourceType**            | Yes | Function | Defines that you use inline code or Git repository as the source of the Function's code and dependencies. It must be set either to `inline` or `git`.         |
| **source.sourcePath**             | Yes | Function | Specifies the absolute path to the directory with the Function's source code.         |
| **subscriptions.name**           |  Yes | Subscription | Specifies the name of the Subscription custom resource. It takes the name from the Function unless you specify otherwise.    |
| **subscriptions.protocol**           | Yes  | Subscription | Defines the rules and formats applied for exchanging messages between the components of a given messaging system. Subscriptions in Kyma CLI use the [NATS](https://docs.nats.io/) messaging protocol by default. Must be set to `""`.         |
| **subscriptions.filter.dialect**            | No | Subscription | Indicates the filter expression language supported by an event producer. Subscriptions specifying the **filter** property must specify the dialect as well. All other properties are dependent on the dialect being used. In the current implementation, this field is treated as a constant which is blank.    |
| **subscriptions.filter.filters.eventSource**            | Yes | Subscription | The origin from which events are published.         |
| **subscriptions.filter.filters.eventSource.property**            | Yes | Subscription | Must be set to `source`.         |
| **subscriptions.filter.filters.eventSource.type**           | No  | Subscription | Must be set to `exact`.         |
| **subscriptions.filter.filters.eventSource.value**            | Yes | Subscription | Must be set to `""` for the NATS backend.         |
| **subscriptions.filter.filters.eventType**           | Yes  | Subscription | The type of events used to trigger workloads.         |
| **subscriptions.filter.filters.eventType.property**           | Yes  | Subscription | Must be set to `type`.         |
| **subscriptions.filter.filters.eventType.type**          |  No  | Subscription | Must be set to `exact`.         |
| **subscriptions.filter.filters.eventType.value**           | Yes  | Subscription | Name of the event the Function is subscribed to, for example `sap.kyma.custom.demo-app.order.created.v1`.         |
| **apiRules.name**            | Yes | API Rule | Specifies the name of the exposed service. It takes the name from the Function unless you specify otherwise.        |
| **apiRules.service.host**            | Yes | API Rule | Specifies the service's communication address for inbound external traffic.         |
| **apiRules.rules**            | Yes | API Rule | Specifies the array of [Oathkeeper](https://www.ory.sh/oathkeeper/) access rules.         |
| **apiRules.rules.methods**            | No | API Rule | Specifies the list of HTTP request methods available for **apiRules.rules.path** .        |
| **apiRules.rules.accessStrategies**            | Yes | API Rule | Specifies the array of [Oathkeeper authenticators](https://www.ory.sh/oathkeeper/docs/pipeline/authn/). The supported authenticators are `oauth2_introspection`, `jwt`, `noop`, and `allow`.         |
| **apiRules.rules.path**            | Yes | API Rule | Specifies the path of the exposed service.         |
| **apiRules.rules.path.accessStrategies.handler**            |  | API Rule | Specifies one of authenticators used: `oauth2_introspection`, `jwt`, `noop`, or `allow`.       |
| **apiRules.rules.path.accessStrategies.config.** | No | API Rule |  Defines the handler used. It can be specified globally or per access rule.         |
| **env.name**            | No | Function | Specifies the name of the environment variable to export for the Function.  |
| **env.value**            | No | Function | Specifies the value of the environment variable to export for the Function.          |

## Related resources

See the detailed descriptions of all related custom resources referred to in the `config.yaml`:

- [Function](https://kyma-project.io/docs/main/components/serverless/#custom-resource-function)
- [Subscription](https://kyma-project.io/docs/main/components/eventing/#custom-resource-subscription)
- [API Rule](https://kyma-project.io/docs/main/components/api-gateway/#custom-resource-api-rule)
