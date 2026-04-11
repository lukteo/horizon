import { useAuth } from "@clerk/clerk-react";
import { Button, Card, Form, Input, Spin, Typography } from "antd";
import { useEffect } from "react";
import { Navigate, useNavigate } from "react-router";
import { RoutesEnum } from "@/routes/Routes";
import {
	useCreateOrganizationMutation,
	useListOrganizationsQuery,
} from "@/store/api/api.generated";

type FormValues = {
	name: string;
	slug: string;
};

const slugify = (value: string) =>
	value
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "")
		.slice(0, 100);

export const OnboardingPage = () => {
	const { isLoaded, isSignedIn } = useAuth();
	const navigate = useNavigate();
	const [form] = Form.useForm<FormValues>();
	const [createOrg, { isLoading: isCreating }] =
		useCreateOrganizationMutation();
	const { data: orgs, isLoading: isLoadingOrgs } = useListOrganizationsQuery();

	// If user already has an org, skip onboarding
	useEffect(() => {
		if (!isLoadingOrgs && orgs && orgs.length > 0) {
			navigate(RoutesEnum.orgDashboard(orgs[0].slug), { replace: true });
		}
	}, [orgs, isLoadingOrgs, navigate]);

	if (!isLoaded || isLoadingOrgs) {
		return <Spin fullscreen />;
	}

	if (!isSignedIn) {
		return <Navigate to={RoutesEnum.login} replace />;
	}

	const handleSubmit = async (values: FormValues) => {
		try {
			const org = await createOrg({
				createOrganizationRequest: {
					name: values.name,
					slug: values.slug || slugify(values.name),
				},
			}).unwrap();
			navigate(RoutesEnum.orgDashboard(org.slug), { replace: true });
		} catch {
			// errors handled by Form.Item via setFields or server-side
		}
	};

	const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const currentSlug = form.getFieldValue("slug") as string;
		// Only auto-fill slug if user hasn't manually edited it
		if (
			!currentSlug ||
			currentSlug === slugify(form.getFieldValue("name") ?? "")
		) {
			form.setFieldValue("slug", slugify(e.target.value));
		}
	};

	return (
		<div
			style={{
				minHeight: "100vh",
				display: "flex",
				alignItems: "center",
				justifyContent: "center",
				padding: 24,
				background: "#f5f5f5",
			}}
		>
			<Card style={{ width: "100%", maxWidth: 440 }}>
				<Typography.Title level={3} style={{ marginBottom: 4 }}>
					Create your organisation
				</Typography.Title>
				<Typography.Text
					type="secondary"
					style={{ display: "block", marginBottom: 24 }}
				>
					Organisations are the top-level workspace in Horizon. You can invite
					team members and manage log sources within an organisation.
				</Typography.Text>

				<Form
					form={form}
					layout="vertical"
					onFinish={handleSubmit}
					requiredMark={false}
				>
					<Form.Item
						name="name"
						label="Organisation name"
						rules={[
							{ required: true, message: "Name is required" },
							{ min: 2, message: "At least 2 characters" },
						]}
					>
						<Input
							placeholder="Acme Security"
							onChange={handleNameChange}
							autoFocus
						/>
					</Form.Item>

					<Form.Item
						name="slug"
						label="URL identifier"
						rules={[
							{ required: true, message: "Slug is required" },
							{
								pattern: /^[a-z0-9-]+$/,
								message: "Lowercase letters, numbers, and hyphens only",
							},
							{ min: 2, message: "At least 2 characters" },
						]}
						extra="Used in URLs — e.g. /app/acme-security"
					>
						<Input placeholder="acme-security" />
					</Form.Item>

					<Form.Item style={{ marginBottom: 0 }}>
						<Button type="primary" htmlType="submit" loading={isCreating} block>
							Create organisation
						</Button>
					</Form.Item>
				</Form>
			</Card>
		</div>
	);
};
