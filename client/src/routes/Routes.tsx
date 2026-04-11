export const RoutesEnum = {
	catch: "/*",

	// public
	root: "/",
	login: "/login",
	register: "/register",

	// post-auth (no org yet)
	onboarding: "/onboarding",

	// org-scoped absolute path helpers
	orgDashboard: (slug: string) => `/app/${slug}`,
	orgAlerts: (slug: string) => `/app/${slug}/alerts`,
	orgIncidents: (slug: string) => `/app/${slug}/incidents`,
	orgLogs: (slug: string) => `/app/${slug}/logs`,
	orgSources: (slug: string) => `/app/${slug}/sources`,
	orgRules: (slug: string) => `/app/${slug}/rules`,
	orgSettings: (slug: string) => `/app/${slug}/settings`,
	orgMembers: (slug: string) => `/app/${slug}/settings/members`,
	orgApiKeys: (slug: string) => `/app/${slug}/settings/api-keys`,
};
