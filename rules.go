package yamlfmt

// DefaultOpenAPIRules contains an opinionated ordering of an 'openapi.yaml' file based on the tables
// documented on https://swagger.io/specification/#schema-1
func DefaultOpenAPIRules() []Rule {
	operationFn := NewSimpleOrdering("tags", "summary", "description", "externalDocs", "operationId", "parameters", "requestBody", "responses", "callbacks", "deprecated", "security", "servers")
	mediaTypeFn := NewSimpleOrdering("schema", "example", "examples", "encoding")
	schemaFn := NewSimpleOrdering("title", "type", "format", "required", "oneOf", "anyOf", "allOf", "properties", "additionalProperties")
	return []Rule{
		NewRule("$", StringOrderingFn, NewSimpleOrdering("openapi", "info", "jsonSchemaDialect", "servers", "paths", "webhooks", "components", "security", "tags", "externalDocs")),
		NewRule("$.info", StringOrderingFn, NewSimpleOrdering("title", "summary", "description", "termsOfService", "contact", "license", "version")),
		NewRule("$.info.contact", StringOrderingFn, NewSimpleOrdering("name", "url", "email")),
		NewRule("$.info.license", StringOrderingFn, NewSimpleOrdering("name", "identifier", "url")),
		NewRule("$.servers[*]", StringOrderingFn, NewSimpleOrdering("url", "description", "variables")),
		NewRule("$.servers[*].variables", StringOrderingFn),
		NewRule("$.servers[*].variables[*]", StringOrderingFn, NewSimpleOrdering("enum", "default", "description")),
		NewRule("$.components", StringOrderingFn, NewSimpleOrdering("schemas", "responses", "parameters", "examples", "requestBodies", "headers", "securitySchemes", "links", "callbacks", "pathItems")),
		NewRule("$.paths", StringOrderingFn),
		NewRule("$.paths[*]", StringOrderingFn, NewSimpleOrdering("$ref", "summary", "description", "get", "put", "post", "delete", "options", "head", "patch", "trace", "servers", "parameters")),
		NewRule("$.paths[*].get", StringOrderingFn, operationFn),
		NewRule("$.paths[*].put", StringOrderingFn, operationFn),
		NewRule("$.paths[*].post", StringOrderingFn, operationFn),
		NewRule("$.paths[*].delete", StringOrderingFn, operationFn),
		NewRule("$.paths[*].options", StringOrderingFn, operationFn),
		NewRule("$.paths[*].head", StringOrderingFn, operationFn),
		NewRule("$.paths[*].patch", StringOrderingFn, operationFn),
		NewRule("$.paths[*].trace", StringOrderingFn, operationFn),
		NewRule("$.paths[*][*].externalDocs", StringOrderingFn, NewSimpleOrdering("description", "url")),
		NewRule("$.paths[*][*].parameters[*]", StringOrderingFn, NewSimpleOrdering("$ref", "name", "in", "description", "required", "deprecated", "allowEmptyValue", "style", "explode", "allowReserved", "schema")),
		NewRule("$.paths[*][*].requestBody", StringOrderingFn, NewSimpleOrdering("description", "content", "required")),
		NewRule("$.paths[*][*].requestBody.content", StringOrderingFn),
		NewRule("$.paths[*][*].requestBody.content[*]", StringOrderingFn, mediaTypeFn),
		NewRule("$.paths[*][*].responses", StringOrderingFn),
		NewRule("$.paths[*][*].responses[*]", StringOrderingFn, NewSimpleOrdering("description", "headers", "content", "links")),
		NewRule("$.paths[*][*].responses[*].content", StringOrderingFn, mediaTypeFn),
		NewRule(".schema", StringOrderingFn, schemaFn),
		NewRule("$.components.schemas[*]", StringOrderingFn, schemaFn),
		NewRule(".schema.properties", StringOrderingFn),
		NewRule(".schemas[*].properties", StringOrderingFn),
	}
}
