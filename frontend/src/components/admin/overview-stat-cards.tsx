"use client";

import type { ReactNode } from "react";
import { Users } from "lucide-react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useLanguage } from "@/contexts/theme";

interface GenderBreakdownProps {
	male: number;
	female: number;
	unknown: number;
	total: number;
}

export function StatCard({
	title,
	value,
	icon: Icon,
	iconClassName,
	children,
}: {
	title: string;
	value: number;
	icon: typeof Users;
	iconClassName: string;
	children: ReactNode;
}) {
	const { language } = useLanguage();
	return (
		<Card className="gap-2 py-3">
			<CardHeader className="grid grid-cols-[1fr_auto] items-start px-4">
				<div className="space-y-0.5">
					<CardDescription>{title}</CardDescription>
					<CardTitle className="text-2xl tabular-nums">
						{value.toLocaleString(language === "vi" ? "vi-VN" : "en-US")}
					</CardTitle>
				</div>
				<div className={`rounded-lg p-2 ${iconClassName}`}>
					<Icon className="h-4 w-4" />
				</div>
			</CardHeader>
			<CardContent className="px-4">{children}</CardContent>
		</Card>
	);
}

export function GenderBreakdown({
	male,
	female,
	unknown,
	total,
}: GenderBreakdownProps) {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const items = [
		{ label: t("male"), value: male, color: "bg-cyan-500" },
		{ label: t("female"), value: female, color: "bg-pink-500" },
		{
			label: t("adminUnspecified"),
			value: unknown,
			color: "bg-muted-foreground/40",
		},
	];

	return (
		<div className="grid grid-cols-3 gap-2 border-t pt-2">
			{items.map((item) => (
				<div key={item.label} className="min-w-0">
					<div className="flex items-center gap-1 text-[10px] text-muted-foreground">
						<span className={`h-2 w-2 shrink-0 rounded-full ${item.color}`} />
						<span className="truncate">{item.label}</span>
					</div>
					<p className="mt-0.5 text-xs font-semibold tabular-nums">
						{item.value.toLocaleString(locale)}
						<span className="ml-1 text-[10px] font-normal text-muted-foreground">
							{percentage(item.value, total)}%
						</span>
					</p>
				</div>
			))}
		</div>
	);
}

function percentage(value: number, total: number) {
	if (total === 0) return 0;
	return Math.round((value / total) * 100);
}
