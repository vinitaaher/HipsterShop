FROM adoptopenjdk/openjdk11 as builder
#ENV EXPORTER=jaeger
#ENV EXPORTER_ENDPOINT=jaeger:14250

WORKDIR /usr/app
COPY ./build/libs/hipstershop-0.1.0-SNAPSHOT-fat.jar /usr/app/
#COPY ./tracinglib/opentelemetry-javaagent-all.jar /usr/app/

RUN curl -LJO https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v0.8.0/opentelemetry-javaagent-all.jar

WORKDIR /usr/app
EXPOSE 9555

#ENTRYPOINT ["java","-javaagent:opentelemetry-javaagent-all.jar","-Dotel.exporter=jaeger","-Dotel.exporter.jaeger.service.name=AdService","-Dotel.exporter.jaeger.endpoint=jaeger:14250","-jar", "hipstershop-0.1.0-SNAPSHOT-fat.jar"]
ARG EXPORTER
ENV EXPORTER={EXPORT_TYPE}
ARG EXPORTER_ENDPOINT
ENV EXPORT_TYPE={EXPORT_TYPE}
ARG SERVICE_NAME
ENV SERVICE_NAME={SERVICE_NAME}
ARG HOST_NAME
ENV HOST_NAME={HOST_NAME}
ARG SERVICE_NAMESPACE
ENV SERVICE_NAMESPACE={SERVICE_NAMESPACE}
ARG MY_POD_IP
ENV MY_POD_IP={MY_POD_IP}
ARG RESOURCE_TYPE
ENV RESOURCE_TYPE={RESOURCE_TYPE}
ENTRYPOINT java -javaagent:opentelemetry-javaagent-all.jar -Dotel.exporter=$EXPORTER -Dotel.resource.attributes=service.name=$SERVICE_NAME,host.name=$HOST_NAME,service.namespace=$SERVICE_NAMESPACE,resource.type=$RESOURCE_TYPE,ip=$MY_POD_IP -Dotel.exporter.$EXPORTER.endpoint=$EXPORTER_ENDPOINT -Dotel.exporter.otlp.insecure=true -jar hipstershop-0.1.0-SNAPSHOT-fat.jar

