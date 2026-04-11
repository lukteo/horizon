import { SignUp } from "@clerk/clerk-react";
import { Flex } from "antd";

export const RegisterPage = () => (
	<Flex justify="center" style={{ padding: "48px 24px" }}>
		<SignUp
			routing="path"
			path="/register"
			signInUrl="/login"
			forceRedirectUrl="/onboarding"
		/>
	</Flex>
);
