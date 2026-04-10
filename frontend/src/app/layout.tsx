import type { Metadata } from "next";
import { Nunito, Changa_One } from "next/font/google";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";
import { AppProvider } from "@/contexts/app";
import { SITE_DESCRIPTION, SITE_KEYWORDS, SITE_NAME, SITE_TAGLINE, absoluteUrl, getMetadataBase } from "@/lib/site";

const nunito = Nunito({ subsets: ["latin", "vietnamese"] });
const changaOne = Changa_One({ subsets: ["latin"], weight: "400", variable: "--font-changa-one" });

export const metadata: Metadata = {
	metadataBase: getMetadataBase(),
	title: {
		default: `${SITE_NAME} | ${SITE_TAGLINE}`,
		template: `%s | ${SITE_NAME}`,
	},
	description: SITE_DESCRIPTION,
	applicationName: SITE_NAME,
	keywords: SITE_KEYWORDS,
	openGraph: {
		title: `${SITE_NAME} | ${SITE_TAGLINE}`,
		description: SITE_DESCRIPTION,
		siteName: SITE_NAME,
		type: "website",
		locale: "vi_VN",
		images: [
			{
				url: absoluteUrl("/icon.svg"),
				alt: SITE_NAME,
			},
		],
	},
	twitter: {
		card: "summary",
		title: `${SITE_NAME} | ${SITE_TAGLINE}`,
		description: SITE_DESCRIPTION,
		images: [absoluteUrl("/icon.svg")],
	},
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="vi" suppressHydrationWarning>
			<head>
				<script
					dangerouslySetInnerHTML={{
						__html: `try{var d=document.documentElement;var t=localStorage.getItem('theme');var l=localStorage.getItem('language');if(t==='dark')d.classList.add('dark');else if(t==='pink')d.classList.add('pink');if(l==='vi'||l==='en')d.lang=l;}catch(e){}`,
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
