import type React from "react";

export default async function AuthLayout({ children }: { children: React.ReactNode }) {
	return (
		<div className="flex-1">{children}</div>
	);
}
