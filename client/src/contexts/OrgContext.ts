import { createContext, useContext } from "react";
import type { Organization } from "@/store/api/api.generated";

type OrgContextType = {
	org: Organization;
};

export const OrgContext = createContext<OrgContextType | null>(null);

export const useOrg = (): OrgContextType => {
	const ctx = useContext(OrgContext);
	if (!ctx) {
		throw new Error("useOrg must be used within OrgLayout");
	}
	return ctx;
};
