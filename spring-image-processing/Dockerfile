# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM maven:3.9.6-eclipse-temurin-21 AS build
WORKDIR /build

# Cache dependencies layer
COPY pom.xml .
RUN mvn dependency:go-offline -q

# Build the fat jar (skip tests; CI runs them separately)
COPY src ./src
RUN mvn package -DskipTests -q

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Non-root user for security
RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=build /build/target/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
