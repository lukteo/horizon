import { createBrowserRouter } from "react-router";
import { PublicLayout } from "@/components/Layout/PublicLayout";
import { AppIndex, AppLayout } from "@/layouts/AppLayout";
import { OrgLayout } from "@/layouts/OrgLayout";
import { AlertsPage } from "@/pages/alerts/AlertsPage";
import { LoginPage } from "@/pages/auth/LoginPage";
import { OnboardingPage } from "@/pages/auth/OnboardingPage";
import { RegisterPage } from "@/pages/auth/RegisterPage";
import { DashboardPage } from "@/pages/dashboard/DashboardPage";
import { IncidentsPage } from "@/pages/incidents/IncidentsPage";
import { LandingPage } from "@/pages/LandingPage";
import { LogsPage } from "@/pages/logs/LogsPage";
import { RulesPage } from "@/pages/rules/RulesPage";
import { SettingsPage } from "@/pages/settings/SettingsPage";
import { SourcesPage } from "@/pages/sources/SourcesPage";

export const router = createBrowserRouter([
	// Public routes
	{
		path: "/",
		Component: PublicLayout,
		children: [
			{ index: true, Component: LandingPage },
			{ path: "login", Component: LoginPage },
			{ path: "register", Component: RegisterPage },
		],
	},

	// Authenticated routes (auth guard in AppLayout)
	{
		Component: AppLayout,
		children: [
			{ path: "onboarding", Component: OnboardingPage },
			{
				path: "app",
				children: [
					// /app with no org — pick first org or go to onboarding
					{ index: true, Component: AppIndex },
					{
						path: ":orgSlug",
						Component: OrgLayout,
						children: [
							{ index: true, Component: DashboardPage },
							{ path: "alerts", Component: AlertsPage },
							{ path: "incidents", Component: IncidentsPage },
							{ path: "logs", Component: LogsPage },
							{ path: "sources", Component: SourcesPage },
							{ path: "rules", Component: RulesPage },
							{ path: "settings", Component: SettingsPage },
						],
					},
				],
			},
		],
	},
]);
