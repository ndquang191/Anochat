"use client";

import { LoginForm } from "@/components/login-form";
import { Particles } from "@/components/ui/particles";
import { useTheme } from "@/contexts/theme";

export default function Page() {
	const { theme } = useTheme();
	const particleColor = theme === "dark" ? "#ffffff" : "#000000";

	return (
		<div className="relative flex min-h-svh w-full items-center justify-center p-6 md:p-10 overflow-hidden">
			<Particles
				className="absolute inset-0"
				quantity={120}
				staticity={40}
				ease={60}
				size={0.5}
				color={particleColor}
			/>
			<div className="relative z-10 w-full max-w-sm">
				<LoginForm />
			</div>
		</div>
	);
}
