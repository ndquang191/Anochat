"use client";

import { useCallback } from "react";
import { flushSync } from "react-dom";
import { useTheme, Theme } from "@/contexts/theme";
import { cn } from "@/lib/utils";

const themes: { id: Theme; label: string; gradient: string; ring: string }[] = [
	{
		id: "blue",
		label: "Blue",
		gradient: "linear-gradient(135deg, #516b91 0%, #59c4e6 100%)",
		ring: "#59c4e6",
	},
	{
		id: "dark",
		label: "Dark",
		gradient: "linear-gradient(135deg, #1a1a1a 0%, #3a3a3a 100%)",
		ring: "#888888",
	},
	{
		id: "pink",
		label: "Pink",
		gradient: "linear-gradient(135deg, #c94f8a 0%, #f0a0cc 100%)",
		ring: "#f0a0cc",
	},
];

export function ThemeToggle({ collapsed }: { collapsed?: boolean }) {
	const { theme, setTheme } = useTheme();

	const handleClick = useCallback(
		(t: Theme, e: React.MouseEvent<HTMLButtonElement>) => {
			if (t === theme) return;

			if (typeof document.startViewTransition !== "function") {
				// Fallback: block transitions briefly to prevent flicker
				document.documentElement.classList.add("no-transition");
				setTheme(t);
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
				flushSync(() => setTheme(t));
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
		},
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
				return (
					<button
						key={t.id}
						onClick={(e) => handleClick(t.id, e)}
						title={t.label}
						className={cn(
							"relative rounded-md transition-all duration-150 ease-out focus-visible:outline-none cursor-pointer",
							active ? "scale-105 shadow-md" : "opacity-55 hover:opacity-85 hover:scale-102"
						)}
						style={{
							width: 22,
							height: 22,
							background: t.gradient,
							boxShadow: active
								? `0 0 0 2px var(--sidebar), 0 0 0 3.5px ${t.ring}, 0 3px 10px ${t.ring}55`
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
