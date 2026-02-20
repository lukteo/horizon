import { Breadcrumb, Button, Flex, Layout, Typography, theme } from "antd";
import { Content, Footer, Header } from "antd/es/layout/layout";

export const PublicLayout = () => {
	const {
		token: { colorBgContainer, borderRadiusLG },
	} = theme.useToken();

	return (
		<Layout>
			<Header
				style={{
					position: "sticky",
					top: 0,
					zIndex: 1,
					width: "100%",
					display: "flex",
					alignItems: "center",
				}}
			>
				<Flex justify="space-between" style={{ width: "100%" }}>
					<Typography
						style={{ color: "white", fontWeight: "bold", fontSize: 20 }}
					>
						Horizon
					</Typography>
					<Flex gap={8}>
						<Button type="text" style={{ color: "white" }}>
							Login
						</Button>
						<Button type="primary">Register</Button>
					</Flex>
				</Flex>
			</Header>
			<Content style={{ padding: "0 48px" }}>
				<Breadcrumb
					style={{ margin: "16px 0" }}
					items={[{ title: "Home" }, { title: "List" }, { title: "App" }]}
				/>
				<div
					style={{
						padding: 24,
						minHeight: 380,
						background: colorBgContainer,
						borderRadius: borderRadiusLG,
					}}
				>
					Content something is here
				</div>
			</Content>
			<Footer style={{ textAlign: "center" }}>
				Ant Design ©{new Date().getFullYear()} Created by Ant UED
			</Footer>
		</Layout>
	);
};
