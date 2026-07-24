"use client";

import { useCallback } from "react";
import { flushSync } from "react-dom";
import { useLanguage, useTheme, Theme } from "@/contexts/theme";
import { cn } from "@/lib/utils";
import type { TranslationKey } from "@/lib/i18n";

const themes: { id: Theme; labelKey: TranslationKey; gradient: string; ring: string }[] = [
	{
		id: "blue",
		labelKey: "blueTheme",
		gradient: "linear-gradient(135deg, #516b91 0%, #59c4e6 100%)",
		ring: "#59c4e6",
	},
	{
		id: "dark",
		labelKey: "darkTheme",
		gradient: "linear-gradient(135deg, #111111 0%, #52525b 100%)",
		ring: "#f4f4f5",
	},
	{
		id: "pink",
		labelKey: "pinkTheme",
		gradient: "linear-gradient(135deg, #c94f8a 0%, #f0a0cc 100%)",
		ring: "#f0a0cc",
	},
];

export function changeThemeWithTransition(
	nextTheme: Theme,
	currentTheme: Theme,
	setTheme: (theme: Theme) => void,
	origin: HTMLElement
) {
	if (nextTheme === currentTheme) return;

	if (typeof document.startViewTransition !== "function") {
		document.documentElement.classList.add("no-transition");
		setTheme(nextTheme);
		requestAnimationFrame(() =>
			requestAnimationFrame(() =>
				document.documentElement.classList.remove("no-transition")
			)
		);
		return;
	}

	const { top, left, width, height } = origin.getBoundingClientRect();
	const x = left + width / 2;
	const y = top + height / 2;
	const vw = window.visualViewport?.width ?? window.innerWidth;
	const vh = window.visualViewport?.height ?? window.innerHeight;
	const maxRadius = Math.hypot(Math.max(x, vw - x), Math.max(y, vh - y));

	const transition = document.startViewTransition(() => {
		flushSync(() => setTheme(nextTheme));
	});

	transition.ready.then(() => {
		document.documentElement.animate(
			{
				clipPath: [
					`circle(0px at ${x}px ${y}px)`,
					`circle(${maxRadius}px at ${x}px ${y}px)`,
				],
			},
			{
				duration: 500,
				easing: "ease-in-out",
				pseudoElement: "::view-transition-new(root)",
			}
		);
	});
}

export function ThemeToggle({ collapsed }: { collapsed?: boolean }) {
	const { theme, setTheme } = useTheme();
	const { t: translate } = useLanguage();

	const handleClick = useCallback(
		(t: Theme, e: React.MouseEvent<HTMLButtonElement>) =>
			changeThemeWithTransition(t, theme, setTheme, e.currentTarget),
		[theme, setTheme]
	);

	return (
		<div
			className={cn(
				"flex items-center gap-3 px-2 py-1.5 w-fit rounded-xl bg-black/5 dark:bg-white/5",
				collapsed ? "justify-center" : "justify-start"
			)}
		>
			{themes.map((t) => {
				const active = theme === t.id;
				const label = translate(t.labelKey);
				return (
					<button
						key={t.id}
						onClick={(e) => handleClick(t.id, e)}
						title={label}
						aria-label={label}
						className={cn(
							"relative rounded-md transition-all duration-150 ease-out focus-visible:outline-none cursor-pointer",
							active ? "scale-110 opacity-100 shadow-md" : "opacity-55 hover:opacity-85 hover:scale-102"
						)}
						style={{
							width: 22,
							height: 22,
							background: t.gradient,
							boxShadow: active
								? `0 0 0 2px var(--sidebar), 0 0 0 4px ${t.ring}, 0 3px 12px ${t.ring}80`
								: undefined,
						}}
						aria-pressed={active}
					/>
				);
			})}
			{/*{!collapsed && (
				<span className="ml-1 text-xs text-muted-foreground select-none capitalize">
					{theme}
				</span>
			)}*/}
		</div>
	);
}
