import { createBrowserRouter } from "react-router";
import { PublicLayout } from "@/components/Layout/PublicLayout";
import { LandingPage } from "@/pages/LandingPage";
import { RoutesEnum } from "./Routes";

export const router = createBrowserRouter([
	{
		path: RoutesEnum.root,
		Component: PublicLayout,
		children: [{ index: true, Component: LandingPage }],
	},
]);
