"use client";

import { useQuery } from "@tanstack/react-query";
import { Clock3, MessageCircleMore, Users } from "lucide-react";
import { useState } from "react";
import type { PointerEvent as ReactPointerEvent, ReactNode } from "react";
import { adminAPI } from "@/lib/api";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import type {
	AdminOverviewDTO,
	DailyOverviewMetricDTO,
} from "@/types";
import { useLanguage } from "@/contexts/theme";
import type { Language } from "@/lib/i18n";

const REFRESH_INTERVAL_MS = 2 * 60 * 1000;

interface GenderBreakdownProps {
	male: number;
	female: number;
	unknown: number;
	total: number;
}

export function OverviewTab() {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const {
		data: overview,
		isLoading,
		isError,
		dataUpdatedAt,
	} = useQuery({
		queryKey: ["admin", "overview"],
		queryFn: async () => {
			const response = await adminAPI.getOverview();
			return response.data as AdminOverviewDTO;
		},
		refetchInterval: REFRESH_INTERVAL_MS,
	});

	if (isLoading) {
		return (
			<div className="grid grid-cols-1 gap-3 lg:grid-cols-3">
				{Array.from({ length: 3 }).map((_, index) => (
					<div
						key={index}
						className="h-32 animate-pulse rounded-xl border bg-muted/40"
					/>
				))}
			</div>
		);
	}

	if (isError || !overview) {
		return (
			<div className="rounded-lg border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
				{t("adminUnableLoadOverview")}
			</div>
		);
	}

	return (
		<div className="space-y-3">
			<div className="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<StatCard
					title={t("adminTotalUsers")}
					value={overview.total_users}
					icon={Users}
					iconClassName="bg-blue-500/10 text-blue-600 dark:text-blue-400"
				>
					<GenderBreakdown
						male={overview.male_users}
						female={overview.female_users}
						unknown={overview.unspecified_users}
						total={overview.total_users}
					/>
				</StatCard>

				<StatCard
					title={t("adminInQueue")}
					value={overview.in_queue}
					icon={Clock3}
					iconClassName="bg-amber-500/10 text-amber-600 dark:text-amber-400"
				>
					<GenderBreakdown
						male={overview.in_queue_male}
						female={overview.in_queue_female}
						unknown={overview.in_queue_unknown}
						total={overview.in_queue}
					/>
				</StatCard>

				<StatCard
					title={t("adminActiveRooms")}
					value={overview.active_rooms}
					icon={MessageCircleMore}
					iconClassName="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
				>
					<div className="flex items-end justify-between gap-3 border-t pt-2">
						<p className="text-xs text-muted-foreground">
							{t("adminUsersInRooms", {
								count: (overview.active_rooms * 2).toLocaleString(locale),
							})}
						</p>
						<p className="shrink-0 text-right text-[10px] leading-4 text-muted-foreground">
							{t("adminAutoRefresh")}
							{dataUpdatedAt > 0 && (
								<>
									<br />
									{t("adminUpdatedAt", {
										time: new Date(dataUpdatedAt).toLocaleTimeString(locale, {
											hour: "2-digit",
											minute: "2-digit",
											second: "2-digit",
										}),
									})}
								</>
							)}
						</p>
					</div>
				</StatCard>
			</div>

			<SevenDayTrends data={overview.daily_metrics} />
		</div>
	);
}

function SevenDayTrends({ data }: { data: DailyOverviewMetricDTO[] }) {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	if (data.length === 0) return null;

	const matches = data.map((point) => point.matches);
	const users = data.map((point) => point.total_users);
	const activeRooms = data.map((point) => point.active_rooms);
	const totalMatches = matches.reduce((sum, value) => sum + value, 0);
	const userGrowth = users.at(-1)! - users[0];
	const roomChange = activeRooms.at(-1)! - activeRooms.at(-2)!;

	return (
		<div className="grid grid-cols-1 gap-3 md:grid-cols-3">
			<TrendCard
				title={t("adminMatches")}
				value={totalMatches}
				detail={t("adminSevenDayTotal")}
				color="#3b82f6"
				startDate={data[0].date}
				endDate={data.at(-1)!.date}
			>
				<BarSparkline
					data={data}
					values={matches}
					color="#3b82f6"
					label={t("adminMatches")}
				/>
			</TrendCard>

			<TrendCard
				title={t("adminTotalUsers")}
				value={users.at(-1)!}
				detail={t("adminValueChangeSevenDays", {
					value: formatSigned(userGrowth, locale),
				})}
				color="#8b5cf6"
				startDate={data[0].date}
				endDate={data.at(-1)!.date}
			>
				<LineSparkline
					data={data}
					values={users}
					color="#8b5cf6"
					label={t("adminTotalUsers")}
				/>
			</TrendCard>

			<TrendCard
				title={t("adminActiveRooms")}
				value={activeRooms.at(-1)!}
				detail={t("adminValueVersusYesterday", {
					value: formatSigned(roomChange, locale),
				})}
				color="#10b981"
				startDate={data[0].date}
				endDate={data.at(-1)!.date}
			>
				<LineSparkline
					data={data}
					values={activeRooms}
					color="#10b981"
					label={t("adminActiveRooms")}
					fill
				/>
			</TrendCard>
		</div>
	);
}

function TrendCard({
	title,
	value,
	detail,
	color,
	startDate,
	endDate,
	children,
}: {
	title: string;
	value: number;
	detail: string;
	color: string;
	startDate: string;
	endDate: string;
	children: ReactNode;
}) {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	return (
		<Card className="gap-1 py-3">
			<CardHeader className="grid grid-cols-[1fr_auto] items-start px-4">
				<div className="space-y-0.5">
					<CardDescription>
						{title} · {t("adminSevenDays")}
					</CardDescription>
					<CardTitle className="text-xl tabular-nums">
						{value.toLocaleString(locale)}
					</CardTitle>
				</div>
				<span
					className="mt-1 h-2.5 w-2.5 rounded-full"
					style={{ backgroundColor: color }}
				/>
			</CardHeader>
			<CardContent className="px-4">
				<p className="mb-1 text-[10px] text-muted-foreground">{detail}</p>
				{children}
				<div className="mt-0.5 flex justify-between text-[9px] text-muted-foreground">
					<span>{formatChartDate(startDate, language)}</span>
					<span>{formatChartDate(endDate, language)}</span>
				</div>
			</CardContent>
		</Card>
	);
}

function BarSparkline({
	data,
	values,
	color,
	label,
}: {
	data: DailyOverviewMetricDTO[];
	values: number[];
	color: string;
	label: string;
}) {
	const { language, t } = useLanguage();
	const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
	const width = 240;
	const height = 48;
	const gap = 7;
	const barWidth = (width - gap * (values.length - 1)) / values.length;
	const maxValue = Math.max(...values, 1);

	return (
		<div className="relative">
			{hoveredIndex !== null && (
				<ChartTooltip
					index={hoveredIndex}
					count={data.length}
					date={data[hoveredIndex].date}
					label={label}
					value={values[hoveredIndex]}
					color={color}
				/>
			)}
			<svg
				viewBox={`0 0 ${width} ${height}`}
				className="h-12 w-full cursor-crosshair touch-none"
				role="img"
				aria-label={t("adminDailyMatchesAria")}
				onPointerMove={(event) =>
					setHoveredIndex(getPointerIndex(event, values.length))
				}
				onPointerLeave={() => setHoveredIndex(null)}
			>
				{hoveredIndex !== null && (
					<line
						x1={hoveredIndex * (barWidth + gap) + barWidth / 2}
						x2={hoveredIndex * (barWidth + gap) + barWidth / 2}
						y1="0"
						y2={height}
						className="stroke-foreground/25"
						strokeDasharray="2 2"
					/>
				)}
				{values.map((value, index) => {
					const barHeight = Math.max(2, (value / maxValue) * (height - 4));
					const isHovered = hoveredIndex === index;
					return (
						<rect
							key={data[index].date}
							x={index * (barWidth + gap)}
							y={height - barHeight}
							width={barWidth}
							height={barHeight}
							r="3"
							fill={color}
							fillOpacity={
								hoveredIndex === null
									? index === values.length - 1
										? 1
										: 0.55
									: isHovered
										? 1
										: 0.25
							}
						>
							<title>
								{formatPointTitle(data[index].date, value, label, language)}
							</title>
						</rect>
					);
				})}
			</svg>
		</div>
	);
}

function LineSparkline({
	data,
	values,
	color,
	label,
	fill = false,
}: {
	data: DailyOverviewMetricDTO[];
	values: number[];
	color: string;
	label: string;
	fill?: boolean;
}) {
	const { language, t } = useLanguage();
	const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
	const width = 240;
	const height = 48;
	const padding = 4;
	const minValue = Math.min(...values);
	const maxValue = Math.max(...values);
	const range = Math.max(1, maxValue - minValue);
	const isFlat = maxValue === minValue;
	const x = (index: number) =>
		padding + (index * (width - padding * 2)) / Math.max(1, values.length - 1);
	const y = (value: number) =>
		isFlat
			? height / 2
			: padding + ((maxValue - value) / range) * (height - padding * 2);
	const points = values.map((value, index) => `${x(index)},${y(value)}`).join(" ");
	const areaPoints = `${padding},${height} ${points} ${width - padding},${height}`;

	return (
		<div className="relative">
			{hoveredIndex !== null && (
				<ChartTooltip
					index={hoveredIndex}
					count={data.length}
					date={data[hoveredIndex].date}
					label={label}
					value={values[hoveredIndex]}
					color={color}
				/>
			)}
			<svg
				viewBox={`0 0 ${width} ${height}`}
				className="h-12 w-full cursor-crosshair touch-none"
				role="img"
				aria-label={t("adminTrendAria", { label: label.toLowerCase() })}
				onPointerMove={(event) =>
					setHoveredIndex(getPointerIndex(event, values.length))
				}
				onPointerLeave={() => setHoveredIndex(null)}
			>
				{fill && (
					<polygon points={areaPoints} fill={color} fillOpacity="0.12" />
				)}
				{hoveredIndex !== null && (
					<line
						x1={x(hoveredIndex)}
						x2={x(hoveredIndex)}
						y1="0"
						y2={height}
						className="stroke-foreground/25"
						strokeDasharray="2 2"
					/>
				)}
				<polyline
					points={points}
					fill="none"
					stroke={color}
					strokeWidth="2.5"
					strokeLinejoin="round"
					strokeLinecap="round"
				/>
				{values.map((value, index) => {
					const isHovered = hoveredIndex === index;
					return (
						<circle
							key={data[index].date}
							cx={x(index)}
							cy={y(value)}
							r={isHovered ? 4 : index === values.length - 1 ? 3 : 2}
							fill={color}
							fillOpacity={
								hoveredIndex === null || isHovered ? 1 : 0.35
							}
							className="stroke-card"
							strokeWidth={isHovered ? 2 : 1.5}
						>
							<title>
								{formatPointTitle(data[index].date, value, label, language)}
							</title>
						</circle>
					);
				})}
			</svg>
		</div>
	);
}

function ChartTooltip({
	index,
	count,
	date,
	label,
	value,
	color,
}: {
	index: number;
	count: number;
	date: string;
	label: string;
	value: number;
	color: string;
}) {
	const { language } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const alignment =
		index === 0
			? "translate-x-0"
			: index === count - 1
				? "-translate-x-full"
				: "-translate-x-1/2";

	return (
		<div
			className={`pointer-events-none absolute top-0 z-10 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-[10px] text-popover-foreground shadow-md ${alignment}`}
			style={{
				left: `${(index / Math.max(1, count - 1)) * 100}%`,
				borderColor: color,
			}}
		>
			<span className="text-muted-foreground">
				{formatChartDate(date, language)}
			</span>
			<span className="mx-1">·</span>
			<span>{label}</span>
			<span className="ml-1 font-semibold tabular-nums">
				{value.toLocaleString(locale)}
			</span>
		</div>
	);
}

function getPointerIndex(
	event: ReactPointerEvent<SVGSVGElement>,
	count: number,
) {
	const bounds = event.currentTarget.getBoundingClientRect();
	const ratio = Math.min(
		1,
		Math.max(0, (event.clientX - bounds.left) / bounds.width),
	);
	return Math.round(ratio * Math.max(0, count - 1));
}

function formatSigned(value: number, locale: string) {
	if (value > 0) return `+${value.toLocaleString(locale)}`;
	return value.toLocaleString(locale);
}

function formatChartDate(date: string, language: Language) {
	return new Date(`${date}T00:00:00`).toLocaleDateString(
		language === "vi" ? "vi-VN" : "en-US",
		{
		day: "2-digit",
		month: "2-digit",
		}
	);
}

function formatPointTitle(
	date: string,
	value: number,
	label: string,
	language: Language
) {
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const formattedDate = new Date(`${date}T00:00:00`).toLocaleDateString(
		locale,
		{
			day: "2-digit",
			month: "2-digit",
			year: "numeric",
		},
	);
	return `${label} · ${formattedDate}: ${value.toLocaleString(locale)}`;
}

function StatCard({
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

function GenderBreakdown({ male, female, unknown, total }: GenderBreakdownProps) {
	const { language, t } = useLanguage();
	const locale = language === "vi" ? "vi-VN" : "en-US";
	const items = [
		{ label: t("male"), value: male, color: "bg-cyan-500" },
		{ label: t("female"), value: female, color: "bg-pink-500" },
		{ label: t("adminUnspecified"), value: unknown, color: "bg-muted-foreground/40" },
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
