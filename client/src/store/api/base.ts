import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

// The Horizon API base URL — override via VITE_API_URL in .env
const BASE_URL =
	(import.meta.env.VITE_API_URL as string | undefined) ??
	"http://localhost:8080/api/v1";

export const api = createApi({
	reducerPath: "horizonApi",
	baseQuery: fetchBaseQuery({
		baseUrl: BASE_URL,
		prepareHeaders: async (headers) => {
			// Inject Clerk session token if available.
			// window.Clerk is set by the Clerk provider before any API call is made.
			const clerk = (
				window as unknown as {
					Clerk?: { session?: { getToken: () => Promise<string | null> } };
				}
			).Clerk;
			const token = await clerk?.session?.getToken();
			if (token) {
				headers.set("Authorization", `Bearer ${token}`);
			}
			return headers;
		},
	}),
	endpoints: () => ({}),
});
