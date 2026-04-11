import { Empty, Typography } from "antd";

export const RulesPage = () => (
	<div>
		<Typography.Title level={4} style={{ marginBottom: 24 }}>
			Detection Rules
		</Typography.Title>
		<Empty description="No detection rules yet — Sigma rule management will be available in Phase 4." />
	</div>
);
