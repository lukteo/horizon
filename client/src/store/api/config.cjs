/** @type {import('@rtk-query/codegen-openapi').ConfigFile} */
const config = {
	schemaFile: "../../../../openapi/openapi.yml",
	apiFile: "./base.ts",
	apiImportSpecifier: "@/store/api/base",
	outputFile: "./api.generated.ts",
	exportName: "horizonApi",
	hooks: {
		queries: true,
		lazyQueries: true,
		mutations: true,
	},
};

module.exports = config;
