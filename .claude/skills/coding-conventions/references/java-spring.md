# Java / Spring Boot Conventions

## Layout (package by feature)
```
com.company.app
  config/
  common/ (exceptions, error handler, security utils)
  user/
    UserController.java
    UserService.java
    UserRepository.java
    User.java          (entity)
    dto/ CreateUserRequest.java, UserResponse.java
```

## Style
- Java 17+ (records for DTOs), Lombok only if the project already uses it.
- Constructor injection (final fields), no field `@Autowired`.
- Bean Validation (`@Valid`, `@NotBlank`, `@Email`) on request DTOs.
- Never expose JPA entities in controllers; map to DTOs (MapStruct or manual).

## Errors
- Domain exceptions (`NotFoundException`, `ConflictException`) + one `@RestControllerAdvice` producing the standard error envelope.

## Transactions
- `@Transactional` on service methods, not controllers; read-only where applicable.

## Security
- Spring Security with method-level authorization (`@PreAuthorize`); actuator endpoints restricted.

## Tooling
- Checkstyle/Spotless, SpotBugs + FindSecBugs, OWASP Dependency-Check.
