//tracing.js
'use strict';
const { NodeTracerProvider } = require('@opentelemetry/node');
const { BatchSpanProcessor } = require("@opentelemetry/tracing");
const { JaegerExporter } = require('@opentelemetry/exporter-jaeger');
const { CollectorTraceExporter } = require('@opentelemetry/exporter-collector-grpc');
const { Resource, SERVICE_RESOURCE } = require('@opentelemetry/resources')
const os = require('os')

const exportType = process.env.EXPORT_TYPE || "OTLP";
const svcname = process.env.SERVICE_NAME || "CurrencyService";
console.log("Exporter type Set To: " + exportType)

const identifier = process.env.HOSTNAME || os.hostname()
const instanceResource = new Resource({
    [SERVICE_RESOURCE.INSTANCE_ID]: identifier,
    [SERVICE_RESOURCE.NAME]: svcname + exportType,
    [SERVICE_RESOURCE.NAMESPACE]: process.env.SERVICE_NAMESPACE,

})
const mergedResource = Resource.createTelemetrySDKResource().merge(instanceResource)

function getExporter(exporterType) {
    switch (exporterType) {
        case 'OTLP':
            console.log("OTLP Set  ")
            return new CollectorTraceExporter({
                // serviceName: "Currency",
                url: process.env.OTLP_ENDPOINT || "http://localhost:55680",
                attributes: {
                         "host.name": process.env.HOST_NAME,
                          "host.ip": process.env.MY_POD_IP,
                          "resource.type":process.env.RESOURCE_TYPE,
                          "service.namespace":process.env.SERVICE_NAMESPACE,
                  }
            })
        case 'JAEGER':
        default:
            console.log("Jaeger Set  ")
            return new JaegerExporter({
                serviceName: process.env.SERVICE_NAME || "currency",
                endpoint: process.env.ENDPOINT,
                username: process.env.USER_NAME,
                password: process.env.PASSWORD

            })
    }
}

const exporter = getExporter(exportType)
const traceProvider = new NodeTracerProvider({
    resource: mergedResource

})
traceProvider.addSpanProcessor(
    new BatchSpanProcessor(exporter)
)

traceProvider.register()

console.log("tracing initialized");