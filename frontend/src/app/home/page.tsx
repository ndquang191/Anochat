import type { Metadata } from "next";
import Link from "next/link";
import type { LucideIcon } from "lucide-react";
import { ArrowRight, Globe2, MessageCircle, Search, Shield, Sparkles } from "lucide-react";
import { AuroraText } from "@/components/aurora-text";
import { HomeCtaActions } from "@/components/home/home-cta-actions";
import { ShineBorder } from "@/components/ui/shine-border";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { SITE_DESCRIPTION, SITE_NAME, SITE_TAGLINE, absoluteUrl } from "@/lib/site";

const pageTitle = "Chat ẩn danh với người lạ tại Việt Nam";
const pageDescription =
	"AnoChat là landing page cho nhu cầu anonymous chat VN: kết nối ngẫu nhiên, nói chuyện với người lạ, giữ trải nghiệm riêng tư và vào phòng chat rất nhanh.";

const highlights: Array<{ icon: LucideIcon; title: string; description: string }> = [
	{
		icon: Shield,
		title: "Ẩn danh",
		description: "Không cần mở hồ sơ dài để bắt đầu.",
	},
	{
		icon: MessageCircle,
		title: "Nhanh",
		description: "Vào hàng đợi và chat gần như ngay lập tức.",
	},
	{
		icon: Globe2,
		title: "Đúng intent",
		description: "Viết cho anonymous chat VN và random chat Việt Nam.",
	},
	{
		icon: Search,
		title: "Ngẫu nhiên",
		description: "Tập trung vào gặp người lạ thay vì lướt profile.",
	},
];

const steps = [
	{
		label: "01",
		title: "Đăng nhập",
		description: "Google login dùng để xác thực phiên, không biến sản phẩm thành mạng xã hội.",
	},
	{
		label: "02",
		title: "Vào hàng đợi",
		description: "Hệ thống ghép ngẫu nhiên để bạn bắt đầu anonymous chat nhanh.",
	},
	{
		label: "03",
		title: "Chat hoặc rời phòng",
		description: "Nếu không hợp, bạn có thể thoát nhanh và ghép lại.",
	},
];

const faqs = [
	{
		question: "AnoChat dành cho ai?",
		answer: "Cho người muốn chat ẩn danh, nói chuyện với người lạ hoặc thử random chat Việt Nam mà không cần mở profile công khai.",
	},
	{
		question: "Route nào được index?",
		answer: "Trang public chính là /home. Route chat và auth được gắn noindex để không cạnh tranh crawl.",
	},
	{
		question: "Nếu đã đăng nhập?",
		answer: "CTA trên landing sẽ đưa bạn thẳng về / để tiếp tục chat thay vì lặp lại flow.",
	},
	{
		question: "Điểm khác với app chat thường?",
		answer: "AnoChat ưu tiên hội thoại 1-1 nhanh, gọn, ít ma sát hơn so với các app nặng profile và feed.",
	},
];

const keywordPills = ["Anonymous chat VN", "Chat ẩn danh", "Nói chuyện với người lạ", "Random chat Việt Nam"];

export const metadata: Metadata = {
	title: pageTitle,
	description: pageDescription,
	keywords: [
		"anonymous chat VN",
		"chat ẩn danh",
		"chat với người lạ",
		"nói chuyện với người lạ",
		"random chat Việt Nam",
	],
	alternates: {
		canonical: "/home",
	},
	robots: {
		index: true,
		follow: true,
	},
	openGraph: {
		title: `${pageTitle} | ${SITE_NAME}`,
		description: pageDescription,
		url: absoluteUrl("/home"),
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
		title: `${pageTitle} | ${SITE_NAME}`,
		description: pageDescription,
		images: [absoluteUrl("/icon.svg")],
	},
};

export const revalidate = 86400;

export default function HomePage() {
	const structuredData = [
		{
			"@context": "https://schema.org",
			"@type": "WebSite",
			name: SITE_NAME,
			url: absoluteUrl("/"),
			description: pageDescription,
			inLanguage: "vi-VN",
		},
		{
			"@context": "https://schema.org",
			"@type": "Organization",
			name: SITE_NAME,
			url: absoluteUrl("/"),
			logo: absoluteUrl("/icon.svg"),
			description: SITE_DESCRIPTION,
		},
		{
			"@context": "https://schema.org",
			"@type": "FAQPage",
			mainEntity: faqs.map((faq) => ({
				"@type": "Question",
				name: faq.question,
				acceptedAnswer: {
					"@type": "Answer",
					text: faq.answer,
				},
			})),
		},
	];

	return (
		<main className="relative min-h-screen bg-background text-foreground lg:h-screen lg:w-screen lg:overflow-hidden">
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
			/>

			<div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(81,107,145,0.18),_transparent_34%),radial-gradient(circle_at_88%_14%,_rgba(89,196,230,0.14),_transparent_24%),radial-gradient(circle_at_50%_100%,_rgba(203,176,227,0.14),_transparent_30%)]" />

			<div className="mx-auto flex min-h-screen w-full max-w-[100vw] px-3 py-3 sm:px-4 sm:py-4 lg:h-screen lg:px-4 lg:py-4">
				<div className="grid w-full gap-3 lg:h-full lg:grid-cols-[1.12fr_0.88fr] lg:grid-rows-1">
					<section className="relative overflow-hidden rounded-[1.75rem] border border-border/70 bg-card/90 shadow-[0_18px_60px_rgba(15,23,42,0.08)] backdrop-blur lg:h-full">
						<ShineBorder
							borderWidth={1}
							duration={12}
							shineColor={["rgba(81,107,145,0.12)", "rgba(147,183,227,0.42)", "rgba(89,196,230,0.16)"]}
						/>
						<div className="flex h-full min-h-[calc(100svh-1.5rem)] flex-col p-4 sm:p-5 lg:min-h-0 lg:p-5 xl:p-6">
							<div className="mb-3 inline-flex w-fit items-center gap-2 rounded-full border border-border/80 bg-background/85 px-3 py-1.5 text-xs text-muted-foreground shadow-sm">
								<Sparkles className="size-3.5 text-primary" />
								<span>{SITE_TAGLINE}</span>
							</div>

							<div className="space-y-3">
								<p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-primary/80 sm:text-xs">
									Anonymous Chat VN
								</p>
								<h1 className="max-w-4xl text-[2rem] font-black leading-[0.95] tracking-tight sm:text-[2.6rem] lg:text-[3rem] xl:text-[3.5rem]">
									Chat ẩn danh với người lạ, gói gọn trong đúng một màn hình.
								</h1>
								<p className="max-w-3xl text-sm leading-6 text-muted-foreground sm:text-[15px] sm:leading-7">
									{SITE_NAME} được tối ưu cho{" "}
									<strong className="font-semibold text-foreground">chat ẩn danh</strong>,{" "}
									<strong className="font-semibold text-foreground">nói chuyện với người lạ</strong> và{" "}
									<strong className="font-semibold text-foreground">random chat Việt Nam</strong>. Route public
									/home giữ toàn bộ intent SEO, còn route / vẫn là khu chat thực tế.
								</p>
							</div>

							<div className="mt-4">
								<HomeCtaActions />
							</div>

							<div className="mt-4 flex flex-wrap gap-2">
								{keywordPills.map((pill) => (
									<span
										key={pill}
										className="rounded-full border border-border/80 bg-background/70 px-2.5 py-1 text-[11px] font-medium text-muted-foreground shadow-sm sm:px-3 sm:py-1.5 sm:text-xs"
									>
										{pill}
									</span>
								))}
							</div>

							<div className="mt-4 grid min-h-0 flex-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
								{highlights.map((item) => {
									const Icon = item.icon;

									return (
										<div
											key={item.title}
											className="rounded-[1.35rem] border border-border/70 bg-background/75 p-3 shadow-sm"
										>
											<div className="mb-2 flex size-9 items-center justify-center rounded-2xl bg-primary/10 text-primary">
												<Icon className="size-4" />
											</div>
											<h2 className="text-sm font-semibold">{item.title}</h2>
											<p className="mt-1.5 text-xs leading-5 text-muted-foreground sm:text-[13px]">
												{item.description}
											</p>
										</div>
									);
								})}
							</div>

							<div className="mt-4 rounded-[1.45rem] border border-border/70 bg-background/72 p-3 shadow-sm">
								<div className="mb-2 flex items-center justify-between text-[10px] uppercase tracking-[0.24em] text-muted-foreground sm:text-[11px]">
									<span>Product Intent</span>
									<span>1-1 chat</span>
								</div>
								<div className="grid gap-2.5 sm:grid-cols-[0.92fr_1.08fr]">
									<div className="space-y-2.5">
										<div className="max-w-[88%] rounded-3xl rounded-bl-md bg-primary px-3 py-2 text-xs leading-5 text-primary-foreground shadow-sm sm:text-[13px]">
											Hi, có ai muốn nói chuyện nhanh với người lạ không?
										</div>
										<div className="ml-auto max-w-[86%] rounded-3xl rounded-br-md border border-border/70 bg-card px-3 py-2 text-xs leading-5 text-foreground shadow-sm sm:text-[13px]">
											Có, nhưng mình không muốn phải lướt profile hay mở quá nhiều thông tin cá nhân.
										</div>
										<div className="max-w-[90%] rounded-3xl rounded-bl-md bg-secondary/45 px-3 py-2 text-xs leading-5 text-foreground shadow-sm sm:text-[13px]">
											AnoChat nhắm đúng điểm đó: anonymous chat gọn, rõ và đi thẳng vào hội thoại.
										</div>
									</div>
									<div className="grid gap-2 rounded-[1.2rem] border border-border/70 bg-card/70 p-3">
										<p className="text-xs leading-5 text-muted-foreground sm:text-[13px]">
											Trang này được nén lại để mọi block đều nằm gọn trong viewport desktop thay vì
											bắt bạn phải zoom out mới nhìn hết.
										</p>
										<p className="text-xs leading-5 text-muted-foreground sm:text-[13px]">
											Nội dung dài hơn cho SEO được giữ ở metadata, FAQ schema và copy cô đặc thay vì
											trải thành nhiều section cao.
										</p>
										<Button variant="outline" asChild className="mt-1 h-8 justify-between text-xs">
											<Link href="/login">
												Đăng nhập để bắt đầu
												<ArrowRight />
											</Link>
										</Button>
									</div>
								</div>
							</div>
						</div>
					</section>

					<div className="grid gap-3 lg:h-full lg:min-h-0 lg:grid-rows-[minmax(0,0.84fr)_minmax(0,1fr)]">
						<Card className="rounded-[1.75rem] border-border/70 bg-card/92 shadow-[0_18px_60px_rgba(15,23,42,0.06)] lg:min-h-0">
							<CardContent className="flex h-full min-h-0 flex-col p-4 sm:p-5">
								<div className="mb-3">
									<p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary/75 sm:text-xs">
										SEO Entry Page
									</p>
									<div style={{ fontFamily: "var(--font-changa-one)" }} className="mt-2">
										<AuroraText className="text-2xl tracking-[0.2em] sm:text-3xl">
											ANOCHAT
										</AuroraText>
									</div>
									<p className="mt-2 text-xs leading-5 text-muted-foreground sm:text-[13px]">
										/home là route public để Google index. /login, /callback, /error và / đều
										được tách vai trò rõ ràng để không tranh crawl với landing này.
									</p>
								</div>

								<div
									id="cach-hoat-dong"
									className="grid min-h-0 flex-1 gap-2.5 sm:grid-cols-3 lg:grid-cols-1 xl:grid-cols-3"
								>
									{steps.map((step) => (
										<div
											key={step.title}
											className="rounded-[1.25rem] border border-border/70 bg-background/72 p-3 shadow-sm"
										>
											<p className="text-[10px] font-semibold uppercase tracking-[0.24em] text-primary/75 sm:text-[11px]">
												{step.label}
											</p>
											<h2 className="mt-2 text-sm font-semibold leading-5">{step.title}</h2>
											<p className="mt-1.5 text-xs leading-5 text-muted-foreground sm:text-[13px]">
												{step.description}
											</p>
										</div>
									))}
								</div>
							</CardContent>
						</Card>

						<Card className="rounded-[1.75rem] border-border/70 bg-card/92 shadow-[0_18px_60px_rgba(15,23,42,0.06)] lg:min-h-0">
							<CardContent className="flex h-full min-h-0 flex-col p-4 sm:p-5">
								<div className="mb-3 flex items-start justify-between gap-3">
									<div>
										<p className="text-[11px] font-semibold uppercase tracking-[0.22em] text-primary/75 sm:text-xs">
											FAQ
										</p>
										<h2 className="mt-1 text-xl font-bold tracking-tight sm:text-2xl">
											Compact, nhưng vẫn đủ tín hiệu SEO.
										</h2>
									</div>
									<Button variant="outline" asChild className="hidden h-8 shrink-0 justify-between text-xs sm:inline-flex">
										<Link href="/">
											Vào chat
											<ArrowRight />
										</Link>
									</Button>
								</div>

								<div className="grid min-h-0 flex-1 gap-2.5 overflow-auto pr-1 sm:grid-cols-2">
									{faqs.map((faq) => (
										<div
											key={faq.question}
											className="rounded-[1.25rem] border border-border/70 bg-background/72 p-3 shadow-sm"
										>
											<h3 className="text-sm font-semibold leading-5">{faq.question}</h3>
											<p className="mt-1.5 text-xs leading-5 text-muted-foreground sm:text-[13px]">
												{faq.answer}
											</p>
										</div>
									))}
								</div>

								<p className="mt-3 text-xs leading-5 text-muted-foreground sm:text-[13px]">
									Bạn có thể{" "}
									<Link
										href="/login"
										className="font-medium text-foreground underline decoration-primary/50 underline-offset-4"
									>
										đăng nhập bằng Google
									</Link>{" "}
									để tạo phiên, hoặc vào{" "}
									<Link
										href="/"
										className="font-medium text-foreground underline decoration-primary/50 underline-offset-4"
									>
										khu chat chính
									</Link>{" "}
									nếu đã đăng nhập sẵn.
								</p>
							</CardContent>
						</Card>
					</div>
				</div>
			</div>
		</main>
	);
}
