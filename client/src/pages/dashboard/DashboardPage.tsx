import { Card, Col, Row, Statistic, Typography } from "antd";
import { useOrg } from "@/contexts/OrgContext";

export const DashboardPage = () => {
	const { org } = useOrg();

	return (
		<div>
			<Typography.Title level={4} style={{ marginBottom: 24 }}>
				Dashboard
			</Typography.Title>

			<Row gutter={[16, 16]}>
				<Col xs={24} sm={12} lg={6}>
					<Card>
						<Statistic title="Open Alerts" value={0} />
					</Card>
				</Col>
				<Col xs={24} sm={12} lg={6}>
					<Card>
						<Statistic title="Open Incidents" value={0} />
					</Card>
				</Col>
				<Col xs={24} sm={12} lg={6}>
					<Card>
						<Statistic title="Log Sources" value={0} />
					</Card>
				</Col>
				<Col xs={24} sm={12} lg={6}>
					<Card>
						<Statistic title="Detection Rules" value={0} />
					</Card>
				</Col>
			</Row>

			<Typography.Text
				type="secondary"
				style={{ display: "block", marginTop: 32 }}
			>
				Organisation: <strong>{org.name}</strong> · Plan: {org.plan}
			</Typography.Text>
		</div>
	);
};
