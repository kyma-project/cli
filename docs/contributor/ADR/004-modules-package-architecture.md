# ADR 004: Modules Package Architecture Refactoring

**Creation Date:** 2025-12-17
**Authors:** Antoni Pstraś

## Context

The original `modules` package in the Kyma CLI suffered from several architectural problems that made it difficult to maintain, extend, and test:

### Problems with the Original `modules` Package

1. **Lack of Abstraction and Separation of Concerns**
   - Business logic was tightly coupled with data access logic
   - Functions directly accessed the Kubernetes client and external APIs without proper abstraction layers
   - No clear boundaries between different responsibilities (data fetching, business logic, presentation)
   - Mapping between Kubernetes resources and domain models was scattered throughout the codebase

2. **Code Duplication**
   - Similar logic for handling core and community modules was duplicated across multiple functions
   - Validation and filtering logic was repeated in multiple places

3. **Poor Readability and Maintainability**
   - Mixed concerns in single files (data access, business logic, presentation)

4. **Testing Difficulties**
   - Tight coupling made unit testing complex, requiring extensive mocking
   - No dependency injection framework, leading to hard-coded dependencies
   - Difficult to test business logic in isolation from infrastructure concerns

5. **Lack of Domain Modeling**
   - Used raw Kubernetes types (`kyma.ModuleTemplate`) directly in business logic
   - No clear domain entities representing the business concepts
   - Data Transfer Objects (DTOs) mixed with domain entities

## Decision

We decided to create a new `modulesv2` package with a clean, layered architecture based on established design patterns and SOLID principles.

### Architectural Layers

The new package is organized into three clear layers with well-defined responsibilities:

#### Architectural Layers Diagram

Below is a graphical representation of the architecture and the flow of communication between layers:

```text
┌────────────┐
│   User     │
└─────┬──────┘
      │
      ▼
┌────────────────────┐
│   UI Layer         │
│ (Cobra CLI)        │
└─────┬──────────────┘
      │
      ▼
┌──────────────────────────────┐
│      Domain Layer            │
│ (Services, Entities,         │
│  Value Objects)              │
└─────┬────────────────────────┘
      │
      ▼
┌──────────────────────────────┐
│ Infrastructure Layer         │
│ (K8s, DB, REST/gRPC, etc.)   │
└──────────────────────────────┘
```

**Flow:**

- The **User** interacts with the **UI Layer** (Cobra CLI)
- The **UI Layer** invokes operations in the **Domain Layer**
- The **Domain Layer** encapsulates business logic in services, operates on entities and (optionally) value objects, and interacts with the **Infrastructure Layer**
- The **Infrastructure Layer** provides data access (e.g., Kubernetes, databases, external APIs)

#### 1. **UI Layer** (Cobra CLI)

- The entry point for user interaction with the Kyma CLI
- Implements the command-line interface using Cobra
- Responsible for parsing user input, displaying output, and invoking domain layer operations
- Decoupled from business logic and data access

#### 2. **Domain Layer** (`modulesv2/` root, `entities/`)

- Contains business logic and orchestration in services (e.g., `CatalogService`)
- Operates on domain entities (e.g., `CoreModuleTemplate`, `CommunityModuleTemplate`)
- May use value objects to encapsulate domain concepts (currently not present, but recommended for future evolution)
- Uses dependency injection for loose coupling
- Returns DTOs for presentation layer

#### 3. **Infrastructure Layer** (`repository/`)

- Implements the Repository Pattern for data access
- Provides abstraction over data sources (Kubernetes cluster, databases, external APIs)
- Responsible for reading and writing data
- Maps between infrastructure types (raw Kubernetes resources, DB records, API responses) and domain entities

### Design Patterns Applied

#### Entity Pattern

In Domain-Driven Design (DDD), an Entity is an object defined by its identity and lifecycle, typically persisted in a data store. In this project, we use the Entity pattern to represent Kubernetes data as domain objects, such as `CoreModuleTemplate` and `CommunityModuleTemplate`.

- These entities encapsulate business rules and represent the core concepts of the module system.
- While they may not always have a traditional unique ID field, their identity and lifecycle are tied to their persistence in the Kubernetes cluster (e.g., name and namespace).
- The entity is created by mapping raw Kubernetes resources (e.g., `kyma.ModuleTemplate`) to a domain-specific struct, which is then used throughout the domain and service layers.

**Example:**

```go
// internal/modulesv2/entities/core_module_template.go
type CoreModuleTemplate struct {
   Name        string
   Version     string
   Description string
   // ...other business fields...
}

// Mapping function from k8s resource to entity
func mapToCoreModuleTemplate(raw kyma.ModuleTemplate) CoreModuleTemplate {
   return CoreModuleTemplate{
      Name:        raw.Metadata.Name,
      Version:     raw.Spec.Version,
      Description: raw.Spec.Description,
      // ...
   }
}
```

**How it's used:**

- The repository fetches raw data from Kubernetes and maps it to entities.
- The service layer operates on these entities, applying business logic and orchestrating workflows.
- Entities are passed to the UI layer (CLI) for presentation or further processing.

This approach allows us to encapsulate business logic within entities, maintain a clear separation from infrastructure concerns, and model the lifecycle of domain objects that are persisted in the cluster.

**When to use:**

- Represent core concepts with identity and lifecycle (e.g., core/community module templates).
- Centralize invariants/rules close to the data that they govern.

**Anti-patterns:**

- Passing raw Kubernetes types through the domain/services.
- Embedding persistence (API calls) inside entities.

#### Repository Pattern

The repository pattern is responsible for accessing infrastructure data (e.g., from a Kubernetes cluster) and mapping it into domain entities. It provides a clean interface for the domain layer to interact with data sources, hiding the details of data fetching and transformation.

**Simple example:**

Suppose we want to access module templates from a Kubernetes cluster and expose them as domain entities:

1. **Define the repository interface:**

   ```go
   // internal/modulesv2/repository/moduletemplates.go
   type ModuleTemplatesRepository interface {
      ListCore() ([]CoreModuleTemplate, error)
   }
   ```

2. **Implement the repository:**

   ```go
   type moduleTemplatesRepository struct {
      kubeClient kube.Client
   }

   func (r *moduleTemplatesRepository) ListCore() ([]CoreModuleTemplate, error) {
      // Fetch raw k8s resources
      rawTemplates, err := r.kubeClient.ListModuleTemplates()
      if err != nil {
         return nil, err
      }
      // Map raw resources to domain entities
      var result []CoreModuleTemplate
      for _, raw := range rawTemplates {
         result = append(result, mapToCoreModuleTemplate(raw))
      }
      return result, nil
   }
   ```

3. **Use the repository via the interface:**

   ```go
   func PrintCoreModules(repo ModuleTemplatesRepository) error {
      modules, err := repo.ListCore()
      if err != nil {
         return err
      }
      for _, m := range modules {
         fmt.Println(m.Name)
      }
      return nil
   }
   ```

This approach allows the domain and service layers to remain decoupled from infrastructure details, and makes testing easier by allowing the use of fake repositories.

**When to use:**

- Access external data sources (Kubernetes, DBs, APIs) and return domain entities.
- Centralize mapping from infrastructure types into entities.

**Anti-patterns:**

- Calling Kubernetes clients directly from services.
- Placing complex business logic inside repositories (beyond mapping/filtering).

The Repository Pattern was selected over the Active Record pattern for several reasons:

**Why Repository Pattern:**

- ✅ **Separation of Concerns**: Business logic is completely separated from data access
- ✅ **Testability**: Easy to mock repositories in service tests
- ✅ **Flexibility**: Can change data sources without affecting business logic
- ✅ **SOLID Compliance**: Single Responsibility and Dependency Inversion principles

**Why NOT Active Record Pattern:**

- ❌ **Tight Coupling**: Domain objects would be coupled to k8s API access
- ❌ **Testing Complexity**: Difficult to test business logic without infrastructure
- ❌ **Violation of SRP**: Objects would have both business logic and persistence logic
- ❌ **Limited Flexibility**: Harder to swap data sources

#### Dependency Injection Pattern

The Kyma CLI uses a custom dependency injection (DI) container (`internal/di`) to manage and resolve dependencies for services and repositories in a flexible, testable, and loosely coupled way.

**How it works:**

- The DI container is responsible for resolving dependencies for each service, doing so in a lazy (on-demand) manner.
- It uses Go’s reflection to map types to factory functions, which know how to construct each dependency.
- When a service is requested, the container checks if an instance exists; if not, it calls the registered factory, resolving further dependencies recursively.
- This enables loose coupling, easy testing, and clear dependency graphs.

**Key features:**

- Lazy instantiation: dependencies are only created when needed.
- Singleton scope: each type is instantiated once per container.
- Type safety: generic helpers (`GetTyped`, `RegisterTyped`) ensure correct types.
- Testability: you can register fake factories for testing.

**Example (based on actual implementation):**

Suppose we want to inject a `ModuleTemplatesRepository` into a `CatalogService` using the DI container:

1. **Registering factories:**

   ```go
   di.RegisterTyped(container, func(c *di.Container) (repository.ModuleTemplatesRepository, error) {
      kubeClient, err := di.GetTyped[kube.Client](c)
      externalRepo, err := di.GetTyped[repository.ExternalModuleTemplateRepository](c)
      return repository.NewModuleTemplatesRepository(kubeClient, externalRepo), nil
   })
   di.RegisterTyped(container, func(c *di.Container) (*CatalogService, error) {
      moduleRepo, err := di.GetTyped[repository.ModuleTemplatesRepository](c)
      metadataRepo, err := di.GetTyped[repository.ClusterMetadataRepository](c)
      return NewCatalogService(moduleRepo, metadataRepo), nil
   })
   ```

2. **Lazy resolution:**

   ```go
   catalogService, err := di.GetTyped[*CatalogService](container)
   ```

   - The container will resolve all dependencies recursively, using reflection and the registered factories.

3. **Step-by-step explanation:**
   1. Factory functions for each type are registered in the container.
   2. When a service is requested, the container checks if an instance exists; if not, it calls the factory, resolving further dependencies recursively.
   3. Each dependency is only created once and reused for subsequent requests.
   4. For testing, you can register fake factories to inject mock dependencies.

**Summary**:

- Uses a custom DI container (`internal/di`)
- Constructor injection for explicit dependencies
- Enables:
  - Easy testing with fake implementations
  - Loose coupling between components
  - Clear dependency graphs
  - Runtime configuration flexibility

#### Factory Pattern (for testing purposes)

The Factory Pattern is used to streamline the creation of test data objects for unit tests. This approach is especially useful for testing purposes, as it allows developers to generate objects with sensible default values and override only the fields relevant to a specific test scenario. Files like [internal/modulesv2/fake/coremoduletemplate.go](../../../internal/modulesv2/fake/coremoduletemplate.go) are examples of such factories, but the pattern is used for various domain entities, not just core module templates.

**Key points:**

- Centralized test data creation: Instead of manually defining long sections of test data in each test file, the factory provides a single place to create instances of core module templates.
- Default values: The factory sets up default values for all fields, ensuring that each test object is valid and complete by default.
- Selective overrides: When writing tests, you can override only the fields that matter for your test case, keeping your test code concise and focused.
- Maintainability: Changes to the structure or defaults of the test objects are made in one place, reducing duplication and maintenance effort.

**Example usage:**

Suppose you have a factory function like `fake.CoreModuleTemplate()` in [coremoduletemplate.go](../../../internal/modulesv2/fake/coremoduletemplate.go):

```go
// Params struct allows you to specify only the fields you want to override
tmpl := fake.CoreModuleTemplate(&fake.Params{
    ModuleName: "module1",
    Version:    "1.0.0",
    Channel:    "fast",
})

// Passing nil applies all default values
tmpl := fake.CoreModuleTemplate(nil)
```

Any field not provided (or if you pass nil) will be filled with sensible defaults. This lets you create valid test objects with minimal boilerplate, overriding only what matters for your scenario.

**Benefits:**

- Reduces boilerplate: No need to repeat full object definitions in every test.
- Improves readability: Tests focus on what's relevant, not on setup.
- Encourages reuse: Factories can be extended for other test objects.

**When to use:**

- You need valid test data quickly with safe defaults.
- You want to override only a few fields per test scenario.

**Anti-patterns:**

- Sprawling test fixtures duplicated across files.
- Encoding assertions/business logic inside factories (they should only construct data).

#### Data Transfer Object (DTO) Pattern

The Data Transfer Object (DTO) pattern is a well-established design pattern used to transfer data between software application subsystems, layers, or services. In the Kyma CLI, DTOs serve as plain data containers that wrap and transport information:

- From the user to the application (input)
- From the application to the user (output/server response)
- Between internal layers of the application

DTOs are intentionally kept free of business logic and validation. Their purpose is to carry data, not to enforce rules or perform computations. The only logic they may contain is related to their own construction or mapping (e.g., converting from domain entities to DTOs and vice versa). Any business validation or processing should be handled in the domain or service layers, not in DTOs themselves. This clear separation ensures maintainability, testability, and a clean architecture.

**Example with DTO creation logic:**

```go
// Domain entity
type User struct {
   ID       string
   Name     string
   Email    string
   Password string // domain-only, not exposed in DTO
}

// DTO for transferring user data (e.g., to API response)
type UserDTO struct {
   ID    string `json:"id"`
   Name  string `json:"name"`
   Email string `json:"email"`
}

// Logic for creating a DTO from a domain entity (mapping only, no business logic)
func NewUserDTO(user User) UserDTO {
   return UserDTO{
      ID:    user.ID,
      Name:  user.Name,
      Email: user.Email,
   }
}
```

**When to use:**

- Accept user input in a defined shape, decoupled from domain models.
- Return responses from services/CLI without leaking internal types.
- Pass data between layers where business context is not required.

**Anti-patterns:**

- Embedding business rules or validation logic into DTOs.
- Using DTOs as domain entities in business logic.

### Key Improvements

1. **Clear Separation of Concerns**

   ```text
    modulesv2/
     ├── repository/
     │   ├── moduletemplates.go
     │   ├── externalmoduletemplates.go
     │   └── clustermetadata.go
     ├── entities/
     │   ├── coremoduletemplate.go
     │   └── communitymoduletemplate.go
     ├── dtos/
     │   ├── catalogconfig.go
     │   └── catalogresult.go
     ├── catalog.go                   (Service layer)
     └── dependencies.go              (DI configuration)
   ```

2. **Focused Interfaces**

3. **Testability**
   - Services can be tested with fake repositories
   - No need to mock Kubernetes clients in service tests
   - Repository tests focus on data access logic only

4. **Maintainability**
   - Each file has a single, clear purpose
   - Easy to locate and modify functionality
   - New features can be added without modifying existing code (Open/Closed Principle)

5. **Type Safety and Domain Modeling**
   - Domain entities with strong typing
   - Business rules enforced at compile time
   - Clear distinction between core and community modules

## Consequences

### Positive

- **Better Code Organization**: Clear layers with well-defined responsibilities
- **Improved Testability**: Easy to write unit tests for business logic
- **Reduced Duplication**: Common logic centralized in appropriate layers
- **Enhanced Maintainability**: Changes are localized and don't ripple through the codebase
- **Easier Onboarding**: New developers can understand the architecture quickly
- **Foundation for Growth**: Architecture supports adding new features without major refactoring

### Negative

- **Initial Learning Curve**: Team members need to understand the new patterns
- **More Files**: Separation of concerns results in more files, which could be perceived as complexity
- **Migration Effort**: Existing commands need to be migrated from `modules` to `modulesv2`
- **Temporary Duplication**: During migration, both packages will coexist

## References

- [Repository Pattern - Martin Fowler](https://martinfowler.com/eaaCatalog/repository.html)
- [Data Transfer Object (DTO) - Martin Fowler](https://martinfowler.com/eaaCatalog/dataTransferObject.html)
- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Layered Architecture - Tomasz Świacko-Świackiewicz](https://tswiackiewicz.github.io/inside-the-source-code/architecture/ddd-layered-architecture/)
