import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight, MessageCircle, Shield, Sparkles } from "lucide-react";
import { BrandLogo } from "@/components/brand-logo";
import { HomeCtaActions } from "@/components/home/home-cta-actions";
import { ShineBorder } from "@/components/ui/shine-border";
import { Button } from "@/components/ui/button";
import {
	SITE_NAME,
	SITE_TAGLINE,
	absoluteUrl,
} from "@/lib/site";

const pageTitle = "Chat ẩn danh với người lạ tại Việt Nam | AnoChat";
const pageDescription =
	"Chat ẩn danh 1-1 với người lạ tại Việt Nam trên AnoChat. Ghép đôi ngẫu nhiên, không profile công khai và có thể rời cuộc trò chuyện bất cứ lúc nào.";

const steps = [
	{
		label: "01",
		title: "Đăng nhập",
		description: "Một bước xác thực nhanh để giữ trải nghiệm an toàn.",
	},
	{
		label: "02",
		title: "Gặp người mới",
		description: "Hệ thống ghép ngẫu nhiên với ai đó đang chờ.",
	},
	{
		label: "03",
		title: "Lắng nghe và sẻ chia",
		description: "Ở lại nếu thấy thoải mái, hoặc rời đi bất cứ lúc nào.",
	},
];

const faqs = [
	{
		question: "Chat trên AnoChat có ẩn danh không?",
		answer: "Có. Người được ghép đôi không thấy tên, email hay hồ sơ Google của bạn. Bạn chỉ chia sẻ những gì mình chủ động nói trong cuộc trò chuyện.",
	},
	{
		question: "Tại sao AnoChat yêu cầu đăng nhập?",
		answer: "Bước xác thực giúp hạn chế spam và hỗ trợ các công cụ an toàn. Danh tính đăng nhập không được hiển thị cho người trò chuyện cùng bạn.",
	},
	{
		question: "AnoChat ghép đôi như thế nào?",
		answer: "Khi bạn bắt đầu tìm, AnoChat sẽ ghép ngẫu nhiên hai người đang chờ để trò chuyện riêng tư 1-1.",
	},
	{
		question: "Tôi có thể kết thúc cuộc trò chuyện không?",
		answer: "Có. Bạn có thể rời đi bất cứ lúc nào nếu không còn thoải mái, sau đó tìm một cuộc trò chuyện mới.",
	},
];

const keywordPills = ["Ẩn danh mặc định", "Ghép đôi ngẫu nhiên", "Rời đi bất cứ lúc nào"];

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
		title: pageTitle,
		description: pageDescription,
		url: absoluteUrl("/home"),
		siteName: SITE_NAME,
		type: "website",
		locale: "vi_VN",
		images: [
			{
				url: absoluteUrl("/opengraph-image"),
				width: 1200,
				height: 630,
				alt: "AnoChat — Nói điều thật lòng",
			},
		],
	},
	twitter: {
		card: "summary_large_image",
		title: pageTitle,
		description: pageDescription,
		images: [absoluteUrl("/opengraph-image")],
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
			description: pageDescription,
		},
		{
			"@context": "https://schema.org",
			"@type": "WebPage",
			name: pageTitle,
			url: absoluteUrl("/home"),
			description: pageDescription,
			inLanguage: "vi-VN",
			isPartOf: {
				"@type": "WebSite",
				name: SITE_NAME,
				url: absoluteUrl("/"),
			},
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
		<main className="relative min-h-svh overflow-x-hidden bg-background text-foreground">
			<script
				type="application/ld+json"
				dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
			/>

			<div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_8%_8%,_rgba(81,107,145,0.18),_transparent_32%),radial-gradient(circle_at_92%_18%,_rgba(89,196,230,0.12),_transparent_24%),radial-gradient(circle_at_52%_100%,_rgba(203,176,227,0.13),_transparent_28%)]" />

			<div className="relative mx-auto w-full max-w-[1440px] p-3 sm:p-4">
				<section className="relative min-h-[calc(100svh-1.5rem)] overflow-hidden rounded-[1.75rem] border border-border/70 bg-card/88 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur">
					<ShineBorder
						borderWidth={1}
						duration={12}
						shineColor={["rgba(81,107,145,0.12)", "rgba(147,183,227,0.42)", "rgba(89,196,230,0.16)"]}
					/>

					<div className="grid min-h-[calc(100svh-1.5rem)] grid-rows-[auto_auto_auto] gap-5 p-4 sm:p-5 lg:grid-rows-[auto_minmax(0,1fr)_auto] lg:gap-4 lg:p-6">
						<header className="flex items-center justify-between gap-4">
							<BrandLogo
								iconClassName="size-9"
								sloganClassName="text-[13px]"
							/>
							<div className="flex items-center gap-2">
								<div className="hidden items-center gap-2 rounded-full border border-border/70 bg-background/65 px-3 py-1.5 text-xs text-muted-foreground sm:flex">
									<span className="size-1.5 rounded-full bg-emerald-500" />
									{SITE_TAGLINE}
								</div>
								<Button variant="ghost" size="sm" asChild className="h-9 rounded-full px-4">
									<Link href="/">
										Vào chat
										<ArrowRight className="size-4" />
									</Link>
								</Button>
							</div>
						</header>

						<div className="grid min-h-0 items-center gap-6 lg:grid-cols-[1.06fr_0.94fr] lg:gap-8 xl:gap-12">
							<div className="flex min-w-0 flex-col justify-center">
								<div className="inline-flex w-fit items-center gap-2 rounded-full border border-primary/15 bg-primary/8 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.2em] text-primary sm:text-xs">
									<Sparkles className="size-3.5" />
									Anonymous Chat VN
								</div>

								<h1 className="mt-4 max-w-3xl text-[2rem] font-black leading-[1.02] tracking-[-0.04em] sm:text-[2.65rem] lg:text-[3rem] xl:text-[3.5rem]">
									Chat ẩn danh với người lạ. Nói điều thật lòng.
								</h1>

								<p className="mt-4 max-w-2xl text-sm leading-6 text-muted-foreground sm:text-base sm:leading-7">
									AnoChat ghép đôi ngẫu nhiên hai người tại Việt Nam trong một cuộc
									trò chuyện riêng tư—không feed, không profile công khai, không áp lực.
								</p>

								<div className="mt-5">
									<HomeCtaActions />
								</div>

								<div className="mt-4 flex flex-wrap gap-2">
									{keywordPills.map((pill) => (
										<span
											key={pill}
											className="rounded-full border border-border/70 bg-background/60 px-3 py-1.5 text-[11px] font-medium text-muted-foreground sm:text-xs"
										>
											{pill}
										</span>
									))}
								</div>
							</div>

							<div className="relative mx-auto flex w-full max-w-xl items-center justify-center lg:h-full lg:min-h-0">
								<div className="pointer-events-none absolute left-8 top-4 size-36 rounded-full bg-primary/10 blur-3xl" />
								<div className="pointer-events-none absolute bottom-4 right-4 size-40 rounded-full bg-cyan-400/10 blur-3xl" />

								<div className="relative w-full overflow-hidden rounded-[1.6rem] border border-border/70 bg-background/82 p-3 shadow-[0_20px_60px_rgba(15,23,42,0.10)] sm:p-4 lg:max-h-[420px]">
									<div className="flex items-center justify-between border-b border-border/60 pb-3">
										<div className="flex items-center gap-2.5">
											<div className="flex size-9 items-center justify-center rounded-full bg-primary/10 text-primary">
												<MessageCircle className="size-4" />
											</div>
											<div>
												<p className="text-sm font-semibold">Một người đang ở đây</p>
												<p className="text-[11px] text-muted-foreground">
													Riêng tư · vừa kết nối
												</p>
											</div>
										</div>
										<span className="flex items-center gap-1.5 text-[11px] text-emerald-600 dark:text-emerald-400">
											<span className="size-1.5 rounded-full bg-emerald-500" />
											Trực tuyến
										</span>
									</div>

									<div className="space-y-2.5 py-4">
										<div className="max-w-[78%] rounded-2xl rounded-bl-md bg-muted px-3 py-2 text-sm leading-5">
											Hôm nay của bạn thế nào?
										</div>
										<div className="ml-auto max-w-[82%] rounded-2xl rounded-br-md bg-primary px-3 py-2 text-sm leading-5 text-primary-foreground">
											Hơi dài một chút. Nhưng có người hỏi vậy là mình thấy nhẹ
											hơn rồi.
										</div>
										<div className="max-w-[76%] rounded-2xl rounded-bl-md bg-muted px-3 py-2 text-sm leading-5">
											Vậy thì cứ kể từ từ nhé. Mình đang nghe.
										</div>
									</div>

									<div className="flex items-center justify-between rounded-xl border border-border/60 bg-card/70 px-3 py-2.5 text-[11px] text-muted-foreground">
										<span className="flex items-center gap-1.5">
											<Shield className="size-3.5 text-primary" />
											Không lưu hồ sơ công khai
										</span>
										<span>Ghép đôi 1–1</span>
									</div>
								</div>
							</div>
						</div>

						<div
							id="cach-hoat-dong"
							className="grid gap-2.5 border-t border-border/60 pt-4 sm:grid-cols-3"
						>
							{steps.map((step) => (
								<div
									key={step.label}
									className="grid grid-cols-[auto_1fr] gap-3 rounded-2xl bg-background/55 px-3 py-3"
								>
									<span className="flex size-8 items-center justify-center rounded-full border border-primary/15 bg-primary/8 text-[10px] font-bold tracking-wider text-primary">
										{step.label}
									</span>
									<div className="min-w-0">
										<h2 className="text-sm font-semibold">{step.title}</h2>
										<p className="mt-0.5 line-clamp-2 text-[11px] leading-4 text-muted-foreground sm:text-xs">
											{step.description}
										</p>
									</div>
								</div>
							))}
						</div>
					</div>
				</section>

				<section className="mx-auto max-w-5xl px-2 py-14 sm:px-6 sm:py-20" aria-labelledby="gioi-thieu-anochat">
					<div className="grid gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:gap-16">
						<div>
							<p className="text-xs font-bold uppercase tracking-[0.2em] text-primary">Chat ẩn danh Việt Nam</p>
							<h2 id="gioi-thieu-anochat" className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
								Một nơi đơn giản để bắt đầu trò chuyện
							</h2>
						</div>
						<div className="space-y-4 text-sm leading-7 text-muted-foreground sm:text-base">
							<p>
								AnoChat dành cho những lúc bạn muốn nói chuyện với một người lạ mà không cần xây dựng profile, theo dõi feed hay tạo ấn tượng với cả một cộng đồng. Bạn chỉ cần đăng nhập, bắt đầu tìm và chờ hệ thống ghép với một người khác đang trực tuyến.
							</p>
							<p>
								Cuộc trò chuyện diễn ra 1-1. Người đối diện không thấy tên, email hoặc hồ sơ Google của bạn. Bạn quyết định điều gì muốn chia sẻ và có thể rời đi bất cứ lúc nào nếu cuộc nói chuyện không còn phù hợp.
							</p>
							<p>
								Thay vì cố trở thành một mạng xã hội khác, AnoChat tập trung vào một việc: giúp hai người xa lạ gặp nhau nhanh chóng và có một cuộc trò chuyện chân thành, ít áp lực hơn.
							</p>
						</div>
					</div>
				</section>

				<section className="mx-auto max-w-5xl rounded-[1.75rem] border border-border/70 bg-card/70 px-5 py-8 sm:px-8 sm:py-10" aria-labelledby="cau-hoi-thuong-gap">
					<p className="text-xs font-bold uppercase tracking-[0.2em] text-primary">Tìm hiểu thêm</p>
					<h2 id="cau-hoi-thuong-gap" className="mt-3 text-2xl font-black tracking-tight sm:text-3xl">
						Câu hỏi thường gặp về AnoChat
					</h2>
					<div className="mt-7 grid gap-x-10 gap-y-7 md:grid-cols-2">
						{faqs.map((faq) => (
							<article key={faq.question}>
								<h3 className="font-bold">{faq.question}</h3>
								<p className="mt-2 text-sm leading-6 text-muted-foreground">{faq.answer}</p>
							</article>
						))}
					</div>
					<div className="mt-9 border-t border-border/60 pt-7">
						<Button asChild>
							<Link href="/">
								Bắt đầu chat ẩn danh
								<ArrowRight className="size-4" />
							</Link>
						</Button>
					</div>
				</section>

				<footer className="px-4 py-8 text-center text-xs text-muted-foreground">
					AnoChat — chat ẩn danh 1-1 với người lạ tại Việt Nam.
				</footer>
			</div>
		</main>
	);
}
