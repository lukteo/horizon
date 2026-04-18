import {
	AlertOutlined,
	BellOutlined,
	DashboardOutlined,
	DatabaseOutlined,
	FireOutlined,
	SafetyOutlined,
	SearchOutlined,
	SettingOutlined,
	SwapOutlined,
} from "@ant-design/icons";
import { UserButton, useUser } from "@clerk/clerk-react";
import {
	Avatar,
	Dropdown,
	Flex,
	Layout,
	Menu,
	type MenuProps,
	Spin,
	Typography,
	theme,
} from "antd";
import { Content, Header } from "antd/es/layout/layout";
import Sider from "antd/es/layout/Sider";
import { useMemo, useState } from "react";
import {
	Navigate,
	Outlet,
	useLocation,
	useNavigate,
	useParams,
} from "react-router";
import { OrgContext } from "@/contexts/OrgContext";
import { RoutesEnum } from "@/routes/Routes";
import { useListOrganizationsQuery } from "@/store/api/api.generated";

const NAV_ITEMS: MenuProps["items"] = [
	{
		key: "dashboard",
		icon: <DashboardOutlined />,
		label: "Dashboard",
	},
	{
		key: "alerts",
		icon: <BellOutlined />,
		label: "Alerts",
	},
	{
		key: "incidents",
		icon: <FireOutlined />,
		label: "Incidents",
	},
	{
		key: "logs",
		icon: <SearchOutlined />,
		label: "Log Explorer",
	},
	{
		key: "sources",
		icon: <DatabaseOutlined />,
		label: "Sources",
	},
	{
		key: "rules",
		icon: <SafetyOutlined />,
		label: "Detection Rules",
	},
];

const SETTINGS_ITEMS: MenuProps["items"] = [
	{
		key: "settings",
		icon: <SettingOutlined />,
		label: "Settings",
	},
];

export const OrgLayout = () => {
	const { orgSlug } = useParams<{ orgSlug: string }>();
	const location = useLocation();
	const navigate = useNavigate();
	const { user } = useUser();
	const [collapsed, setCollapsed] = useState(false);
	const {
		token: { colorBgContainer },
	} = theme.useToken();

	const { data: orgs, isLoading } = useListOrganizationsQuery();

	const org = useMemo(
		() => orgs?.find((o) => o.slug === orgSlug),
		[orgs, orgSlug],
	);

	// Derive selected menu key from the current path
	const selectedKey = useMemo(() => {
		const segments = location.pathname.split("/");
		// segments: ["", "app", orgSlug, key]
		const key = segments[3];
		if (!key) return "dashboard";
		if (key === "settings") return "settings";
		return key;
	}, [location.pathname]);

	if (isLoading) {
		return <Spin fullscreen />;
	}

	if (!org) {
		// Org slug not found among user's orgs — redirect to onboarding
		return <Navigate to={RoutesEnum.onboarding} replace />;
	}

	const handleNavClick: MenuProps["onClick"] = ({ key }) => {
		if (key === "dashboard") {
			navigate(RoutesEnum.orgDashboard(org.slug));
		} else {
			navigate(`/app/${org.slug}/${key}`);
		}
	};

	const handleOrgSwitch: MenuProps["onClick"] = ({ key }) => {
		navigate(RoutesEnum.orgDashboard(key));
	};

	const orgDropdownItems: MenuProps["items"] = [
		...(orgs ?? []).map((o) => ({
			key: o.slug,
			label: o.name,
			disabled: o.slug === org.slug,
			icon: (
				<Avatar size="small" style={{ backgroundColor: "#1677ff" }}>
					{o.name.charAt(0).toUpperCase()}
				</Avatar>
			),
		})),
		{ type: "divider" as const },
		{
			key: "__new",
			icon: <AlertOutlined />,
			label: "Create new organisation",
			onClick: () => navigate(RoutesEnum.onboarding),
		},
	];

	return (
		<OrgContext.Provider value={{ org }}>
			<Layout style={{ minHeight: "100vh" }}>
				<Sider
					collapsible
					collapsed={collapsed}
					onCollapse={setCollapsed}
					style={{ background: colorBgContainer }}
					width={220}
				>
					{/* Brand + org switcher */}
					<Dropdown
						menu={{ items: orgDropdownItems, onClick: handleOrgSwitch }}
						trigger={["click"]}
					>
						<Flex
							align="center"
							gap={8}
							style={{
								padding: collapsed ? "16px 8px" : "16px",
								cursor: "pointer",
								borderBottom: "1px solid rgba(0,0,0,0.06)",
								marginBottom: 8,
							}}
						>
							<Avatar
								size="small"
								style={{ backgroundColor: "#1677ff", flexShrink: 0 }}
							>
								{org.name.charAt(0).toUpperCase()}
							</Avatar>
							{!collapsed && (
								<Flex
									align="center"
									justify="space-between"
									style={{ flex: 1, minWidth: 0 }}
								>
									<Typography.Text strong ellipsis style={{ flex: 1 }}>
										{org.name}
									</Typography.Text>
									<SwapOutlined
										style={{
											fontSize: 12,
											color: "rgba(0,0,0,0.45)",
											flexShrink: 0,
										}}
									/>
								</Flex>
							)}
						</Flex>
					</Dropdown>

					{/* Main nav */}
					<Menu
						mode="inline"
						selectedKeys={[selectedKey]}
						items={NAV_ITEMS}
						onClick={handleNavClick}
						style={{ border: "none" }}
					/>

					{/* Bottom nav */}
					<Menu
						mode="inline"
						selectedKeys={[selectedKey]}
						items={SETTINGS_ITEMS}
						onClick={handleNavClick}
						style={{
							border: "none",
							position: "absolute",
							bottom: 48,
							width: "100%",
						}}
					/>
				</Sider>

				<Layout>
					<Header
						style={{
							background: colorBgContainer,
							padding: "0 24px",
							display: "flex",
							alignItems: "center",
							justifyContent: "flex-end",
							borderBottom: "1px solid rgba(0,0,0,0.06)",
						}}
					>
						<Flex align="center" gap={12}>
							{user && (
								<Typography.Text type="secondary" style={{ fontSize: 13 }}>
									{user.primaryEmailAddress?.emailAddress}
								</Typography.Text>
							)}
							<UserButton />
						</Flex>
					</Header>

					<Content style={{ margin: 24, minHeight: 280 }}>
						<Outlet />
					</Content>
				</Layout>
			</Layout>
		</OrgContext.Provider>
	);
};
