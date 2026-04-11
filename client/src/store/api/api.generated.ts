import { api } from "./base";

const injectedRtkApi = api.injectEndpoints({
	endpoints: (build) => ({
		getUsersMe: build.query<GetUsersMeApiResponse, GetUsersMeApiArg>({
			query: () => ({ url: `/users/me` }),
		}),
		updateUsersMe: build.mutation<
			UpdateUsersMeApiResponse,
			UpdateUsersMeApiArg
		>({
			query: (queryArg) => ({
				url: `/users/me`,
				method: "PATCH",
				body: queryArg.updateUserRequest,
			}),
		}),
		listOrganizations: build.query<
			ListOrganizationsApiResponse,
			ListOrganizationsApiArg
		>({
			query: () => ({ url: `/organizations` }),
		}),
		createOrganization: build.mutation<
			CreateOrganizationApiResponse,
			CreateOrganizationApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations`,
				method: "POST",
				body: queryArg.createOrganizationRequest,
			}),
		}),
		getOrganization: build.query<
			GetOrganizationApiResponse,
			GetOrganizationApiArg
		>({
			query: (queryArg) => ({ url: `/organizations/${queryArg.orgId}` }),
		}),
		updateOrganization: build.mutation<
			UpdateOrganizationApiResponse,
			UpdateOrganizationApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}`,
				method: "PATCH",
				body: queryArg.updateOrganizationRequest,
			}),
		}),
		listOrganizationMembers: build.query<
			ListOrganizationMembersApiResponse,
			ListOrganizationMembersApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/members`,
			}),
		}),
		addOrganizationMember: build.mutation<
			AddOrganizationMemberApiResponse,
			AddOrganizationMemberApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/members`,
				method: "POST",
				body: queryArg.addMemberRequest,
			}),
		}),
		updateOrganizationMember: build.mutation<
			UpdateOrganizationMemberApiResponse,
			UpdateOrganizationMemberApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/members/${queryArg.userId}`,
				method: "PATCH",
				body: queryArg.updateMemberRoleRequest,
			}),
		}),
		removeOrganizationMember: build.mutation<
			RemoveOrganizationMemberApiResponse,
			RemoveOrganizationMemberApiArg
		>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/members/${queryArg.userId}`,
				method: "DELETE",
			}),
		}),
		listApiKeys: build.query<ListApiKeysApiResponse, ListApiKeysApiArg>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/api-keys`,
			}),
		}),
		createApiKey: build.mutation<CreateApiKeyApiResponse, CreateApiKeyApiArg>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/api-keys`,
				method: "POST",
				body: queryArg.createApiKeyRequest,
			}),
		}),
		revokeApiKey: build.mutation<RevokeApiKeyApiResponse, RevokeApiKeyApiArg>({
			query: (queryArg) => ({
				url: `/organizations/${queryArg.orgId}/api-keys/${queryArg.keyId}`,
				method: "DELETE",
			}),
		}),
	}),
	overrideExisting: false,
});
export { injectedRtkApi as horizonApi };
export type GetUsersMeApiResponse = /** status 200 OK */ User;
export type GetUsersMeApiArg = void;
export type UpdateUsersMeApiResponse = /** status 200 OK */ User;
export type UpdateUsersMeApiArg = {
	updateUserRequest: UpdateUserRequest;
};
export type ListOrganizationsApiResponse = /** status 200 OK */ Organization[];
export type ListOrganizationsApiArg = void;
export type CreateOrganizationApiResponse =
	/** status 201 Created */ Organization;
export type CreateOrganizationApiArg = {
	createOrganizationRequest: CreateOrganizationRequest;
};
export type GetOrganizationApiResponse = /** status 200 OK */ Organization;
export type GetOrganizationApiArg = {
	orgId: string;
};
export type UpdateOrganizationApiResponse = /** status 200 OK */ Organization;
export type UpdateOrganizationApiArg = {
	orgId: string;
	updateOrganizationRequest: UpdateOrganizationRequest;
};
export type ListOrganizationMembersApiResponse =
	/** status 200 OK */ OrganizationMember[];
export type ListOrganizationMembersApiArg = {
	orgId: string;
};
export type AddOrganizationMemberApiResponse =
	/** status 201 Created */ OrganizationMember;
export type AddOrganizationMemberApiArg = {
	orgId: string;
	addMemberRequest: AddMemberRequest;
};
export type UpdateOrganizationMemberApiResponse =
	/** status 200 OK */ OrganizationMember;
export type UpdateOrganizationMemberApiArg = {
	orgId: string;
	userId: string;
	updateMemberRoleRequest: UpdateMemberRoleRequest;
};
export type RemoveOrganizationMemberApiResponse = unknown;
export type RemoveOrganizationMemberApiArg = {
	orgId: string;
	userId: string;
};
export type ListApiKeysApiResponse = /** status 200 OK */ ApiKey[];
export type ListApiKeysApiArg = {
	orgId: string;
};
export type CreateApiKeyApiResponse = /** status 201 Created */ CreatedApiKey;
export type CreateApiKeyApiArg = {
	orgId: string;
	createApiKeyRequest: CreateApiKeyRequest;
};
export type RevokeApiKeyApiResponse = unknown;
export type RevokeApiKeyApiArg = {
	orgId: string;
	keyId: string;
};
export type BaseEntity = {
	id: string;
	created_at: string;
	updated_at: string;
};
export type User = BaseEntity & {
	email: string;
	first_name?: string | null;
	last_name?: string | null;
	avatar_url?: string | null;
	last_login_at?: string | null;
};
export type ValidationError = {
	/** The JSON path to the field that failed (e.g., "email"). */
	field?: string;
	/** A description of why the field failed validation. */
	message?: string;
};
export type ProblemDetails = {
	/** A URI reference that identifies the problem type. */
	type?: string;
	/** A short, human-readable summary of the problem type. */
	title?: string;
	/** The HTTP status code generated by the origin server. */
	status?: number;
	/** A human-readable explanation specific to this occurrence. */
	detail?: string;
	/** A URI reference that identifies the specific occurrence. */
	instance?: string;
	/** Optional list of individual field errors (common for 400 errors). */
	errors?: ValidationError[];
};
export type UpdateUserRequest = {
	first_name?: string;
	last_name?: string;
};
export type OrgRole = "owner" | "admin" | "analyst" | "viewer";
export type Organization = BaseEntity & {
	name: string;
	slug: string;
	plan: string;
	member_count?: number;
	my_role?: OrgRole;
};
export type CreateOrganizationRequest = {
	name: string;
	/** URL-safe identifier. Auto-generated from name if omitted. */
	slug?: string;
};
export type UpdateOrganizationRequest = {
	name?: string;
};
export type OrganizationMember = BaseEntity & {
	org_id: string;
	user_id: string;
	role: OrgRole;
	user?: User;
};
export type AddMemberRequest = {
	email: string;
	role: OrgRole;
};
export type UpdateMemberRoleRequest = {
	role: OrgRole;
};
export type ApiKey = BaseEntity & {
	org_id: string;
	name: string;
	scopes: string[];
	last_used_at?: string | null;
	revoked_at?: string | null;
};
export type CreatedApiKey = ApiKey & {
	/** The raw API key. Only returned once on creation — store it securely. */
	key: string;
};
export type CreateApiKeyRequest = {
	name: string;
	/** Allowed values: ingest, read */
	scopes: string[];
};
export const {
	useGetUsersMeQuery,
	useLazyGetUsersMeQuery,
	useUpdateUsersMeMutation,
	useListOrganizationsQuery,
	useLazyListOrganizationsQuery,
	useCreateOrganizationMutation,
	useGetOrganizationQuery,
	useLazyGetOrganizationQuery,
	useUpdateOrganizationMutation,
	useListOrganizationMembersQuery,
	useLazyListOrganizationMembersQuery,
	useAddOrganizationMemberMutation,
	useUpdateOrganizationMemberMutation,
	useRemoveOrganizationMemberMutation,
	useListApiKeysQuery,
	useLazyListApiKeysQuery,
	useCreateApiKeyMutation,
	useRevokeApiKeyMutation,
} = injectedRtkApi;
