import { KeyOutlined, TeamOutlined } from "@ant-design/icons";
import { Card, Col, Row, Typography } from "antd";
import { useNavigate } from "react-router";
import { useOrg } from "@/contexts/OrgContext";
import { RoutesEnum } from "@/routes/Routes";

export const SettingsPage = () => {
	const { org } = useOrg();
	const navigate = useNavigate();

	return (
		<div>
			<Typography.Title level={4} style={{ marginBottom: 8 }}>
				Settings
			</Typography.Title>
			<Typography.Text
				type="secondary"
				style={{ display: "block", marginBottom: 24 }}
			>
				Manage settings for <strong>{org.name}</strong>
			</Typography.Text>

			<Row gutter={[16, 16]}>
				<Col xs={24} sm={12} lg={8}>
					<Card
						hoverable
						onClick={() => navigate(RoutesEnum.orgMembers(org.slug))}
					>
						<TeamOutlined style={{ fontSize: 24, marginBottom: 8 }} />
						<Typography.Title level={5} style={{ margin: 0 }}>
							Members
						</Typography.Title>
						<Typography.Text type="secondary">
							Manage team members and roles
						</Typography.Text>
					</Card>
				</Col>
				<Col xs={24} sm={12} lg={8}>
					<Card
						hoverable
						onClick={() => navigate(RoutesEnum.orgApiKeys(org.slug))}
					>
						<KeyOutlined style={{ fontSize: 24, marginBottom: 8 }} />
						<Typography.Title level={5} style={{ margin: 0 }}>
							API Keys
						</Typography.Title>
						<Typography.Text type="secondary">
							Create and manage API keys for programmatic access
						</Typography.Text>
					</Card>
				</Col>
			</Row>
		</div>
	);
};
