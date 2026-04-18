import { Empty, Typography } from "antd";

export const SourcesPage = () => (
	<div>
		<Typography.Title level={4} style={{ marginBottom: 24 }}>
			Log Sources
		</Typography.Title>
		<Empty description="No log sources yet — sources will be configurable in Phase 2." />
	</div>
);
