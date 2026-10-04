# Java Testing

- JUnit 5 + AssertJ + Mockito.
- `@WebMvcTest` for controllers, `@DataJpaTest` for repositories, `@SpringBootTest` sparingly.
- Testcontainers for Postgres/MySQL/Kafka.
- Parameterized tests (`@ParameterizedTest`) for boundary cases.
- Spring Security tests with `@WithMockUser` and cross-tenant cases.
