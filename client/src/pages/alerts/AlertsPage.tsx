import { Empty, Typography } from "antd";

export const AlertsPage = () => (
	<div>
		<Typography.Title level={4} style={{ marginBottom: 24 }}>
			Alerts
		</Typography.Title>
		<Empty description="No alerts yet — detection rules will surface alerts here once logs are flowing." />
	</div>
);
