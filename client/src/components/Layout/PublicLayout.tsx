import { Button, Flex, Layout, Typography } from "antd";
import { Content, Footer, Header } from "antd/es/layout/layout";
import { Link, Outlet } from "react-router";
import { RoutesEnum } from "@/routes/Routes";

export const PublicLayout = () => (
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
				<Link to={RoutesEnum.root} style={{ textDecoration: "none" }}>
					<Typography.Text
						style={{ color: "white", fontWeight: "bold", fontSize: 20 }}
					>
						Horizon
					</Typography.Text>
				</Link>
				<Flex gap={8}>
					<Link to={RoutesEnum.login}>
						<Button type="text" style={{ color: "white" }}>
							Login
						</Button>
					</Link>
					<Link to={RoutesEnum.register}>
						<Button type="primary">Register</Button>
					</Link>
				</Flex>
			</Flex>
		</Header>

		<Content>
			<Outlet />
		</Content>

		<Footer style={{ textAlign: "center" }}>
			Horizon © {new Date().getFullYear()}
		</Footer>
	</Layout>
);
