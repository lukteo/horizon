import { SignIn } from "@clerk/clerk-react";
import { Flex } from "antd";

export const LoginPage = () => (
	<Flex justify="center" style={{ padding: "48px 24px" }}>
		<SignIn
			routing="path"
			path="/login"
			signUpUrl="/register"
			forceRedirectUrl="/app"
		/>
	</Flex>
);
