const { NodeSDK } = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');
const resources = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');

// The exporter sends traces to Jaeger via OTLP gRPC
const traceExporter = new OTLPTraceExporter({
    url: 'http://localhost:4317',
});

// Configure the resource to identify this service
const resource = resources.resourceFromAttributes({
    [SemanticResourceAttributes.SERVICE_NAME]: 'node-client',
});

const sdk = new NodeSDK({
    resource: resource,
    traceExporter,
    instrumentations: [getNodeAutoInstrumentations()],
});

// Start the SDK and ensure we shut it down when the process exits
sdk.start();

console.log('OpenTelemetry initialized');

process.on('SIGTERM', () => {
    sdk.shutdown()
        .then(() => console.log('Tracing terminated'))
        .catch((error) => console.log('Error terminating tracing', error))
        .finally(() => process.exit(0));
});

module.exports = sdk;
