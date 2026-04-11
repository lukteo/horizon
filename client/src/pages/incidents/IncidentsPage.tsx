import { Empty, Typography } from "antd";

export const IncidentsPage = () => (
	<div>
		<Typography.Title level={4} style={{ marginBottom: 24 }}>
			Incidents
		</Typography.Title>
		<Empty description="No incidents yet — create an incident to start tracking an investigation." />
	</div>
);
