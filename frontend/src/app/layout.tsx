import type { Metadata } from "next";
import { Nunito, Changa_One } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";
import { AppProvider } from "@/contexts/app";

const nunito = Nunito({ subsets: ["latin", "vietnamese"] });
const changaOne = Changa_One({ subsets: ["latin"], weight: "400", variable: "--font-changa-one" });

export const metadata: Metadata = {
	title: "AnoChat",
	description: "Real-time chat application",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<head>
				<script
					dangerouslySetInnerHTML={{
						__html: `try{var t=localStorage.getItem('theme');if(t==='dark')document.documentElement.classList.add('dark');else if(t==='pink')document.documentElement.classList.add('pink');}catch(e){}`,
					}}
				/>
			</head>
			<body className={`${nunito.className} ${changaOne.variable}`}>
				<AppProvider>
					{children}
					<Toaster />
				</AppProvider>
			</body>
		</html>
	);
}
