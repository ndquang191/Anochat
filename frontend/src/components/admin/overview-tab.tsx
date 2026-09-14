"use client";

import { useQuery } from "@tanstack/react-query";
import { Clock3, MessageCircleMore, Users } from "lucide-react";
import { adminAPI } from "@/lib/api";
import type { AdminOverviewDTO } from "@/types";
import { useLanguage } from "@/contexts/theme";
import {
	GenderBreakdown,
	StatCard,
} from "@/components/admin/overview-stat-cards";
import { SevenDayTrends } from "@/components/admin/overview-trends";

const REFRESH_INTERVAL_MS = 2 * 60 * 1000;

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
