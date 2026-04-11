import { useAuth } from "@clerk/clerk-react";
import { Spin } from "antd";
import { Navigate, Outlet } from "react-router";
import { RoutesEnum } from "@/routes/Routes";
import { useListOrganizationsQuery } from "@/store/api/api.generated";

// AppIndex: after auth check, redirect to first org or onboarding
const AppIndex = () => {
	const { data: orgs, isLoading } = useListOrganizationsQuery();

	if (isLoading) {
		return <Spin fullscreen />;
	}

	if (!orgs || orgs.length === 0) {
		return <Navigate to={RoutesEnum.onboarding} replace />;
	}

	return <Navigate to={RoutesEnum.orgDashboard(orgs[0].slug)} replace />;
};

export { AppIndex };

// AppLayout: auth guard — redirects to /login if not authenticated
export const AppLayout = () => {
	const { isLoaded, isSignedIn } = useAuth();

	if (!isLoaded) {
		return <Spin fullscreen />;
	}

	if (!isSignedIn) {
		return <Navigate to={RoutesEnum.login} replace />;
	}

	return <Outlet />;
};
