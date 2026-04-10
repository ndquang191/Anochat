"use client";

import { useCallback } from "react";
import { flushSync } from "react-dom";
import { useLanguage } from "@/contexts/theme";
import { cn } from "@/lib/utils";
import type { Language } from "@/lib/i18n";

const languages: {
	id: Language;
	label: string;
	ring: string;
}[] = [
	{
		id: "vi",
		label: "Vietnamese",
		ring: "#d61f26",
	},
	{
		id: "en",
		label: "English",
		ring: "#3c3b6e",
	},
];

export function LanguageToggle({ collapsed }: { collapsed?: boolean }) {
	const { language, setLanguage } = useLanguage();

	const handleClick = useCallback(
		(nextLanguage: Language, e: React.MouseEvent<HTMLButtonElement>) => {
			if (nextLanguage === language) return;

			if (typeof document.startViewTransition !== "function") {
				document.documentElement.classList.add("no-transition");
				setLanguage(nextLanguage);
				requestAnimationFrame(() =>
					requestAnimationFrame(() =>
						document.documentElement.classList.remove("no-transition")
					)
				);
				return;
			}

			const { top, left, width, height } = e.currentTarget.getBoundingClientRect();
			const x = left + width / 2;
			const y = top + height / 2;
			const vw = window.visualViewport?.width ?? window.innerWidth;
			const vh = window.visualViewport?.height ?? window.innerHeight;
			const maxRadius = Math.hypot(Math.max(x, vw - x), Math.max(y, vh - y));

			const transition = document.startViewTransition(() => {
				flushSync(() => setLanguage(nextLanguage));
			});

			transition.ready.then(() => {
				document.documentElement.animate(
					{
						clipPath: [
							`circle(0px at ${x}px ${y}px)`,
							`circle(${maxRadius * 0.72}px at ${x}px ${y}px)`,
							`circle(${maxRadius}px at ${x}px ${y}px)`,
						],
						filter: [
							"blur(18px) saturate(1.08) brightness(1.1)",
							"blur(10px) saturate(1.06) brightness(1.06)",
							"blur(0px) saturate(1) brightness(1)",
						],
						boxShadow: [
							`0 0 0 0 rgba(255,255,255,0), 0 0 0 0 rgba(255,255,255,0)`,
							`0 0 0 14px rgba(255,255,255,0.18), 0 0 44px 20px rgba(255,255,255,0.22)`,
							`0 0 0 0 rgba(255,255,255,0), 0 0 0 0 rgba(255,255,255,0)`,
						],
						opacity: [0.7, 1, 1],
					},
					{
						duration: 620,
						easing: "ease-in-out",
						pseudoElement: "::view-transition-new(root)",
					}
				);

				document.documentElement.animate(
					{
						clipPath: [
							`circle(0px at ${x}px ${y}px)`,
							`circle(${maxRadius * 0.78}px at ${x}px ${y}px)`,
							`circle(${maxRadius}px at ${x}px ${y}px)`,
						],
						filter: [
							"blur(0px) brightness(1)",
							"blur(10px) brightness(1.08)",
							"blur(18px) brightness(1.12)",
						],
						boxShadow: [
							`0 0 0 0 rgba(255,255,255,0), 0 0 0 0 rgba(255,255,255,0)`,
							`0 0 0 10px rgba(255,255,255,0.12), 0 0 30px 12px rgba(255,255,255,0.16)`,
							`0 0 0 0 rgba(255,255,255,0), 0 0 0 0 rgba(255,255,255,0)`,
						],
						opacity: [0, 0.22, 0],
					},
					{
						duration: 620,
						easing: "ease-out",
						pseudoElement: "::view-transition-old(root)",
					}
				);
			});
		},
		[language, setLanguage]
	);

	return (
		<div
			className={cn(
				"flex w-fit items-center gap-3 rounded-xl bg-black/5 px-2 py-1.5 dark:bg-white/5",
				collapsed ? "justify-center" : "justify-start"
			)}
		>
			{languages.map((item) => {
				const active = language === item.id;
				return (
					<button
						key={item.id}
						type="button"
						onClick={(e) => handleClick(item.id, e)}
						title={item.label}
						aria-label={item.label}
						aria-pressed={active}
						className={cn(
							"relative cursor-pointer rounded-md transition-all duration-150 ease-out focus-visible:outline-none",
							active ? "scale-105 shadow-md" : "opacity-55 hover:scale-102 hover:opacity-85"
						)}
						style={{
							width: 22,
							height: 22,
							boxShadow: active
								? `0 0 0 2px var(--sidebar), 0 0 0 3.5px ${item.ring}, 0 3px 10px ${item.ring}55`
								: undefined,
						}}
					>
						{item.id === "vi" ? (
							<span className="absolute inset-0 flex items-center justify-center rounded-md bg-[#d61f26] text-[10px] leading-none text-[#f6d04d]">
								★
							</span>
						) : (
							<span className="absolute inset-0 overflow-hidden rounded-md bg-[#1f3c88]">
								<span className="absolute inset-y-0 left-[38%] w-[24%] bg-white" />
								<span className="absolute inset-x-0 top-[38%] h-[24%] bg-white" />
								<span className="absolute inset-y-0 left-[42%] w-[16%] bg-[#d61f26]" />
								<span className="absolute inset-x-0 top-[42%] h-[16%] bg-[#d61f26]" />
							</span>
						)}
					</button>
				);
			})}
		</div>
	);
}
