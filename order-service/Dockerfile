FROM maven:3.9-eclipse-temurin-17 AS build
WORKDIR /src
COPY pom.xml .
RUN mvn -B -q dependency:go-offline
COPY src ./src
RUN mvn -B -q package -DskipTests

FROM eclipse-temurin:17-jre-alpine
COPY --from=build /src/target/spring-boot-service-*.jar /app.jar
EXPOSE 8081
ENTRYPOINT ["java", "-jar", "/app.jar"]
