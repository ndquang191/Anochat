import type { Metadata } from "next";
import { Nunito, Changa_One } from "next/font/google";
import { cookies } from "next/headers";
import "./globals.css";
import { Toaster } from "@/components/ui/sonner";
import { AppProvider } from "@/contexts/app";
import type { Language } from "@/lib/i18n";
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

export default async function RootLayout({ children }: { children: React.ReactNode }) {
	const cookieStore = await cookies();
	const languageCookie = cookieStore.get("language")?.value;
	const hasLanguageCookie = languageCookie === "vi" || languageCookie === "en";
	const initialLanguage: Language = hasLanguageCookie ? languageCookie : "vi";

	return (
		<html lang={initialLanguage} suppressHydrationWarning>
			<head>
				<script
					dangerouslySetInnerHTML={{
						__html: `try{var d=document.documentElement;var t=localStorage.getItem('theme');if(t==='dark')d.classList.add('dark');else if(t==='pink')d.classList.add('pink');var l=localStorage.getItem('language');var h=${hasLanguageCookie};var s='${initialLanguage}';if(!h&&(l==='vi'||l==='en')&&l!==s){document.cookie='language='+l+';max-age=31536000;path=/;SameSite=Lax';var w=document.cookie.split('; ').some(function(c){return c==='language='+l});if(w){d.style.visibility='hidden';location.reload();}else{localStorage.setItem('language',s);}}else{if(!h)document.cookie='language='+s+';max-age=31536000;path=/;SameSite=Lax';localStorage.setItem('language',s);}}catch(e){}`,
					}}
				/>
			</head>
			<body className={`${nunito.className} ${changaOne.variable}`}>
				<AppProvider initialLanguage={initialLanguage}>
					{children}
					<Toaster />
				</AppProvider>
			</body>
		</html>
	);
}
