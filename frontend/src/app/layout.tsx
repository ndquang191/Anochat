import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";
import { AppProvider } from "@/contexts/app";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
	title: "AnoChat",
	description: "Real-time chat application",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en">
			<body className={inter.className}>
				<AppProvider>
					{children}
					<Toaster />
				</AppProvider>
			</body>
		</html>
	);
}
