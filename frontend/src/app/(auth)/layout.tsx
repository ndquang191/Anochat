import type React from "react";
import "../globals.css"; // Ensure your global CSS is imported

export default async function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en">
			<body>
				<div className="flex-1">{children}</div>
			</body>
		</html>
	);
}
